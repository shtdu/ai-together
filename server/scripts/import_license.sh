#!/bin/bash

# License Import Script
# This script imports a license by calling the HTTP API.
# It uses email/password authentication to obtain a JWT token, then activates the license.

set -e

# Default values
SERVER_URL="${SERVER_URL:-http://localhost:8080}"
TIMEOUT=30
FORCE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] [EMAIL PASSWORD LICENSE_FILE]

Import a license by calling the HTTP API.

Arguments:
  EMAIL          User email (used as username)
  PASSWORD       User password
  LICENSE_FILE   Path to license file (.pem)

Options:
  --server-url URL    Server URL (default: http://localhost:8080)
  --force             Skip confirmation when overwriting a valid license
  -h, --help          Show this help message

Environment Variables:
  LICENSE_EMAIL       User email (alternative to first argument)
  LICENSE_PASSWORD    User password (alternative to second argument)
  LICENSE_FILE        Path to license file (alternative to third argument)
  SERVER_URL          Server URL (alternative to --server-url)

Examples:
  # Via arguments
  $0 manager@example.com mypassword /path/to/license.pem

  # Via environment variables
  export LICENSE_EMAIL="manager@example.com"
  export LICENSE_PASSWORD="mypassword"
  export LICENSE_FILE="/path/to/license.pem"
  $0

  # With custom server URL
  $0 manager@example.com mypassword /path/to/license.pem --server-url https://api.example.com

  # Force overwrite without confirmation
  $0 manager@example.com mypassword /path/to/license.pem --force

EOF
    exit 0
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --server-url)
            SERVER_URL="$2"
            shift 2
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        -*)
            print_error "Unknown option: $1"
            usage
            ;;
        *)
            # Positional arguments
            if [[ -z "$EMAIL" ]]; then
                EMAIL="$1"
            elif [[ -z "$PASSWORD" ]]; then
                PASSWORD="$1"
            elif [[ -z "$LICENSE_FILE_ARG" ]]; then
                LICENSE_FILE_ARG="$1"
            else
                print_error "Too many arguments"
                usage
            fi
            shift
            ;;
    esac
done

# Use environment variables if arguments not provided
EMAIL="${EMAIL:-$LICENSE_EMAIL}"
PASSWORD="${PASSWORD:-$LICENSE_PASSWORD}"
LICENSE_FILE_ARG="${LICENSE_FILE_ARG:-$LICENSE_FILE}"

# Validate required parameters
if [[ -z "$EMAIL" ]]; then
    print_error "Email is required (provide as argument or set LICENSE_EMAIL env var)"
    usage
fi

if [[ -z "$PASSWORD" ]]; then
    print_error "Password is required (provide as argument or set LICENSE_PASSWORD env var)"
    usage
fi

if [[ -z "$LICENSE_FILE_ARG" ]]; then
    print_error "License file path is required (provide as argument or set LICENSE_FILE env var)"
    usage
fi

# Check if license file exists
if [[ ! -f "$LICENSE_FILE_ARG" ]]; then
    print_error "License file not found: $LICENSE_FILE_ARG"
    exit 1
fi

# Read license key from file
LICENSE_KEY=$(cat "$LICENSE_FILE_ARG")

if [[ -z "$LICENSE_KEY" ]]; then
    print_error "License file is empty: $LICENSE_FILE_ARG"
    exit 1
fi

