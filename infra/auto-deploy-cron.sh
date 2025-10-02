#!/bin/bash

# Auto-deployment cron script for Terraforming Mars
# Checks for new commits and deploys automatically

set -e

# Configuration
REPO_DIR="/home/mhm/terraforming-mars"
BRANCH="infra"
LOG_FILE="/tmp/terraforming-mars-autodeploy.log"
LOCK_FILE="/tmp/terraforming-mars-deploy.lock"

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if deployment is already running
if [ -f "$LOCK_FILE" ]; then
    log "⏭️  Deployment already in progress, skipping..."
    exit 0
fi

# Create lock file
touch "$LOCK_FILE"
trap "rm -f $LOCK_FILE" EXIT

# Navigate to repository
cd "$REPO_DIR" || exit 1

# Get current commit
CURRENT_COMMIT=$(git rev-parse HEAD)

# Fetch latest from origin
git fetch origin "$BRANCH" --quiet

# Get latest commit on remote
REMOTE_COMMIT=$(git rev-parse origin/$BRANCH)

# Check if update is needed
if [ "$CURRENT_COMMIT" = "$REMOTE_COMMIT" ]; then
    log "✅ Already up to date ($(git rev-parse --short HEAD))"
    exit 0
fi

log "🆕 New commit detected: $(git rev-parse --short $REMOTE_COMMIT)"
log "🔄 Pulling latest code..."
git reset --hard origin/$BRANCH

log "🏗️  Building and restarting services..."
cd "$REPO_DIR/infra"
docker compose up -d --build

log "⏳ Waiting for services to stabilize..."
sleep 5

log "📊 Container status:"
docker compose ps | tee -a "$LOG_FILE"

log "🎉 Deployment complete!"