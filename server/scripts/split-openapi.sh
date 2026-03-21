#!/bin/bash
# Split swagger.yaml into Member and Manager API specs

set -e

DOCS_DIR="$(cd "$(dirname "$0")/.." && pwd)/docs"
SWAGGER_FILE="$DOCS_DIR/swagger.yaml"

if [ ! -f "$SWAGGER_FILE" ]; then
    echo "Error: $SWAGGER_FILE not found. Run 'make swag-init' first."
    exit 1
fi

echo "Splitting OpenAPI specs..."

# Manager API = full spec (all endpoints including manager-only)
cp "$SWAGGER_FILE" "$DOCS_DIR/server_api_manager.yaml"
echo "✓ Generated Manager API (full spec): server_api_manager.yaml"

# Member API = copy of full spec for now
# TODO: Filter to only member-accessible endpoints (those with "Member" tag or public)
# For now, clients can use the tags to determine access levels
cp "$SWAGGER_FILE" "$DOCS_DIR/server_api.yaml"
echo "✓ Generated Member API: server_api.yaml"

echo ""
echo "Note: Both specs currently contain all endpoints."
echo "Endpoints are tagged with 'Member' or 'Manager' for access control."
echo "Manager API includes all endpoints; Member clients should use 'Member' tagged endpoints."

echo "✓ OpenAPI specs ready"
