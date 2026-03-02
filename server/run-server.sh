#!/bin/bash
# Load .env file if it exists
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

# Run the server
exec ./tmp/server