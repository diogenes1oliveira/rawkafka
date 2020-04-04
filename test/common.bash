#!/usr/bin/env bash

export MAX_WAIT_TIME=10

function wait-for-ping {
  local url="$1"
  local tries=0

  while ! ( curl -m 1 -s "$url" 2>&1 | grep -q "${3:-pong}" ) ; do
    tries="$((tries+1))"

    if [ "$tries" -ge "$MAX_WAIT_TIME" ]; then
      if curl -m 1 -sv "$url" 2>&1 | tap "curl"; then
        break
      fi
      fatal "Timeout expired while waiting for pong: ${MAX_WAIT_TIME}s"
    fi
    sleep 1
  done

}

function wait-for-status-code {
  local url="$1"
  local status="${2:-404}"
  local tries=0
  local output="$BATS_TMPDIR/curl.log"
  local exit_code=0

  while true; do
    (
      set +e
      curl -m 1 -s -I -H 'Accept: application/json' "$url" 2>&1 > "$output"
      exit_code="$?"
    )

    if [ "$exit_code" -eq 0 ] && [ "$(cat "$output" | head -n1 | awk '{print $2}')" = "$status" ]; then
      break
    fi

    tries="$((tries+1))"

    if [ "$tries" -ge "$MAX_WAIT_TIME" ]; then
      fatal "Timeout expired while waiting for a $status: ${MAX_WAIT_TIME}s"
    fi

    sleep 1
  done

}

function info {
  echo "# INFO: $@" >&3
}

function fatal {
  echo "# FATAL: $@" >&3
  exit 1
}

function tap {
  cat | sed "s/^/# $1: /" >&3
}

function get-open-port {
  python -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'
}
