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

## Deployment Modes

### Production (Raspberry Pi) - Pull pre-built images

Uses `docker-compose.prod.yml` - pulls pre-built images from ghcr.io.

```bash
cd infra

# First time setup
./pi-setup.sh

# Manual deploy
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

See [CRON_SETUP.md](./CRON_SETUP.md) for automatic deployment setup.

### Local Development - Build from source

Uses `docker-compose.yml` - builds images locally from Dockerfiles.

```bash
cd infra
docker compose up -d --build
```

## Quick Start (Production)

### 1. Initial Setup

```bash
cd /path/to/terraforming-mars/infra

# Create .env file
cp .env.example .env
nano .env  # Add TUNNEL_TOKEN and GHCR_TOKEN
```

### 2. Deploy

```bash
chmod +x pi-setup.sh auto-deploy.sh
./pi-setup.sh
```

### 3. Verify

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f
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

1. **Code Push** → GitHub repository (main branch)
2. **GitHub Actions** → Builds and pushes images to ghcr.io
3. **Cron Job** → Checks for new images every 5 minutes
4. **Pull & Restart** → Downloads new images and restarts containers

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
