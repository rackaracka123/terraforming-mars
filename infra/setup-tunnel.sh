#!/bin/bash

# Automated Cloudflare Tunnel Setup using API Key
# This script creates and configures a Cloudflare Tunnel without browser authentication

set -e

echo "🔧 Automated Cloudflare Tunnel Setup"
echo "===================================="
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Check if API key is set
if [ -z "$CLOUDFLARE_API_KEY" ]; then
    echo "❌ Error: CLOUDFLARE_API_KEY not found in .env file"
    echo "Please add your Cloudflare API key to infra/.env"
    exit 1
fi

# Prompt for required information
read -p "Enter your Cloudflare account email: " CLOUDFLARE_EMAIL
read -p "Enter your domain (e.g., example.com): " DOMAIN
read -p "Enter subdomain (leave empty for root domain): " SUBDOMAIN
read -p "Enter tunnel name (default: terraforming-mars): " TUNNEL_NAME

# Set defaults
TUNNEL_NAME=${TUNNEL_NAME:-terraforming-mars}
if [ -z "$SUBDOMAIN" ]; then
    FULL_DOMAIN="$DOMAIN"
else
    FULL_DOMAIN="$SUBDOMAIN.$DOMAIN"
fi

echo ""
echo "📋 Configuration:"
echo "  Email: $CLOUDFLARE_EMAIL"
echo "  Domain: $FULL_DOMAIN"
echo "  Tunnel: $TUNNEL_NAME"
echo ""

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo "📦 Installing cloudflared..."

    ARCH=$(uname -m)
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        echo "Detected ARM64 architecture"
        wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64 -O cloudflared
    elif [ "$ARCH" = "armv7l" ]; then
        echo "Detected ARMv7 architecture"
        wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm -O cloudflared
    else
        echo "Detected x86_64 architecture"
        wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -O cloudflared
    fi

    chmod +x cloudflared
    sudo mv cloudflared /usr/local/bin/
    echo "✅ cloudflared installed"
else
    echo "✅ cloudflared already installed"
fi

echo ""
echo "🔑 Authenticating with Cloudflare API..."

# Create credentials file for cloudflared
mkdir -p ~/.cloudflared
cat > ~/.cloudflared/cert.pem << EOF
# Cloudflare API credentials
# Account Email: $CLOUDFLARE_EMAIL
# API Key: $CLOUDFLARE_API_KEY
EOF

# Alternative: Use environment variables for authentication
export CLOUDFLARE_API_KEY
export CLOUDFLARE_API_EMAIL=$CLOUDFLARE_EMAIL

echo "✅ Authentication configured"

echo ""
echo "🌐 Creating tunnel: $TUNNEL_NAME"

# Create tunnel
TUNNEL_OUTPUT=$(cloudflared tunnel create $TUNNEL_NAME 2>&1 || true)
echo "$TUNNEL_OUTPUT"

# Extract tunnel ID from output or list
TUNNEL_ID=$(cloudflared tunnel list | grep "$TUNNEL_NAME" | head -1 | awk '{print $1}')

if [ -z "$TUNNEL_ID" ]; then
    echo "❌ Failed to create or find tunnel"
    echo "Please check your API credentials and try again"
    exit 1
fi

echo "✅ Tunnel created with ID: $TUNNEL_ID"

echo ""
echo "📝 Configuring DNS..."

# Route DNS to tunnel
cloudflared tunnel route dns $TUNNEL_NAME $FULL_DOMAIN
echo "✅ DNS configured for $FULL_DOMAIN"

# Ask about webhook subdomain
echo ""
read -p "Do you want to set up a webhook subdomain for GitHub auto-deployment? (y/n): " SETUP_WEBHOOK

if [ "$SETUP_WEBHOOK" = "y" ] || [ "$SETUP_WEBHOOK" = "Y" ]; then
    read -p "Enter webhook subdomain (default: webhook): " WEBHOOK_SUB
    WEBHOOK_SUB=${WEBHOOK_SUB:-webhook}
    WEBHOOK_DOMAIN="$WEBHOOK_SUB.$DOMAIN"

    cloudflared tunnel route dns $TUNNEL_NAME $WEBHOOK_DOMAIN
    echo "✅ Webhook DNS configured for $WEBHOOK_DOMAIN"
fi

echo ""
echo "🎫 Generating tunnel token..."

# Get tunnel token
TUNNEL_TOKEN=$(cloudflared tunnel token $TUNNEL_NAME)

echo ""
echo "✅ Setup Complete!"
echo "=================="
echo ""
echo "Add these to your .env file:"
echo ""
echo "TUNNEL_TOKEN=$TUNNEL_TOKEN"
echo "CLOUDFLARE_EMAIL=$CLOUDFLARE_EMAIL"
echo ""
echo "Your application will be accessible at: https://$FULL_DOMAIN"
if [ ! -z "$WEBHOOK_DOMAIN" ]; then
    echo "Webhook endpoint: https://$WEBHOOK_DOMAIN/webhook"
fi
echo ""
echo "Next steps:"
echo "1. Add the TUNNEL_TOKEN to your .env file"
echo "2. Run: docker compose up -d"
echo "3. Visit: https://$FULL_DOMAIN"
echo ""

# Optionally update .env file automatically
read -p "Do you want to automatically update .env file? (y/n): " UPDATE_ENV

if [ "$UPDATE_ENV" = "y" ] || [ "$UPDATE_ENV" = "Y" ]; then
    # Check if TUNNEL_TOKEN already exists
    if grep -q "^TUNNEL_TOKEN=" .env; then
        # Update existing
        sed -i "s|^TUNNEL_TOKEN=.*|TUNNEL_TOKEN=$TUNNEL_TOKEN|" .env
    else
        # Add new
        echo "TUNNEL_TOKEN=$TUNNEL_TOKEN" >> .env
    fi

    # Add email if not exists
    if ! grep -q "^CLOUDFLARE_EMAIL=" .env; then
        echo "CLOUDFLARE_EMAIL=$CLOUDFLARE_EMAIL" >> .env
    fi

    echo "✅ .env file updated!"
fi

echo ""
echo "🎉 Cloudflare Tunnel is ready!"