# Trim leading/trailing whitespace but preserve internal newlines for PEM format
LICENSE_KEY=$(echo "$LICENSE_KEY" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')

print_info "Server URL: $SERVER_URL"
print_info "Email: $EMAIL"
print_info "License File: $LICENSE_FILE_ARG"
echo ""

# Step 1: Login to get JWT token
print_info "Logging in to get access token..."

LOGIN_RESPONSE=$(curl -s -X POST "${SERVER_URL}/auth/login" \
  --max-time "$TIMEOUT" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")

# Check if curl succeeded
if [[ $? -ne 0 ]]; then
    print_error "Failed to connect to server at ${SERVER_URL}"
    exit 1
fi

# Extract access token using jq or fallback to grep/sed
if command -v jq &> /dev/null; then
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')
    ERROR_MSG=$(echo "$LOGIN_RESPONSE" | jq -r '.error // .message // empty')
else
    # Fallback: use grep and sed
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*"' | sed 's/"access_token":"\(.*\)"/\1/')
    ERROR_MSG=$(echo "$LOGIN_RESPONSE" | grep -o '"error":"[^"]*"' | sed 's/"error":"\(.*\)"/\1/')
    ERROR_MSG="${ERROR_MSG:-$(echo "$LOGIN_RESPONSE" | grep -o '"message":"[^"]*"' | sed 's/"message":"\(.*\)"/\1/')}"
fi

if [[ -z "$ACCESS_TOKEN" ]]; then
    print_error "Login failed"
    if [[ -n "$ERROR_MSG" ]]; then
        print_error "Server response: $ERROR_MSG"
    else
        print_error "Server response: $LOGIN_RESPONSE"
    fi
    exit 1
fi

print_success "Login successful"
echo ""

# Step 2: Check current license status
print_info "Checking current license status..."

CURRENT_LICENSE_RESPONSE=$(curl -s -X GET "${SERVER_URL}/api/v1/license" \
  --max-time "$TIMEOUT" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}")

# Check response
if command -v jq &> /dev/null; then
    HAS_ACTIVE_LICENSE=$(echo "$CURRENT_LICENSE_RESPONSE" | jq -r '.status.has_active_license // false')
    DAYS_REMAINING=$(echo "$CURRENT_LICENSE_RESPONSE" | jq -r '.status.days_remaining // 0')
    CURRENT_TIER=$(echo "$CURRENT_LICENSE_RESPONSE" | jq -r '.license.tier_name // "Unknown"')
    CUSTOMER_NAME=$(echo "$CURRENT_LICENSE_RESPONSE" | jq -r '.license.customer_name // ""')
    EXPIRES_AT=$(echo "$CURRENT_LICENSE_RESPONSE" | jq -r '.license.expires_at // ""')
else
    HAS_ACTIVE_LICENSE=$(echo "$CURRENT_LICENSE_RESPONSE" | grep -o '"has_active_license":true' | wc -l | tr -d ' ')
    if [[ "$HAS_ACTIVE_LICENSE" != "1" ]]; then
        HAS_ACTIVE_LICENSE="false"
    fi
    DAYS_REMAINING=$(echo "$CURRENT_LICENSE_RESPONSE" | grep -o '"days_remaining":[0-9]*' | sed 's/"days_remaining":\([0-9]*\)/\1/')
    DAYS_REMAINING="${DAYS_REMAINING:-0}"
    CURRENT_TIER=$(echo "$CURRENT_LICENSE_RESPONSE" | grep -o '"tier_name":"[^"]*"' | sed 's/"tier_name":"\([^"]*\)"/\1/')
    CUSTOMER_NAME=$(echo "$CURRENT_LICENSE_RESPONSE" | grep -o '"customer_name":"[^"]*"' | sed 's/"customer_name":"\([^"]*\)"/\1/')
    EXPIRES_AT=$(echo "$CURRENT_LICENSE_RESPONSE" | grep -o '"expires_at":"[^"]*"' | sed 's/"expires_at":"\([^"]*\)"/\1/')
fi

if [[ "$HAS_ACTIVE_LICENSE" == "true" && "$DAYS_REMAINING" -gt 0 ]]; then
    print_info "Current License Status:"
    echo "  Customer: ${CUSTOMER_NAME}"
    echo "  Tier: ${CURRENT_TIER}"
    echo "  Expires: ${EXPIRES_AT}"
    echo "  Days Remaining: ${DAYS_REMAINING}"
    echo ""

    if [[ "$FORCE" == "true" ]]; then
        print_info "--force specified, proceeding with license overwrite..."
    else
        # Ask for confirmation
        echo -e "${YELLOW}WARNING: You are about to overwrite a valid license!${NC}"
        read -p "Do you want to proceed? (y/N): " -n 1 -r
        echo ""
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "License import cancelled"
            exit 0
        fi
    fi
elif [[ "$HAS_ACTIVE_LICENSE" == "true" && "$DAYS_REMAINING" -eq 0 ]]; then
    print_info "Current license has expired. Proceeding with activation..."
else
    print_info "No active license found. Proceeding with activation..."
fi

echo ""

# Step 3: Activate the license
print_info "Activating license..."

# Construct JSON payload using jq to handle special characters correctly
if command -v jq &> /dev/null; then
    JSON_PAYLOAD=$(echo "$LICENSE_KEY" | jq -Rs '{license_key: .}')
else
    # Fallback: simple escape (may not handle all edge cases)
    JSON_PAYLOAD="{\"license_key\":\"${LICENSE_KEY}\"}"
fi

ACTIVATE_RESPONSE=$(curl -s -X POST "${SERVER_URL}/api/v1/license/activate" \
  --max-time "$TIMEOUT" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${ACCESS_TOKEN}" \
  -d "$JSON_PAYLOAD")

# Check response
if command -v jq &> /dev/null; then
    MESSAGE=$(echo "$ACTIVATE_RESPONSE" | jq -r '.message // empty')
    ERROR=$(echo "$ACTIVATE_RESPONSE" | jq -r '.error // empty')
else
    MESSAGE=$(echo "$ACTIVATE_RESPONSE" | grep -o '"message":"[^"]*"' | sed 's/"message":"\(.*\)"/\1/')
    ERROR=$(echo "$ACTIVATE_RESPONSE" | grep -o '"error":"[^"]*"' | sed 's/"error":"\(.*\)"/\1/')
fi

if [[ -n "$ERROR" ]]; then
    print_error "License activation failed: $ERROR"
    exit 1
elif [[ -n "$MESSAGE" ]]; then
    print_success "$MESSAGE"
else
    print_error "Unexpected response from server"
    echo "Response: $ACTIVATE_RESPONSE"
    exit 1
fi

echo ""
print_success "License imported successfully!"
