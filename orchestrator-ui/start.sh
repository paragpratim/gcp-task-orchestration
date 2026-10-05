#!/bin/sh
set -eu

if [ "${ENABLE_IAM_AUTH:-true}" = "true" ] || [ "${ENABLE_IAM_AUTH:-true}" = "1" ]; then
  echo "Starting UI proxy in authenticated mode"
else
  echo "Starting UI proxy in local mode"
fi

node /app/server.js
