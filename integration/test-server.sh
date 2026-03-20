#!/bin/bash
# Load .env file if it exists
if [ -f .env.test ]; then
  set -a
  source .env.test
  set +a
fi

# Cleanup function to stop server on exit
cleanup() {
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
  echo "Server stopped"

  # Wait a moment for coverage data to be written
  sleep 1

  # Generate coverage report
  echo ""
  echo "Generating coverage report..."
  cd ../server
  if [ -d "covdata" ] && [ "$(ls -A covdata)" ]; then
    # Find and concatenate all coverage files
    coverage_files=$(ls covdata/coverage.* 2>/dev/null)
    if [ -n "$coverage_files" ]; then
      # Merge all coverage files
      cat $coverage_files > coverage.out
      # Display coverage percentage
      go tool cover -func=coverage.out | tail -1
      # Generate HTML report
      go tool cover -html=coverage.out -o=coverage.html
      echo ""
      echo "Coverage report generated: server/coverage.html"
      echo "Total coverage:"
      go tool cover -func=coverage.out | grep total
    else
      echo "No coverage files found in covdata/"
    fi
  else
    echo "No coverage data found in covdata/"
  fi
  cd ../integration

  exit 0
}

# Trap SIGINT and SIGTERM to cleanup
trap cleanup SIGINT SIGTERM

# Drop and recreate database
dropdb codetogether_test 2>/dev/null || true
createdb codetogether_test

# Run the server in background
pushd .
cd ../server
# Clean old coverage data to ensure fresh collection
rm -rf covdata/*
mkdir -p covdata
# Build test binary with coverage instrumentation
go test -c -cover -covermode=set -coverpkg=./... -o bin/codetogether_test.cover .
# Run the test binary with coverage collection to a file
# Use process ID to create unique coverage file name
TEST_COVERAGE_SERVER=1 PORT=8088 ./bin/codetogether_test.cover -test.v -test.run TestCoverageServer -test.coverprofile=covdata/coverage.$$ &
SERVER_PID=$!
popd

# Wait for server to be ready
echo "Waiting for server to start..."
sleep 3

# Run initial setup
echo "Running initial setup..."
curl -s -X POST http://localhost:8088/api/v1/setup/admin \
  -H "Content-Type: application/json" \
  -d '{"organization_name":"Test Org","admin_email":"admin@example.com","admin_name":"Admin","admin_password":"AdminPassword123!"}' | jq .

echo "Test server is ready with admin user: admin@example.com / AdminPassword123!"
echo ""
echo "Press Ctrl+C to stop the server"

# run integration tests

# Wait for server process
wait $SERVER_PID
