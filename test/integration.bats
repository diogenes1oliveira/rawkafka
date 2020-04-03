#!/usr/bin/env bats

RAWKAFKA_CMD="${RAWKAFKA_CMD:-./build/rawkafka}"
RAWKAFKA_PID=
export RAWKAFKA_PORT=

load common

function setup {
  [ -n "$RAWKAFKA_SCHEMA_REGISTRY_URL" ] || fatal 'Empty $RAWKAFKA_SCHEMA_REGISTRY_URL'
  [ -n "$RAWKAFKA_REST_ENDPOINT" ] || fatal 'Empty $RAWKAFKA_REST_ENDPOINT'
  [ -n "$RAWKAFKA_CMD" ] || fatal 'Empty $RAWKAFKA_CMD'

  RAWKAFKA_PID=""
  export RAWKAFKA_PORT="$(get-open-port)"
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
  $RAWKAFKA_CMD 3>&- 2>&1 >> "$log_file" &
  RAWKAFKA_PID="$!"

  info "running rawkfaka with PID = $RAWKAFKA_PID"

  info "waiting for ping..."
  if ! wait-for-ping "$url/ping" 20; then
    cat "$log_file" | tap "rawkafka"
    exit 1
  fi

  info "ponged"

  run curl -X POST -H "X-Test: 123" "$url/random-endpoint"

  if [ "$status" -ne 0 ]; then
    info "$output"
    cd "$BATS_TEST_DIRNAME"
    pwd | tap "pwd"
    
    curl -I "$RAWKAFKA_SCHEMA_REGISTRY_URL/subjects/NoSchema-key" 2>&1 | tap "curl schema-registry"
    docker-compose ps | tap "docker ps"
    docker-compose config | tap "docker config"
    docker-compose logs schema-registry | tap 'logs schema-registry'
    exit 1
  fi

  info "request succeeded"
}
