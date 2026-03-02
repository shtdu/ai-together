#!/bin/bash
# Code Together Database Restore Script
# 数据库恢复脚本

set -e

# Configuration
DB_CONTAINER="${DB_CONTAINER:-code-together-db}"
DB_USER="${DB_USER:-codetogether}"
DB_NAME="${DB_NAME:-code_together}"

# Usage
usage() {
    echo "Usage: $0 <backup_file.sql.gz>"
    echo ""
    echo "Example:"
    echo "  $0 backups/codetogether_20240115_020000.sql.gz"
    echo ""
    echo "Environment variables:"
    echo "  DB_CONTAINER - Database container name (default: code-together-db)"
    echo "  DB_USER - Database user (default: codetogether)"
    echo "  DB_NAME - Database name (default: code_together)"
    exit 1
}

# Check arguments
if [ $# -ne 1 ]; then
    usage
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file '$BACKUP_FILE' not found!"
    exit 1
fi

echo "========================================="
echo "Code Together Database Restore"
echo "========================================="
echo "Time: $(date)"
echo "Database: $DB_NAME"
echo "Container: $DB_CONTAINER"
echo "Backup file: $BACKUP_FILE"
echo "========================================="

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${DB_CONTAINER}$"; then
    echo "Error: Database container '$DB_CONTAINER' is not running!"
    exit 1
fi

# Warning
echo ""
echo "⚠️  WARNING: This will restore the database from backup."
echo "⚠️  All current data will be replaced!"
echo ""
read -p "Are you sure you want to continue? (yes/no): " -r
echo
if [[ ! $REPLY =~ ^[Yy]es$ ]]; then
    echo "Restore cancelled."
    exit 0
fi

# Determine if file is compressed
if [[ "$BACKUP_FILE" == *.gz ]]; then
    echo "Detected compressed backup (.gz)"
    RESTORE_CMD="gunzip < \"$BACKUP_FILE\" | docker exec -i \"$DB_CONTAINER\" psql -U \"$DB_USER\" \"$DB_NAME\""
else
    echo "Detected uncompressed backup (.sql)"
    RESTORE_CMD="docker exec -i \"$DB_CONTAINER\" psql -U \"$DB_USER\" \"$DB_NAME\" < \"$BACKUP_FILE\""
fi

# Stop dependent services (optional - uncomment if needed)
# echo "Stopping server and manager services..."
# docker-compose stop server manager

# Perform restore
echo "Starting restore..."
if eval "$RESTORE_CMD" 2>&1 | grep -v "ERROR"; then
    echo "✓ Restore completed successfully"
else
    echo "✗ Restore encountered errors!"
    echo "  Some errors may be expected (e.g., 'already exists')"
    echo "  Please check if the data was restored correctly"
fi

# Restart services (optional - uncomment if you stopped them above)
# echo "Restarting server and manager services..."
# docker-compose start server manager

echo ""
echo "========================================="
echo "Restore process completed!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Check if the application is working correctly"
echo "2. Verify that data has been restored"
echo "3. If needed, restart services:"
echo "   docker-compose restart server manager"
