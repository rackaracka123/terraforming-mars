#!/bin/bash

# Raspberry Pi Setup Script for Terraforming Mars
# Run this once to set up the deployment environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=== Terraforming Mars Pi Setup ==="
echo ""

# Check for required environment variables
if [ ! -f ".env" ]; then
    echo "ERROR: .env file not found"
    echo "Copy .env.example to .env and fill in your values:"
    echo "  cp .env.example .env"
    echo "  nano .env"
    exit 1
fi

source .env

if [ -z "$GHCR_TOKEN" ]; then
    echo "ERROR: GHCR_TOKEN not set in .env"
    echo ""
    echo "Create a GitHub PAT at: https://github.com/settings/tokens"
    echo "Required scope: read:packages"
    exit 1
fi

if [ -z "$TUNNEL_TOKEN" ]; then
    echo "ERROR: TUNNEL_TOKEN not set in .env"
    exit 1
fi

# Login to ghcr.io
echo "Logging into GitHub Container Registry..."
echo "$GHCR_TOKEN" | docker login ghcr.io -u rackaracka123 --password-stdin

echo ""
echo "Pulling images..."
docker compose -f docker-compose.prod.yml pull

echo ""
echo "Starting services..."
docker compose -f docker-compose.prod.yml up -d

echo ""
echo "Waiting for services to start..."
sleep 5

echo ""
echo "Container status:"
docker compose -f docker-compose.prod.yml ps

echo ""
echo "=== Setup Complete ==="
echo ""
echo "To enable auto-deployment, add this to crontab (crontab -e):"
echo "  */5 * * * * $SCRIPT_DIR/auto-deploy.sh"
echo ""
echo "View logs:"
echo "  tail -f /tmp/terraforming-mars-autodeploy.log"
