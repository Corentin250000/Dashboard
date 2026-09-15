#!/bin/sh

set -e

if [ -n "$DOPPLER_TOKEN" ]; then
  exec doppler run -- docker-entrypoint.sh "$@"
fi

echo "DOPPLER_TOKEN not set: starting postgres without Doppler"
exec docker-entrypoint.sh "$@"
