#!/bin/bash

OUTPUT_DIR="$1"

if [ -z "$OUTPUT_DIR" ]; then
  OUTPUT_DIR="."
fi

if [ ! -d "$OUTPUT_DIR" ]; then
  echo "Error: $OUTPUT_DIR is not a directory"
  exit 1
fi

for session_dir in "$OUTPUT_DIR"/*/; do
  if [ -d "$session_dir" ]; then
    session_name=$(basename "$session_dir")
    echo "=== Session: $session_name ==="

    ls "$session_dir"/*.json 2>/dev/null | \
      sed 's/.*_\(.*\)\.json/\1/' | \
      sort | uniq -c | sort -rn

    echo ""
  fi
done
