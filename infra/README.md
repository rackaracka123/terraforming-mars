# Infrastructure Directory

This directory contains all the infrastructure configuration needed to deploy Terraforming Mars on a Raspberry Pi with Docker Compose and Cloudflare Tunnel.

## Contents

### Core Files

- **docker-compose.yml** - Main Docker Compose configuration for all services
- **Dockerfile.webhook** - Dockerfile for the GitHub webhook server
- **.env.example** - Environment variables template

### Deployment & Automation

- **deploy.sh** - Automated deployment script triggered by webhooks
- **webhook-server.js** - Node.js server that receives GitHub webhooks
- **cloudflare-tunnel-setup.sh** - Interactive script to configure Cloudflare Tunnel

### Documentation

- **../DEPLOYMENT.md** - Complete deployment guide for Raspberry Pi
- **../WEBHOOK_SETUP.md** - GitHub webhook setup instructions

## Quick Start

### 1. Initial Setup

```bash
# Run from project root
cd /path/to/terraforming-mars

# Set up Cloudflare Tunnel
./infra/cloudflare-tunnel-setup.sh

# Create .env file
cp infra/.env.example .env
nano .env  # Add your TUNNEL_TOKEN and WEBHOOK_SECRET
```

### 2. Deploy

```bash
cd infra
docker compose build
docker compose up -d
```

### 3. Verify

```bash
docker compose ps
docker compose logs -f
```

## Services

The Docker Compose configuration includes:

### 1. Backend (Go Server)
- Port: 3001 (internal)
- Health check enabled
- WebSocket support

### 2. Frontend (React + Nginx)
- Port: 8080 (internal)
- Serves static files
- Reverse proxy to backend
- WebSocket proxy

### 3. Cloudflare Tunnel
- Provides secure HTTPS access
- No exposed ports
- Automatic SSL certificates

### 4. Webhook Server
- Port: 9000 (internal)
- Receives GitHub push events
- Triggers automatic deployments
- Has access to Docker socket

## Architecture

```
Internet
  ↓
Cloudflare CDN
  ↓
Cloudflare Tunnel (container)
  ↓
  ├─→ Frontend (Nginx) → Backend (Go)
  └─→ Webhook Server → Deploy Script → Docker Rebuild
```

## Environment Variables

Create a `.env` file in the project root:

```env
TM_LOG_LEVEL=info
TUNNEL_TOKEN=your_cloudflare_tunnel_token
WEBHOOK_SECRET=your_github_webhook_secret
```

## Management

### View Logs
```bash
docker compose logs -f [service_name]
```

### Restart Services
```bash
docker compose restart [service_name]
```

### Stop Everything
```bash
docker compose down
```

### Rebuild
```bash
docker compose build
docker compose up -d
```

## File Relationships

```
infra/
├── docker-compose.yml          → References ../backend and ../frontend
├── Dockerfile.webhook          → Builds webhook server image
├── webhook-server.js           → Runs inside webhook container
├── deploy.sh                   → Executed by webhook container
├── cloudflare-tunnel-setup.sh  → Run manually during setup
└── .env.example                → Template for .env in project root
```

## Deployment Flow

1. **Code Push** → GitHub repository
2. **Webhook** → GitHub sends POST to webhook.yourdomain.com
3. **Webhook Container** → Receives and verifies webhook
4. **Deploy Script** → Pulls latest code
5. **Docker Rebuild** → Builds new images
6. **Container Restart** → Deploys updated application

## Troubleshooting

### Container Won't Start

```bash
docker compose ps
docker compose logs [service_name]
```

### Rebuild from Scratch

```bash
docker compose down --rmi all
docker compose build --no-cache
docker compose up -d
```

### Check Health

```bash
docker compose ps  # Shows health status
docker compose exec frontend wget -O- http://backend:3001/api/health
```

## Security Notes

- All secrets in `.env` file (not committed to git)
- No ports exposed externally (only through Cloudflare Tunnel)
- Webhook signatures verified before deployment
- Containers run as non-root users where possible
- Docker socket access limited to webhook container

## Additional Resources

- [Deployment Guide](../DEPLOYMENT.md) - Complete setup instructions
- [Webhook Setup](../WEBHOOK_SETUP.md) - GitHub integration guide
- [Cloudflare Tunnels Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
