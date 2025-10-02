#!/bin/bash

# Host-side deployment script for Terraforming Mars
# This script runs directly on the Raspberry Pi host (not in container)
# Triggered by the webhook container via volume mount execution

set -e

# Configuration - adjust these paths for your Pi
REPO_DIR="${REPO_DIR:-/home/mhm/terraforming-mars}"
LOG_FILE="${LOG_FILE:-/tmp/terraforming-mars-deploy.log}"
BRANCH="${BRANCH}"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🚀 Starting deployment on host"

# Navigate to repository
cd "$REPO_DIR" || exit 1
log "📂 Repository: $REPO_DIR"

# Pull latest code
log "🔄 Pulling latest code from origin/$BRANCH..."
git fetch origin
git reset --hard origin/$BRANCH

COMMIT_HASH=$(git rev-parse --short HEAD)
log "✅ Updated to commit: $COMMIT_HASH"

# Navigate to infra directory
cd "$REPO_DIR/infra" || exit 1

# Stop existing containers
log "🛑 Stopping existing containers"
docker compose down

# Remove old images to force rebuild
log "🗑️  Removing old images"
docker compose down --rmi local || true

# Rebuild images
log "🏗️  Building new Docker images"
docker compose build --no-cache

# Start services
log "▶️  Starting services"
docker compose up -d

# Wait for services to start
log "⏳ Waiting for services to be healthy..."
sleep 10

# Check container status
log "📊 Container status:"
docker compose ps | tee -a "$LOG_FILE"

# Cleanup old Docker resources
log "🧹 Cleaning up old Docker resources"
docker system prune -f

log "🎉 Deployment complete for commit $COMMIT_HASH"
