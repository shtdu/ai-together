#!/bin/bash
# BDD Test Runner
# Runs all BDD scenarios using godog framework
# Independent server management - does not depend on integration-test.sh

set -e

# Load .env file if it exists
if [ -f .env.test ]; then
  set -a
  source .env.test
  set +a
fi

# Default values
TEST_SERVER_URL="${TEST_SERVER_URL:-http://localhost:8088}"
TEST_SERVER_PORT="${TEST_SERVER_PORT:-8088}"
GODOG_FORMAT="${GODOG_FORMAT:-pretty}"
GODOG_TAGS="${GODOG_TAGS:-~@wip}"
START_SERVER="${START_SERVER:-true}"
SERVER_LOG_PATH="${SERVER_LOG_PATH:-/tmp/bdd-test-server.log}"
COVERAGE_DIR="${COVERAGE_DIR:-../server/covdata}"
GENERATE_COVERAGE="${GENERATE_COVERAGE:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track whether we started the server
SERVER_STARTED_BY_US=false
SERVER_PID=""

# Cleanup function to stop server on exit (for signal handling)
cleanup() {
  echo ""
  echo "Stopping server..."
  if [ -n "$SERVER_PID" ]; then
    # Send SIGTERM for graceful shutdown
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
  echo "Server stopped"
  exit 0
}

# Function to stop server without exiting (for normal shutdown)
stop_server() {
  echo ""
  echo "Stopping server..."
  if [ -n "$SERVER_PID" ]; then
    # Send SIGTERM for graceful shutdown
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
  echo "Server stopped"
}

# Trap SIGINT and SIGTERM to cleanup
trap cleanup SIGINT SIGTERM

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --format)
      GODOG_FORMAT="$2"
      shift 2
      ;;
    --tags)
      GODOG_TAGS="$2"
      shift 2
      ;;
    --no-server)
      START_SERVER="false"
      shift
      ;;
    --help|-h)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --format FORMAT      godog output format (pretty, junit, etc.)"
      echo "  --tags TAGS          godog tags to filter scenarios (default: ~@wip)"
      echo "  --no-server          Don't start server (expect it to be running)"
      echo "  -h, --help           Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0                                    # Run all BDD tests (auto-start server)"
      echo "  $0 --tags @smoke                     # Run smoke tests"
      echo "  $0 --format junit --tags @smoke      # JUnit output for CI/CD"
      echo "  $0 --no-server                       # Expect server to be already running"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Run '$0 --help' for usage"
      exit 1
      ;;
  esac
done

# Function to check if server is running
check_server() {
  if curl -s "$TEST_SERVER_URL/health" > /dev/null 2>&1; then
    return 0
  else
    return 1
  fi
}

# Function to start the test server (independent, like integration-test.sh)
start_server() {
  echo "Starting test server..."
  echo "Log file: $SERVER_LOG_PATH"

  # Get the server directory path
  SERVER_DIR="../server"
  if [ ! -d "$SERVER_DIR" ]; then
    echo "Error: Server directory not found at $SERVER_DIR"
    exit 1
  fi

  # Drop and recreate database (like integration-test.sh)
  echo "Resetting test database..."
  dropdb codetogether_test 2>/dev/null || true
  createdb codetogether_test || {
    echo "Error: Failed to create test database"
    echo "Make sure PostgreSQL is running and you have permissions"
    exit 1
  }

  # Check if server binary exists, if not build it
  pushd "$SERVER_DIR" > /dev/null

  # Build test binary if it doesn't exist or is outdated
  if [ ! -f "bin/codetogether_test.cover" ] || [ "$SERVER_DIR/go.mod" -nt "bin/codetogether_test.cover" ]; then
    echo "Building test server binary..."
    mkdir -p bin
    go test -c -cover -covermode=set -coverpkg=./... -o bin/codetogether_test.cover . 2>&1 | head -20
    if [ ${PIPESTATUS[0]} -ne 0 ]; then
      echo "Error: Failed to build test server binary"
      popd > /dev/null
      exit 1
    fi
  fi

  # Setup coverage directory and clean old BDD coverage files
  mkdir -p "$COVERAGE_DIR"
  rm -f covdata/coverage.bdd.* 2>/dev/null
  rm -f coverage.bdd.out coverage.bdd.html 2>/dev/null
  COVERAGE_FILE="$COVERAGE_DIR/coverage.bdd.$$"

  # Run the test binary in background with coverage collection
  echo "Starting test server on port $TEST_SERVER_PORT..."
  TEST_COVERAGE_SERVER=1 PORT=$TEST_SERVER_PORT ./bin/codetogether_test.cover \
    -test.v -test.run TestCoverageServer -test.coverprofile="$COVERAGE_FILE" \
    > "$SERVER_LOG_PATH" 2>&1 &
  SERVER_PID=$!
  popd > /dev/null

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

  # Run initial setup (create admin user/organization)
  echo "Running initial setup..."
  SETUP_RESPONSE=$(curl -s -X POST http://localhost:$TEST_SERVER_PORT/api/v1/setup/admin \
    -H "Content-Type: application/json" \
    -d '{"organization_name":"BDD Test Org","admin_email":"admin@example.com","admin_name":"BDD Admin","admin_password":"AdminPassword123!"}')

  # Check if setup was successful or already done
  if echo "$SETUP_RESPONSE" | grep -q "already"; then
    echo "Setup already completed (database has existing data)"
  elif echo "$SETUP_RESPONSE" | grep -q "error\|Error\|failed"; then
    echo "Setup response: $SETUP_RESPONSE"
  else
    echo "Initial setup completed"
  fi

  return 0
}

