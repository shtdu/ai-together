#!/bin/bash

set -e

REMOTE_HOST="root@192.168.2.129"
REMOTE_DIR="/opt/code-together"
BINARY_NAME="switch-server-linux-amd64"
LOCAL_BINARY="server/bin/${BINARY_NAME}"

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --sys          Build, deploy binary and migrations [default]"
    echo "  --no-build     Deploy without rebuilding (use existing binary)"
    echo "  --migrations   Only sync migrations (no binary)"
    echo "  -h             Show this help message"
    exit 0
}

build() {
    echo "Building Linux binary..."
    make server-linux
    echo "Build completed."
}

sync_migrations() {
    echo "Syncing migrations..."
    rsync -av --delete server/migrations/ "${REMOTE_HOST}:${REMOTE_DIR}/migrations/"
}

sync_binary() {
    if [ ! -f "${LOCAL_BINARY}" ]; then
        echo "Error: Binary not found at ${LOCAL_BINARY}"
        exit 1
    fi

    # Check if binary changed by file size
    LOCAL_SIZE=$(stat -f%z "${LOCAL_BINARY}")
    REMOTE_SIZE=$(ssh "${REMOTE_HOST}" "stat -c%s ${REMOTE_DIR}/bin/${BINARY_NAME} 2>/dev/null" || echo "0")

    if [ "${LOCAL_SIZE}" = "${REMOTE_SIZE}" ]; then
        echo "Binary unchanged (size: ${LOCAL_SIZE}), skipping upload."
        return 0
    fi

    echo "Binary changed (local: ${LOCAL_SIZE}, remote: ${REMOTE_SIZE}), uploading..."
    scp "${LOCAL_BINARY}" "${REMOTE_HOST}:/tmp/${BINARY_NAME}.tmp"

    echo "Moving binary to final location..."
    ssh "${REMOTE_HOST}" "
        mv /tmp/${BINARY_NAME}.tmp ${REMOTE_DIR}/bin/${BINARY_NAME} && \
        chmod +x ${REMOTE_DIR}/bin/${BINARY_NAME}
    "
}

restart_server() {
    echo "Restarting server container..."
    ssh "${REMOTE_HOST}" "cd ${REMOTE_DIR} && docker compose restart server"
}

deploy_full() {
    build
    echo "Deploying to ${REMOTE_HOST}..."
    sync_binary
    sync_migrations
    restart_server
    echo "Deploy completed successfully!"
}

deploy_no_build() {
    echo "Deploying to ${REMOTE_HOST} (skipping build)..."
    sync_binary
    sync_migrations
    restart_server
    echo "Deploy completed successfully!"
}

deploy_migrations_only() {
    echo "Syncing migrations to ${REMOTE_HOST}..."
    sync_migrations
    restart_server
    echo "Migrations synced successfully!"
}

# Parse arguments
case "${1:-}" in
    ""|--sys)
        deploy_full
        ;;
    --no-build)
        deploy_no_build
        ;;
    --migrations)
        deploy_migrations_only
        ;;
    -h|--help)
        usage
        ;;
    *)
        echo "Unknown option: $1"
        usage
        ;;
esac
