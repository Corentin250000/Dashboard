#!/bin/sh

set -e
if [ -z "$DOPPLER_TOKEN" ] && [ "$1" = "doppler" ] && [ "$2" = "run" ] && [ "$3" = "--" ]; then
  echo "DOPPLER_TOKEN not set: starting without Doppler"
  shift 3
fi
exec "$@"
