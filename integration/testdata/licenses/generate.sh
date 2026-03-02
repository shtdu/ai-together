#!/bin/bash
# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
LICENSE_DIR="$SCRIPT_DIR"
KEY_DIR="$SCRIPT_DIR/../keys"

echo "Generating Open Source License - 1 team, unlimited seats, 7-day retention"
cd "$PROJECT_ROOT"
go run server/scripts/keygen/main.go -customer "Open Source License" -type "opensource" -maxseats -1 -maxteams 1 -retention 7 -days 365 \
  -privkey "$KEY_DIR/private.key" > "$LICENSE_DIR/opensource.pem"

echo "Generating Commercial License - Unlimited teams, unlimited seats, 90-day retention"
go run server/scripts/keygen/main.go -customer "Commercial License" -type "commercial" -maxseats -1 -maxteams -1 -retention 90 -days 365 \
  -privkey "$KEY_DIR/private.key" > "$LICENSE_DIR/commercial.pem"

echo "Generating Expired License - For expiration testing"
go run server/scripts/keygen/main.go -customer "Expired License" -type "opensource" -maxseats -1 -maxteams 1 -retention 7 -days -365 \
  -privkey "$KEY_DIR/private.key" > "$LICENSE_DIR/expired.pem"

echo "Generating Zero Seats License - Edge case testing"
go run server/scripts/keygen/main.go -customer "Zero Seats License" -type "opensource" -maxseats 0 -maxteams 1 -retention 7 -days 365 \
  -privkey "$KEY_DIR/private.key" > "$LICENSE_DIR/zero_seats.pem"

echo "Generating Immediate Expiration License - Boundary testing"
go run server/scripts/keygen/main.go -customer "Immediate Expiry License" -type "opensource" -maxseats -1 -maxteams 1 -retention 7 -days 0 \
  -privkey "$KEY_DIR/private.key" > "$LICENSE_DIR/immediate_expiry.pem"

echo "Generating Valid Signature License - For signature verification testing"
go run server/scripts/keygen/main.go -customer "Test Customer" -type "opensource" -maxseats -1 -maxteams 1 -retention 7 -days 365 \
  -privkey "$KEY_DIR/private.key" > "$LICENSE_DIR/valid_signature.pem"

echo "Generating Invalid Signature License - For signature verification testing"
cd "$LICENSE_DIR"
# First, rename valid signature to invalid signature
mv valid_signature.pem invalid_signature.pem
# Then create invalid version by corrupting signature
# Replace signature block with invalid data
sed -i '' -e 's/-----END LICENSE KEY-----/-----END INVALID SIGNATURE-----/' invalid_signature.pem

echo "License generation complete!"
ls -la "$LICENSE_DIR"
