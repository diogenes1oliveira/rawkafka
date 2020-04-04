#!/usr/bin/env bash

export MAX_WAIT_TIME=120

function wait-for-ping {
  local url="$1"
  local tries=0
  local http_status=0
  local t0="$(get-uptime)"

  info "trying to ping $url"

  while true; do
    dt="$(bc -l <<< "$(get-uptime) - ${t0}")"
    sleep 1

    info ""
    info ""
    info ""
    info ""
    info "ping try (${dt}s/${MAX_WAIT_TIME}s) (last http: $http_status)"
    info ""

    if (( $(bc -l <<< "${dt} > ${MAX_WAIT_TIME}" ) )); then
      info "last curl exit code is $exit_code"
      info "last http status is $http_status"
      echo "$fout" >&3

      [ "$exit_code" -eq 0 ] || info "curl exit code: $exit_code != 0"
      [ "$http_status" -eq 200 ] || info "HTTP status code: $http_status != 200"
      grep -q pong "$fout" || info "no pong in output"

      error "Timeout while waiting for pong: ${MAX_WAIT_TIME}s"
    fi

    info "trying to curl $url"
    run curl -m 1 -sv "$url"
    exit_code="$status"

    if [ "$exit_code" -ne 0 ]; then
      info "temp fail in exit code: $exit_code"
      continue
    fi

    info "trying to get HTTP status"
    http_status="$(get-http-status "$output")"
    if [ "$http_status" -ne 200 ]; then
      info "temp fail in HTTP code: $http_status"
      continue
    fi

    info "trying to grep ping"
    echo "$output" >&3
    info "grepping outside if in $fout"
    if [[ "$output" =~ ping ]] ; then
      info "FOUND ping"
    else
      info "COULDN'T FIND ping"
    fi
    info "grepped outside if"

    if ! [[ "$output" =~ ping ]] ; then
      info "ping failed"
      continue
    fi

    info "ping ponged"
    return 0
  done

  info "SHOULD NEVER REACH HERE"
  exit 1
}

function wait-for-status-code {
  local url="$1"
  local wanted_status="${2:-404}"
  local tries=0
  local exit_code=0
  local t0="$(get-uptime)"

  info "waiting for $wanted_status at $url"

  while true; do
    dt="$(bc -l <<< "$(get-uptime) - ${t0}")"
    sleep 1

    if (( $(bc -l <<< "${dt} > ${MAX_WAIT_TIME}" ) )); then
      info "last curl exit code is $exit_code"
      info "last http status is $http_status"
      echo "$output" | tap "failed output"

      [ "$exit_code" -eq 0 ] || info "curl exit code: $exit_code != 0"
      [ "$http_status" -eq "$wanted_status" ] || info "HTTP status code: $http_status != $wanted_status"

      error "Timeout while waiting for $status: ${MAX_WAIT_TIME}s"
    fi

    run curl -m 1 -s -I -H 'Accept: application/json' "$url"
    exit_code="$status"

    if [ "$exit_code" -ne 0 ]; then
      info "temp fail in exit code: $exit_code"
      continue
    fi

    info "trying to get HTTP status"
    http_status="$(get-http-status "$output")"
    if [ -z "$http_status" ]; then
      echo "$output" | tap "no http status"
    fi

    if [ "$http_status" -ne "$wanted_status" ]; then
      info "temp fail in HTTP code: $http_status"
      continue
    fi

    info "responded with $http_status"
    return 0
  done

  info "SHOULD NEVER REACH HERE"
  exit 1
}

function info {
  echo "# INFO: $@" >&3
}

function fatal {
  echo "# FATAL: $@" >&3
  exit 1
}

function error {
  echo "# ERROR: $@" >&3
  return 1
}

function tap {
  cat | sed "s/^/# $1: /" >&3
}

function get-http-status {
  echo "$1" | sed -n 's;^<\?[[:space:]]*HTTP/[0-9.]\+[[:space:]]*\([[:digit:]]\+\)[[:space:]].*;\1;p'
}

function get-uptime {
  cat /proc/uptime | awk '{print $1}'
}
