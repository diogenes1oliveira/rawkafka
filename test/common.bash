#!/usr/bin/env bash

function wait-for-ping {
  local url="$1"
  local max_wait_time="${2:-10}"
  local tries=0

  while ! ( curl -m 1 -s "$url" 2>&1 | grep -q pong ) ; do
    tries="$((tries+1))"

    if [ "$tries" -ge "$max_wait_time" ]; then
      if curl -m 1 -sv "$url"; then
        break
      fi
      fatal "Timeout expired while waiting for pong: ${max_wait_time}s"
    fi
    sleep 1
  done

}

function info {
  echo "# INFO: $@" >&3
}

function fatal {
  echo "# FATAL: $@" >&3
  return 1
}

function tap {
  cat | sed "s/^/# $1: /" >&3
}

function get-open-port {
  python -c 'import socket; s=socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()'
}
