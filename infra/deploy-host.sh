#!/bin/bash

# Host-side deployment script for Terraforming Mars
# This script runs directly on the Raspberry Pi host (not in container)
# Triggered by the webhook container via volume mount execution

set -e

# Configuration
# When running in container, repo is mounted at /deploy
# REPO_DIR is the path on the HOST for git operations
CONTAINER_REPO_DIR="/deploy"
HOST_REPO_DIR="${REPO_DIR:-/home/mhm/terraforming-mars}"
LOG_FILE="${LOG_FILE:-/tmp/terraforming-mars-deploy.log}"
BRANCH="${DEPLOY_BRANCH}"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🚀 Starting deployment on host"

# Navigate to repository (use container path)
cd "$CONTAINER_REPO_DIR" || exit 1
log "📂 Repository: $CONTAINER_REPO_DIR (host: $HOST_REPO_DIR)"

# Configure git to trust this directory
git config --global --add safe.directory "$CONTAINER_REPO_DIR"

# Pull latest code
log "🔄 Pulling latest code from origin/$BRANCH..."
git fetch origin
git reset --hard origin/$BRANCH

COMMIT_HASH=$(git rev-parse --short HEAD)
log "✅ Updated to commit: $COMMIT_HASH"

# Navigate to infra directory
cd "$CONTAINER_REPO_DIR/infra" || exit 1

# Rebuild and restart services with zero downtime
log "🏗️  Building and restarting services..."
docker compose up -d --build

# Wait for services to be healthy
log "⏳ Waiting for services to stabilize..."
sleep 5

# Check container status
log "📊 Container status:"
docker compose ps | tee -a "$LOG_FILE"

log "🎉 Deployment complete for commit $COMMIT_HASH"
