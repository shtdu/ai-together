#!/bin/zsh
# Load mock data into the database for testing
#
# Usage: ./load_mock_data.sh [database_url]
#
# If DATABASE_URL is not provided, it will use the value from .env file

set -e

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${ZSH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Default database URL (can be overridden by .env or argument)
DEFAULT_DB_URL="postgres://admin:pgsql@localhost:5432/codetogether?sslmode=disable"

# Check if .env file exists and load it
if [[ -f "$PROJECT_ROOT/.env" ]]; then
    echo "Loading .env file..."
    # Extract DATABASE_URL from .env
    ENV_DB_URL=$(grep "^DATABASE_URL=" "$PROJECT_ROOT/.env" | cut -d '=' -f2-)
    if [[ -n "$ENV_DB_URL" ]]; then
        DB_URL="$ENV_DB_URL"
    else
        DB_URL="$DEFAULT_DB_URL"
    fi
else
    DB_URL="$DEFAULT_DB_URL"
fi

# Allow override via argument
if [[ -n "$1" ]]; then
    DB_URL="$1"
fi

echo "Database URL: $DB_URL"
echo ""

# Parse database URL using zsh parameter expansion
# Example: postgres://user:password@localhost:5432/database?sslmode=disable

# Remove protocol prefix
DB_URL_CLEANED="${DB_URL#postgres://}"
DB_URL_CLEANED="${DB_URL_CLEANED#postgresql://}"

# Split on @ to separate credentials from host:port/database
if [[ "$DB_URL_CLEANED" == *@* ]]; then
    CREDENTIALS="${DB_URL_CLEANED%@*}"
    HOST_PART="${DB_URL_CLEANED#*@}"

    # Split credentials into user and password
    if [[ "$CREDENTIALS" == *:* ]]; then
        DB_USER="${CREDENTIALS%:*}"
        DB_PASS="${CREDENTIALS#*:}"
    else
        DB_USER="$CREDENTIALS"
        DB_PASS=""
    fi
else
    HOST_PART="$DB_URL_CLEANED"
    DB_USER="admin"
    DB_PASS="pgsql"
fi

# Split host:port/database part
if [[ "$HOST_PART" == *"/"* ]]; then
    HOST_PORT="${HOST_PART%/*}"
    DB_WITH_QUERY="${HOST_PART#*/}"

    # Remove query parameters from database name
    DB_NAME="${DB_WITH_QUERY%\?*}"
else
    HOST_PORT="$HOST_PART"
    DB_NAME="codetogether"
fi

# Split host and port
if [[ "$HOST_PORT" == *:* ]]; then
    DB_HOST="${HOST_PORT%:*}"
    DB_PORT="${HOST_PORT#*:}"
else
    DB_HOST="$HOST_PORT"
    DB_PORT="5432"
fi

echo "Connecting to:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo "  Database: $DB_NAME"
echo ""

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql is not installed or not in PATH"
    echo "Install PostgreSQL client tools:"
    echo "  macOS: brew install postgresql"
    echo "  Ubuntu: sudo apt-get install postgresql-client"
    exit 1
fi

# Get absolute path to SQL file
SQL_FILE="$SCRIPT_DIR/server/scripts/mock_data.sql"
if [[ ! -f "$SQL_FILE" ]]; then
    echo "Error: SQL file not found at $SQL_FILE"
    echo "Current directory: $(pwd)"
    echo "Script directory: $SCRIPT_DIR"
    exit 1
fi

echo "SQL file: $SQL_FILE"
echo ""

# Load the mock data
echo "Loading mock data..."
PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$SQL_FILE"

echo ""
echo "✅ Mock data loaded successfully!"
echo ""
echo "Test credentials:"
echo "  Email: test-member@example.com"
echo "  Password: password123"
echo ""
echo "  Email: test-admin@example.com"
echo "  Password: password123"
echo ""
echo "You can now use these credentials to test the sync functionality."
