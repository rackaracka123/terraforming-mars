#!/bin/bash

# Cloudflare Tunnel Setup Script for Terraforming Mars
# This script helps you create and configure a Cloudflare Tunnel

set -e

echo "🔧 Cloudflare Tunnel Setup for Terraforming Mars"
echo "=================================================="
echo ""

# Check if cloudflared is installed
if ! command -v cloudflared &> /dev/null; then
    echo "📦 Installing cloudflared..."

    # Detect architecture
    ARCH=$(uname -m)
    if [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
        echo "Detected ARM64 architecture (Raspberry Pi)"
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
echo "🔐 Cloudflare Authentication"
echo "----------------------------"
echo "Please authenticate with Cloudflare..."
cloudflared tunnel login

echo ""
echo "🌐 Tunnel Creation"
echo "------------------"
read -p "Enter a name for your tunnel (e.g., terraforming-mars): " TUNNEL_NAME

if [ -z "$TUNNEL_NAME" ]; then
    TUNNEL_NAME="terraforming-mars"
fi

echo "Creating tunnel: $TUNNEL_NAME"
cloudflared tunnel create $TUNNEL_NAME

echo ""
echo "📋 Getting Tunnel Token"
echo "-----------------------"
TUNNEL_ID=$(cloudflared tunnel list | grep $TUNNEL_NAME | awk '{print $1}')
echo "Tunnel ID: $TUNNEL_ID"

echo ""
echo "🔗 DNS Configuration"
echo "--------------------"
read -p "Enter your domain (e.g., example.com): " DOMAIN

if [ -z "$DOMAIN" ]; then
    echo "❌ Domain is required"
    exit 1
fi

read -p "Enter subdomain (leave empty for root domain): " SUBDOMAIN

if [ -z "$SUBDOMAIN" ]; then
    FULL_DOMAIN="$DOMAIN"
else
    FULL_DOMAIN="$SUBDOMAIN.$DOMAIN"
fi

echo "Configuring DNS for: $FULL_DOMAIN"
cloudflared tunnel route dns $TUNNEL_NAME $FULL_DOMAIN

echo ""
echo "🎫 Generating Tunnel Token"
echo "--------------------------"
TUNNEL_TOKEN=$(cloudflared tunnel token $TUNNEL_NAME)

echo ""
echo "✅ Setup Complete!"
echo "=================="
echo ""
echo "Add this to your .env file:"
echo ""
echo "TUNNEL_TOKEN=$TUNNEL_TOKEN"
echo ""
echo "Your application will be accessible at: https://$FULL_DOMAIN"
echo ""
echo "Next steps:"
echo "1. Copy the TUNNEL_TOKEN to your .env file"
echo "2. Run: docker-compose up -d"
echo "3. Visit: https://$FULL_DOMAIN"
