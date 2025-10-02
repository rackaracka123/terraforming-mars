#!/bin/bash

# Automated deployment script for Terraforming Mars
# Triggered by GitHub webhook on push/merge to main branch

set -e

# Configuration
REPO_DIR="/deploy"
LOG_FILE="/var/log/terraforming-mars-deploy.log"
BRANCH="infra"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "🚀 Starting automated deployment"

# Navigate to repository
cd "$REPO_DIR" || exit 1
log "📂 Changed to repository directory: $REPO_DIR"

# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
log "📍 Current branch: $CURRENT_BRANCH"

# Stash any local changes (if any)
if ! git diff-index --quiet HEAD --; then
    log "💾 Stashing local changes"
    git stash
fi

# Fetch latest changes
log "📥 Fetching latest changes from remote"
git fetch origin

# Checkout main branch if not already on it
if [ "$CURRENT_BRANCH" != "$BRANCH" ]; then
    log "🔀 Switching to $BRANCH branch"
    git checkout "$BRANCH"
fi

# Pull latest changes
log "⬇️  Pulling latest changes"
git pull origin "$BRANCH"

# Get the latest commit hash
COMMIT_HASH=$(git rev-parse --short HEAD)
log "📝 Latest commit: $COMMIT_HASH"

# Navigate to infra directory for docker-compose
cd "$REPO_DIR/infra" || exit 1
log "📂 Changed to infra directory"

# Stop existing containers
log "🛑 Stopping existing containers"
docker compose down

# Remove old images to force rebuild
log "🗑️  Removing old images"
docker compose down --rmi local

# Rebuild images
log "🏗️  Building new Docker images"
docker compose build --no-cache

# Start services
log "▶️  Starting services"
docker compose up -d

# Wait for services to be healthy
log "⏳ Waiting for services to be healthy..."
sleep 10

# Check health status
if docker compose ps | grep -q "unhealthy"; then
    log "❌ Deployment failed - some services are unhealthy"
    docker compose ps
    docker compose logs
    exit 1
fi

log "✅ Deployment successful!"
log "📊 Container status:"
docker compose ps | tee -a "$LOG_FILE"

# Cleanup old Docker resources
log "🧹 Cleaning up old Docker resources"
docker system prune -f

log "🎉 Deployment complete for commit $COMMIT_HASH"
