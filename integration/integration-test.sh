#!/bin/bash
# Integration Test Runner
# Runs integration tests with independent server management

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load .env file if it exists (must be before defaults)
if [ -f .env.test ]; then
  set -a
  source .env.test
  set +a
fi

# Default values (uses pre-set env, .env.test, or default)
TEST_SERVER_URL="${TEST_SERVER_URL:-http://localhost:8088}"
TEST_SERVER_PORT="${TEST_SERVER_PORT:-8088}"
SERVER_LOG_PATH="${SERVER_LOG_PATH:-/tmp/integration-test-server.log}"

# Track whether we started the server
SERVER_STARTED_BY_US=false
SERVER_PID=""

# Function to check if server is running
check_server() {
  if curl -s "$TEST_SERVER_URL/health" > /dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

# Cleanup function to stop server on exit (for signal handling)
cleanup() {
  stop_server
  generate_coverage
  exit 0
}

# Function to stop server without exiting (for normal shutdown)
stop_server() {
  echo ""
  echo "Stopping server..."
  if [ -n "$SERVER_PID" ]; then
    # Send SIGTERM for graceful shutdown (allows coverage flush)
    kill -TERM $SERVER_PID 2>/dev/null || true
    # Wait up to 5 seconds for graceful shutdown
    for i in {1..10}; do
      if ! kill -0 $SERVER_PID 2>/dev/null; then
        break
      fi
      sleep 0.5
    done
    # Force kill if still running
    kill -9 $SERVER_PID 2>/dev/null || true
    wait $SERVER_PID 2>/dev/null || true
  fi
  echo -e "${GREEN}✓ Server stopped${NC}"
}

# Function to generate coverage report
generate_coverage() {
  # Wait a moment for coverage data to be written
  sleep 1

  # Generate coverage report
  echo ""
  echo "================================"
  echo "Generating coverage report..."
  echo "================================"

  cd ../server
  if [ -d "covdata" ] && [ "$(ls -A covdata)" ]; then
    # Find and concatenate all coverage files
    coverage_files=$(ls covdata/coverage.* 2>/dev/null)
    if [ -n "$coverage_files" ]; then
      # Merge all coverage files
      cat $coverage_files > coverage.out
      # Display coverage percentage
      echo ""
      echo -e "${GREEN}Total Coverage:${NC}"
      go tool cover -func=coverage.out | tail -1
      # Generate HTML report
      go tool cover -html=coverage.out -o=coverage.html
      echo ""
      echo "Coverage report generated: server/coverage.html"
    else
      echo "No coverage files found in covdata/"
    fi
  else
    echo "No coverage data found in covdata/"
  fi
  cd ../integration
}

# Trap SIGINT and SIGTERM to cleanup
trap cleanup SIGINT SIGTERM

# Main execution
echo "================================"
echo "Integration Test Runner"
echo "================================"
echo ""
echo "Test server URL: $TEST_SERVER_URL"
echo ""

# Check if server is running
if check_server; then
  echo -e "${GREEN}✓ Test server is already running at $TEST_SERVER_URL${NC}"
  echo ""
  echo "Press Ctrl+C to stop tests"
else
  # Drop and recreate database
  echo "Resetting test database..."
  dropdb codetogether_test 2>/dev/null || true
  createdb codetogether_test || {
    echo -e "${RED}✗ Error: Failed to create test database${NC}"
    echo "Make sure PostgreSQL is running and you have permissions"
    exit 1
  }
  echo -e "${GREEN}✓ Test database reset${NC}"

# Run the server in background
pushd .
cd ../server

# Check if server binary exists, if not build it
if [ ! -f "bin/codetogether_test.cover" ] || [ "../server/go.mod" -nt "bin/codetogether_test.cover" ]; then
  echo "Building test server binary..."
  mkdir -p bin
  go test -c -cover -covermode=set -coverpkg=./... -o bin/codetogether_test.cover . 2>&1 | head -20
  if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo -e "${RED}✗ Error: Failed to build test server binary${NC}"
    popd
    exit 1
  fi
  echo -e "${GREEN}✓ Test server built${NC}"
fi

# Clean old coverage data to ensure fresh collection
rm -rf covdata/*
mkdir -p covdata

# Run the test binary with coverage collection to a file
# Use process ID to create unique coverage file name, override port to 8088
echo ""
echo "Starting test server on port $TEST_SERVER_PORT..."
echo "Log file: $SERVER_LOG_PATH"
TEST_COVERAGE_SERVER=1 PORT=$TEST_SERVER_PORT ./bin/codetogether_test.cover \
  -test.v -test.run TestCoverageServer -test.coverprofile=covdata/coverage.$$ \
  > "$SERVER_LOG_PATH" 2>&1 &
SERVER_PID=$!
popd

# Wait for server to be ready
echo "Waiting for server to start (PID: $SERVER_PID)..."
for i in {1..30}; do
  if check_server; then
    SERVER_STARTED_BY_US=true
    echo -e "${GREEN}✓ Test server started${NC} (PID: $SERVER_PID, Port: $TEST_SERVER_PORT)"
    break
  fi
  sleep 1
done

# Check if server started successfully
if ! check_server; then
  echo -e "${RED}✗ Failed to start test server${NC}"
  echo "Check logs: tail -100 $SERVER_LOG_PATH"
  exit 1
fi

# Run initial setup
echo ""
echo "Running initial setup..."
SETUP_RESPONSE=$(curl -s -X POST http://localhost:$TEST_SERVER_PORT/api/v1/setup/admin \
  -H "Content-Type: application/json" \
  -d '{"organization_name":"Test Org","admin_email":"admin@example.com","admin_name":"Test Admin","admin_password":"AdminPassword123!"}')

# Check if setup was successful or already done
if echo "$SETUP_RESPONSE" | grep -q "already"; then
  echo "Setup already completed (database has existing data)"
elif echo "$SETUP_RESPONSE" | grep -q "error\|Error\|failed"; then
  echo "Setup response: $SETUP_RESPONSE"
else
  echo -e "${GREEN}✓ Initial setup completed${NC}"
fi

echo ""
echo "Test server ready with admin user: admin@example.com / AdminPassword123!"
echo ""
echo "Press Ctrl+C to stop the server"
fi

# Run integration tests
echo ""
echo "Running integration tests..."
echo ""

# Run tests (allow failures so we can generate coverage)
set +e
go test -count=1 -v github.com/code-together/integration
TEST_EXIT_CODE=$?
set -e

echo ""
echo "================================"
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo -e "${GREEN}✓ All integration tests passed${NC}"
else
  echo -e "${RED}✗ Some integration tests failed${NC}"
fi
echo "================================"

# Cleanup if we started the server
if [ "$SERVER_STARTED_BY_US" = "true" ]; then
  stop_server
  generate_coverage
fi

echo ""
exit $TEST_EXIT_CODE
