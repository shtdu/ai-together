#!/bin/bash

# Integration Test Runner for Code Together Server
# This script helps run integration tests with proper setup

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SERVER_URL="${BASE_URL:-http://localhost:9080}"
PG_USER="${PG_USER:-admin}"
PG_HOST="${PG_HOST:-localhost}"
PG_PORT="${PG_PORT:-5432}"
PG_DB="${PG_DB:-codetogether}"
VERBOSE="${VERBOSE:-false}"
COVERAGE="${COVERAGE:-false}"
SPECIFIC_TEST=""
INTERACTIVE="${INTERACTIVE:-true}"

# Password variable (will be set via prompt)
PG_PASSWORD=""

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to prompt for PostgreSQL password
prompt_password() {
    if [ -z "$PGPASSWORD" ]; then
        echo ""
        print_status "$BLUE" "PostgreSQL Authentication"
        print_status "$YELLOW" "Enter PostgreSQL password for user '$PG_USER':"
        read -s -p "Password: " PG_PASSWORD
        echo ""
        echo ""

        # Validate password was entered
        if [ -z "$PG_PASSWORD" ]; then
            print_status "$RED" "✗ Password cannot be empty"
            exit 1
        fi

        # Export password for subprocesses
        export PGPASSWORD="$PG_PASSWORD"
    else
        print_status "$GREEN" "Using password from PGPASSWORD environment variable"
    fi
}

# Function to build database connection string
build_db_url() {
    echo "postgres://${PG_USER}:${PG_PASSWORD}@${PG_HOST}:${PG_PORT}/${PG_DB}?sslmode=disable"
}

# Function to check if server is running
check_server() {
    print_status "$YELLOW" "Checking if server is running on $SERVER_URL..."
    if curl -s -f "$SERVER_URL/health" > /dev/null 2>&1; then
        print_status "$GREEN" "✓ Server is running"
        return 0
    else
        print_status "$RED" "✗ Server is not running on $SERVER_URL"
        print_status "$YELLOW" "Please start the server with: make run-dev"
        return 1
    fi
}

# Function to check if PostgreSQL is running
check_postgres() {
    print_status "$YELLOW" "Checking if PostgreSQL is running on $PG_HOST:$PG_PORT..."
    if pg_isready -h "$PG_HOST" -p "$PG_PORT" > /dev/null 2>&1; then
        print_status "$GREEN" "✓ PostgreSQL is running"
        return 0
    else
        print_status "$RED" "✗ PostgreSQL is not running on $PG_HOST:$PG_PORT"
        print_status "$YELLOW" "Please start PostgreSQL"
        return 1
    fi
}

# Function to verify PostgreSQL credentials
verify_credentials() {
    print_status "$YELLOW" "Verifying PostgreSQL credentials..."
    if psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -c "SELECT 1" > /dev/null 2>&1; then
        print_status "$GREEN" "✓ PostgreSQL credentials are valid"
        return 0
    else
        print_status "$RED" "✗ Invalid PostgreSQL credentials or connection failed"
        print_status "$YELLOW" "Please check your username and password"
        return 1
    fi
}

# Function to check if test database exists
check_test_db() {
    print_status "$YELLOW" "Checking if test database '$PG_DB' exists..."
    if psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d postgres -lqt | cut -d \| -f 1 | grep -qw "$PG_DB"; then
        print_status "$GREEN" "✓ Test database '$PG_DB' exists"
        return 0
    else
        print_status "$YELLOW" "Test database '$PG_DB' does not exist"
        if [ "$INTERACTIVE" = "true" ]; then
            read -p "Create test database? (y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                createdb -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" "$PG_DB"
                print_status "$GREEN" "✓ Test database created"
                return 0
            else
                return 1
            fi
        else
            return 1
        fi
    fi
}

# Function to install dependencies
install_deps() {
    print_status "$YELLOW" "Installing Go dependencies..."
    go mod download
    go mod tidy
    print_status "$GREEN" "✓ Dependencies installed"
}

