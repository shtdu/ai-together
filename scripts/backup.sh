#!/bin/bash
# Code Together Database Backup Script
# 数据库备份脚本

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
DB_CONTAINER="${DB_CONTAINER:-code-together-db}"
DB_USER="${DB_USER:-codetogether}"
DB_NAME="${DB_NAME:-code_together}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Timestamp
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/codetogether_$TIMESTAMP.sql.gz"

echo "========================================="
echo "Code Together Database Backup"
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

# Perform backup
echo "Starting backup..."
if docker exec "$DB_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" | gzip > "$BACKUP_FILE"; then
    echo "✓ Backup completed successfully"

    # Get backup size
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    echo "✓ Backup size: $BACKUP_SIZE"
else
    echo "✗ Backup failed!"
    exit 1
fi

# Clean old backups
echo "Cleaning old backups (older than $RETENTION_DAYS days)..."
DELETED=$(find "$BACKUP_DIR" -name "codetogether_*.sql.gz" -mtime +$RETENTION_DAYS -delete -print | wc -l)
if [ "$DELETED" -gt 0 ]; then
    echo "✓ Deleted $DELETED old backup(s)"
else
    echo "✓ No old backups to delete"
fi

# List current backups
echo ""
echo "Current backups:"
ls -lh "$BACKUP_DIR"/codetogether_*.sql.gz 2>/dev/null || echo "No backups found"

echo ""
echo "========================================="
echo "Backup completed successfully!"
echo "========================================="