# Main execution
echo "================================"
echo "BDD Test Runner"
echo "================================"
echo ""
echo "Test server URL: $TEST_SERVER_URL"
echo ""

# Check if server is running
if check_server; then
  echo "✓ Test server is already running at $TEST_SERVER_URL"
else
  if [ "$START_SERVER" = "true" ]; then
    start_server
  else
    echo "✗ Test server is not running and --no-server was specified"
    echo ""
    echo "Either:"
    echo "  1. Omit --no-server to auto-start the server"
    echo "  2. Start the server manually:"
    echo "     cd ../server && PORT=$TEST_SERVER_PORT ./bin/codetogether_test.cover -test.run TestCoverageServer"
    echo ""
    exit 1
  fi
fi

# Run BDD tests
echo ""
echo "Running BDD tests..."
echo "Format: $GODOG_FORMAT"
if [ -n "$GODOG_TAGS" ]; then
  echo "Tags: $GODOG_TAGS"
fi
echo ""

cd godog

# Set environment variables for godog
export GODOG_FORMAT="$GODOG_FORMAT"
export GODOG_TAGS="$GODOG_TAGS"

# Run tests (allow failures so we can generate coverage)
set +e
go test -v ./...
TEST_EXIT_CODE=$?
set -e

cd ..

echo ""
echo "================================"
if [ $TEST_EXIT_CODE -eq 0 ]; then
  echo -e "${GREEN}✓ All BDD tests passed${NC}"
else
  echo -e "${RED}✗ Some BDD tests failed${NC}"
fi
echo "================================"

# Cleanup if we started the server
if [ "$SERVER_STARTED_BY_US" = "true" ]; then
  stop_server

  # Generate coverage report if we started the server and coverage is enabled
  if [ "$GENERATE_COVERAGE" = "true" ]; then
    echo ""
    echo "================================"
    echo "Generating coverage report..."
    echo "================================"

    # Wait a moment for coverage data to be written
    sleep 1

    cd ../server
    if [ -d "covdata" ] && [ "$(ls -A covdata/coverage.bdd.* 2>/dev/null)" ]; then
      # Find and concatenate BDD coverage files
      coverage_files=$(ls covdata/coverage.bdd.* 2>/dev/null)
      if [ -n "$coverage_files" ]; then
        # Merge BDD coverage files - keep first file's header, skip headers in rest
        first_file=true
        for f in $coverage_files; do
          if [ "$first_file" = "true" ]; then
            cat "$f" > coverage.bdd.out
            first_file=false
          else
            grep -v "^mode:" "$f" >> coverage.bdd.out
          fi
        done

        # Display coverage percentage
        echo ""
        echo -e "${GREEN}Server Coverage (BDD Tests):${NC}"
        go tool cover -func=coverage.bdd.out | tail -1

        # Generate HTML report
        go tool cover -html=coverage.bdd.out -o=coverage.bdd.html
        echo ""
        echo "Coverage report generated: server/coverage.bdd.html"

        # Show total coverage
        echo ""
        echo "Breakdown by module:"
        go tool cover -func=coverage.bdd.out | grep -E "^switch-server/" | head -20
      fi
    else
      echo "No BDD coverage files found in covdata/"
    fi
    cd ../bdd
  fi
fi

echo ""
exit $TEST_EXIT_CODE