# Function to run tests
run_tests() {
    local test_args="-v"

    if [ "$VERBOSE" = "true" ]; then
        test_args="$test_args -v"
    fi

    if [ "$COVERAGE" = "true" ]; then
        test_args="$test_args -coverprofile=coverage.out -covermode=atomic"
    fi

    if [ -n "$SPECIFIC_TEST" ]; then
        test_args="$test_args -run $SPECIFIC_TEST"
        print_status "$GREEN" "Running specific test: $SPECIFIC_TEST"
    else
        print_status "$GREEN" "Running all integration tests..."
    fi

    echo ""
    print_status "$YELLOW" "Starting tests..."
    echo ""

    # Build database URL
    TEST_DB_URL=$(build_db_url)

    # Set environment variables for tests
    export BASE_URL="$SERVER_URL"
    export TEST_DATABASE_URL="$TEST_DB_URL"

    # Run the tests
    if go test $test_args ./integration_test.go; then
        echo ""
        print_status "$GREEN" "✓ All tests passed!"
        if [ "$COVERAGE" = "true" ]; then
            print_status "$YELLOW" "Coverage report generated: coverage.out"
            print_status "$YELLOW" "View with: go tool cover -html=coverage.out"
        fi
        return 0
    else
        echo ""
        print_status "$RED" "✗ Some tests failed"
        return 1
    fi
}

# Function to show help
show_help() {
    cat << EOF
Integration Test Runner for Code Together Server

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -c, --coverage          Generate coverage report
    -t, --test TEST_NAME    Run specific test (e.g., TestHealthEndpoint)
    --skip-checks           Skip pre-flight checks
    --non-interactive       Run without prompting (auto-fail on missing database)

PostgreSQL Options (can be set via environment variables):
    -u, --user USER         PostgreSQL username (default: postgres)
    -H, --host HOST         PostgreSQL host (default: localhost)
    -p, --port PORT         PostgreSQL port (default: 5432)
    -d, --database DB       Database name (default: switch_test)

Environment Variables:
    PGPASSWORD              PostgreSQL password (will prompt if not set)
    PG_USER                 PostgreSQL username
    PG_HOST                 PostgreSQL host
    PG_PORT                 PostgreSQL port
    PG_DB                   Database name
    BASE_URL               Server URL (default: http://localhost:9080)
    TEST_DATABASE_URL      Full database connection string (overrides individual settings)
    VERBOSE                Set to "true" for verbose output
    COVERAGE               Set to "true" to generate coverage
    INTERACTIVE            Set to "false" to skip prompts

Examples:
    # Run all tests (will prompt for password)
    $0

    # Run with password pre-set via environment
    PGPASSWORD=mypass $0

    # Run with coverage
    $0 --coverage

    # Run specific test
    $0 --test TestAuthenticationFlow

    # Run with custom PostgreSQL settings
    $0 --user myuser --host db.example.com --port 5432

    # Run with verbose output
    $0 --verbose

    # Run non-interactively (useful for CI/CD)
    PGPASSWORD=mypass $0 --non-interactive

Test Suites:
    TestHealthEndpoint         - Health check endpoint
    TestAuthenticationFlow     - Authentication endpoints
    TestTeamManagement         - Team CRUD operations
    TestProviderManagement     - Provider CRUD operations
    TestUsageTracking          - Usage tracking endpoints
    TestConfigAndProxyEndpoints - Config and proxy endpoints
    TestUnauthorizedAccess     - Authorization tests

EOF
}

# Parse command line arguments
SKIP_CHECKS=false
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -t|--test)
            SPECIFIC_TEST="$2"
            shift 2
            ;;
        -u|--user)
            PG_USER="$2"
            export PG_USER
            shift 2
            ;;
        -H|--host)
            PG_HOST="$2"
            export PG_HOST
            shift 2
            ;;
        -p|--port)
            PG_PORT="$2"
            export PG_PORT
            shift 2
            ;;
        -d|--database)
            PG_DB="$2"
            export PG_DB
            shift 2
            ;;
        --skip-checks)
            SKIP_CHECKS=true
            shift
            ;;
        --non-interactive)
            INTERACTIVE=false
            shift
            ;;
        *)
            print_status "$RED" "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
print_status "$GREEN" "=================================="
print_status "$GREEN" "Integration Test Runner"
print_status "$GREEN" "=================================="
echo ""

# Prompt for password if not already set
if [ -z "$TEST_DATABASE_URL" ]; then
    prompt_password
fi

# Display connection info
if [ "$VERBOSE" = "true" ]; then
    print_status "$BLUE" "Connection Details:"
    echo "  PostgreSQL Host: $PG_HOST:$PG_PORT"
    echo "  PostgreSQL User: $PG_USER"
    echo "  PostgreSQL DB:   $PG_DB"
    echo "  Server URL:      $SERVER_URL"
    echo ""
fi

# Pre-flight checks
if [ "$SKIP_CHECKS" = "false" ]; then
    check_postgres || exit 1
    verify_credentials || exit 1
    check_test_db || exit 1
    check_server || exit 1
fi

# Install dependencies
install_deps || exit 1

# Run tests
run_tests || exit 1

echo ""
print_status "$GREEN" "=================================="
print_status "$GREEN" "Test run completed!"
print_status "$GREEN" "=================================="
