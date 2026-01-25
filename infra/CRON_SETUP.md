# Auto-Deploy Setup for Raspberry Pi

Automatic deployment using pre-built Docker images from GitHub Container Registry.

## Prerequisites

- Docker and Docker Compose installed
- GitHub Personal Access Token (PAT) with `read:packages` scope

## Initial Setup

### 1. Create GitHub PAT

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Name: "Pi ghcr.io read access"
4. Select scope: `read:packages`
5. Generate and copy the token

### 2. Configure Environment

```bash
cd /path/to/terraforming-mars/infra
cp .env.example .env
nano .env
```

Fill in:
- `TUNNEL_TOKEN` - Your Cloudflare Tunnel token
- `GHCR_TOKEN` - Your GitHub PAT from step 1

### 3. Run Setup Script

```bash
chmod +x pi-setup.sh auto-deploy.sh
./pi-setup.sh
```

This will:
- Login to ghcr.io
- Pull the latest images
- Start all services

### 4. Enable Auto-Deploy

Add to crontab to check for updates every 5 minutes:

```bash
crontab -e
```

Add this line:
```
*/5 * * * * /path/to/terraforming-mars/infra/auto-deploy.sh
```

## How It Works

1. **GitHub Actions** builds multi-arch images on push to main
2. **Images** pushed to ghcr.io/rackaracka123/terraforming-mars-{backend,frontend}
3. **Cron job** checks for new image digests every 5 minutes
4. **If updated**, pulls new images and restarts containers

## Commands

### Manual Deploy
```bash
./auto-deploy.sh
```

### View Logs
```bash
tail -f /tmp/terraforming-mars-autodeploy.log
```

### Check Container Status
```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f
```

### Force Re-pull
```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

### Stop Everything
```bash
docker compose -f docker-compose.prod.yml down
```

## Troubleshooting

### Authentication Failed
```bash
source .env
echo "$GHCR_TOKEN" | docker login ghcr.io -u rackaracka123 --password-stdin
```

### Images Not Found
```bash
docker manifest inspect ghcr.io/rackaracka123/terraforming-mars-backend:latest
docker manifest inspect ghcr.io/rackaracka123/terraforming-mars-frontend:latest
```

### Service Won't Start
```bash
docker compose -f docker-compose.prod.yml logs backend
docker compose -f docker-compose.prod.yml logs frontend
```
