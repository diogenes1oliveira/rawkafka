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
  info "listening at :$RAWKAFKA_PORT"

  $RAWKAFKA_CMD 3>&- &
  RAWKAFKA_PID="$!"

  info "running rawkfaka with PID = $RAWKAFKA_PID"

  info "waiting for ping..."
  wait-for-ping http://localhost:${RAWKAFKA_PORT}/ping
  info "ponged"

  curl -X POST -H "X-Test: 123" http://localhost:${RAWKAFKA_PORT}/random-endpoint
  info "request succeeded"
}
