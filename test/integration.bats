#!/usr/bin/env bats

RAWKAFKA_CMD="${RAWKAFKA_CMD:-./build/rawkafka}"
RAWKAFKA_PID=
export RAWKAFKA_PORT="${RAWKAFKA_PORT:-7000}"

load common

function setup {
  [ -n "$RAWKAFKA_SCHEMA_REGISTRY_URL" ] || fatal 'Empty $RAWKAFKA_SCHEMA_REGISTRY_URL'
  [ -n "$RAWKAFKA_REST_ENDPOINT" ] || fatal 'Empty $RAWKAFKA_REST_ENDPOINT'
  [ -n "$RAWKAFKA_CMD" ] || fatal 'Empty $RAWKAFKA_CMD'

  RAWKAFKA_PID=""

  wait-for-status-code "$RAWKAFKA_SCHEMA_REGISTRY_URL/subjects/NoSchema-key/versions" 404
  wait-for-status-code "$RAWKAFKA_REST_ENDPOINT/topics/NotExistingTopic" 404
  info 'set up'
}

function teardown {
  [ -z "$RAWKAFKA_PID" ] || kill "$RAWKAFKA_PID" || true
}

@test 'can receive requests' {
  local url="http://localhost:${RAWKAFKA_PORT}"
  local log_file="$BATS_TMPDIR/request.log"
  info "listening at $url"

  touch "$log_file"
  $RAWKAFKA_CMD 3>&- 2> "$log_file" >> "$log_file" &
  RAWKAFKA_PID="$!"
  sleep 5

  info "running rawkfaka with PID = $RAWKAFKA_PID"

  info "waiting for ping..."
  if ! MAX_WAIT_TIME=10 wait-for-ping "$url/ping"; then
    cat "$log_file" | tap "rawkafka"
    exit 1
  fi

  info "ponged"

  run curl -sv -X POST -H "Accept: application/json" "$url/random-endpoint"
  http_status="$(get-http-status "$output")"
  info "CURL OUTPUT (http status=$http_status)"
  echo "$output" >&3
  info "LOG FILE"
  cat "$log_file" >&3

  if [ "$status" -ne 0 ] || [ "$http_status" -ne 200 ]; then
    cd "$BATS_TEST_DIRNAME"

    docker-compose ps | tap "docker ps"
    docker-compose config | tap "docker config"
    docker-compose logs kafka-rest | tap 'logs kafka-rest'
    exit 1
  fi

  info "request succeeded"
}
