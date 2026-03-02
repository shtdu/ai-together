#!/bin/bash

# Test script for hook collection endpoint
# BASE_URL can be overridden by environment variable
BASE_URL="${BASE_URL:-http://localhost:18100}"

echo "Testing hook collection endpoint at $BASE_URL"
echo

# Test 1: Valid event
echo "Test 1: Valid event (should succeed)"
RESPONSE=$(curl -s -X POST "${BASE_URL}/collect/claude" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test-session-001","hook_event_name":"UserPromptSubmit","cwd":"/path/to/workspace"}')
echo "Response: $RESPONSE"
echo

# Test 2: Missing session_id
echo "Test 2: Missing session_id (should fail)"
RESPONSE=$(curl -s -X POST "${BASE_URL}/collect/claude" \
  -H "Content-Type: application/json" \
  -d '{"hook_event_name":"UserPromptSubmit"}')
echo "Response: $RESPONSE"
echo

# Test 3: Missing hook_event_name
echo "Test 3: Missing hook_event_name (should fail)"
RESPONSE=$(curl -s -X POST "${BASE_URL}/collect/claude" \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test-session-002"}')
echo "Response: $RESPONSE"
echo

# Test 4: Invalid JSON
echo "Test 4: Invalid JSON (should fail)"
RESPONSE=$(curl -s -X POST "${BASE_URL}/collect/claude" \
  -H "Content-Type: application/json" \
  -d '{invalid json}')
echo "Response: $RESPONSE"
echo

# Test 5: Verify in database
echo "Test 5: Verify stored events in database"
if command -v sqlite3 &> /dev/null; then
  sqlite3 ~/.code-together/hook-events.db \
    "SELECT session_id, tool_name, event_name, ts FROM events LIMIT 5;"
else
  echo "sqlite3 not found, skipping database verification"
  echo "Install with: brew install sqlite3 (macOS) or apt-get install sqlite3 (Linux)"
fi
echo

echo "Tests completed"
