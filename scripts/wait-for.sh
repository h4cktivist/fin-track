#!/usr/bin/env sh

set -e

HOST="$1"
PORT="$2"

if [ -z "$HOST" ] || [ -z "$PORT" ]; then
  echo "usage: wait-for.sh host port"
  exit 1
fi

until nc -z "$HOST" "$PORT"; do
  echo "waiting for $HOST:$PORT..."
  sleep 1
done

echo "$HOST:$PORT is available"

