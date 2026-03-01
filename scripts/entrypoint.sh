#!/bin/sh
set -eu

if [ "$#" -eq 0 ]; then
  set -- /usr/local/bin/gmb-bot
fi

if [ -n "${USER_ID:-}" ] && [ "$(id -u)" = "0" ]; then
  GROUP_ID="${GROUP_ID:-$USER_ID}"
  export HOME="${HOME:-/tmp}"
  exec gosu "${USER_ID}:${GROUP_ID}" "$@"
fi

exec "$@"
