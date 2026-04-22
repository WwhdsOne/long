#!/bin/sh
set -eu

export LONG_LISTEN_HOST="${LONG_LISTEN_HOST:-127.0.0.1}"
export LONG_LISTEN_PORT="${LONG_LISTEN_PORT:-18080}"

/app/backend/long &
app_pid=$!

cleanup() {
    kill "$app_pid" 2>/dev/null || true
}

trap cleanup INT TERM EXIT

nginx -g 'daemon off;'
