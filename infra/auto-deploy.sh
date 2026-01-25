#!/bin/bash

# Auto-deployment script for Terraforming Mars (Production)
# Checks for new Docker images and deploys automatically

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_FILE="/tmp/terraforming-mars-autodeploy.log"
LOCK_FILE="/tmp/terraforming-mars-deploy.lock"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.prod.yml"

# Images to check
BACKEND_IMAGE="ghcr.io/rackaracka123/terraforming-mars-backend:latest"
FRONTEND_IMAGE="ghcr.io/rackaracka123/terraforming-mars-frontend:latest"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if deployment is already running
if [ -f "$LOCK_FILE" ]; then
    log "Deployment already in progress, skipping..."
    exit 0
fi

# Create lock file
touch "$LOCK_FILE"
trap "rm -f $LOCK_FILE" EXIT

# Get current local image digest
get_local_digest() {
    docker inspect --format='{{index .RepoDigests 0}}' "$1" 2>/dev/null | cut -d@ -f2 || echo "none"
}

# Get remote image digest
get_remote_digest() {
    docker manifest inspect "$1" 2>/dev/null | grep -m1 '"digest"' | cut -d'"' -f4 || echo "error"
}

log "Checking for image updates..."

# Get current local digests
LOCAL_BACKEND=$(get_local_digest "$BACKEND_IMAGE")
LOCAL_FRONTEND=$(get_local_digest "$FRONTEND_IMAGE")

# Get remote digests
REMOTE_BACKEND=$(get_remote_digest "$BACKEND_IMAGE")
REMOTE_FRONTEND=$(get_remote_digest "$FRONTEND_IMAGE")

# Check if updates are available
BACKEND_UPDATED=false
FRONTEND_UPDATED=false

if [ "$REMOTE_BACKEND" != "error" ] && [ "$LOCAL_BACKEND" != "$REMOTE_BACKEND" ]; then
    BACKEND_UPDATED=true
    log "New backend image available"
fi

if [ "$REMOTE_FRONTEND" != "error" ] && [ "$LOCAL_FRONTEND" != "$REMOTE_FRONTEND" ]; then
    FRONTEND_UPDATED=true
    log "New frontend image available"
fi

# Exit if no updates
if [ "$BACKEND_UPDATED" = false ] && [ "$FRONTEND_UPDATED" = false ]; then
    log "All images up to date"
    exit 0
fi

log "Pulling new images..."
docker compose -f "$COMPOSE_FILE" pull

log "Restarting services with new images..."
docker compose -f "$COMPOSE_FILE" up -d

log "Waiting for services to stabilize..."
sleep 5

log "Container status:"
docker compose -f "$COMPOSE_FILE" ps | tee -a "$LOG_FILE"

log "Cleaning up old images..."
docker image prune -f

log "Deployment complete!"
