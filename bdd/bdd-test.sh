#!/bin/bash
# BDD Test Runner
# Runs all BDD scenarios using godog framework

set -e

# Load .env file if it exists
if [ -f .env.test ]; then
  set -a
  source .env.test
  set +a
fi

# Default values
TEST_SERVER_URL="${TEST_SERVER_URL:-http://localhost:8088}"
GODOG_FORMAT="${GODOG_FORMAT:-pretty}"
GODOG_TAGS="${GODOG_TAGS:-~@wip}"
START_SERVER="${START_SERVER:-true}"

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

# Function to start the test server
start_server() {
  echo "Starting test server..."

  # Check if integration test server script exists
  if [ ! -f "../integration/test-server.sh" ]; then
    echo "Error: test-server.sh not found in ../integration/"
    exit 1
  fi

  # Start the server in background
  cd ../integration
  ./test-server.sh > /tmp/test-server.log 2>&1 &
  SERVER_PID=$!
  cd ../bdd

  # Wait for server to be ready
  echo "Waiting for server to start..."
  for i in {1..30}; do
    if check_server; then
      SERVER_STARTED_BY_US=true
      echo "✓ Test server started (PID: $SERVER_PID)"
      return 0
    fi
    sleep 1
  done

  echo "✗ Failed to start test server"
  echo "Check logs: tail -100 /tmp/test-server.log"
  exit 1
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
    echo "Start the test server:"
    echo "  cd ../integration && ./test-server.sh"
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

# Run tests
go test -v ./...
TEST_EXIT_CODE=$?

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
fi

exit $TEST_EXIT_CODE

