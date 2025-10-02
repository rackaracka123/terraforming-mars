# Raspberry Pi Deployment - Quick Start Guide

Deploy Terraforming Mars on your Raspberry Pi with Docker Compose and automatic GitHub deployments.

## Prerequisites

- Raspberry Pi (3/4/5 or similar ARM device)
- Docker and Docker Compose installed
- Cloudflare account
- Domain managed by Cloudflare
- Git repository on GitHub

## Quick Setup (5 Steps)

### 1. Install Docker (if not already installed)

```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
newgrp docker
```

### 2. Clone Repository

```bash
git clone https://github.com/your-username/terraforming-mars.git
cd terraforming-mars
```

### 3. Set Up Cloudflare Tunnel

```bash
./infra/cloudflare-tunnel-setup.sh
```

Follow the prompts to:
- Authenticate with Cloudflare
- Create a tunnel
- Configure DNS for your domain
- Get your tunnel token

### 4. Configure Environment

```bash
cp infra/.env.example .env
nano .env
```

Add your secrets:
```env
TM_LOG_LEVEL=info
TUNNEL_TOKEN=your_tunnel_token_from_step_3
WEBHOOK_SECRET=$(openssl rand -hex 32)
```

### 5. Deploy

```bash
cd infra
docker compose build
docker compose up -d
```

## Verify Deployment

```bash
# Check all services are running
docker compose ps

# View logs
docker compose logs -f

# Test your domain
curl https://yourdomain.com
```

Your application is now live at `https://yourdomain.com`! 🎉

## Set Up Auto-Deployment (Optional)

To automatically deploy when you push to GitHub:

### 1. Configure GitHub Webhook

1. Go to: GitHub Repository → Settings → Webhooks → Add webhook
2. **Payload URL**: `https://webhook.yourdomain.com/webhook`
3. **Content type**: `application/json`
4. **Secret**: Use the `WEBHOOK_SECRET` from your `.env` file
5. **Events**: Select "Just the push event"
6. Save the webhook

### 2. Configure Cloudflare Tunnel for Webhooks

Create `infra/cloudflared-config.yml`:

```yaml
tunnel: your-tunnel-id
credentials-file: /root/.cloudflared/your-tunnel-id.json

ingress:
  - hostname: webhook.yourdomain.com
    service: http://webhook:9000
  - hostname: yourdomain.com
    service: http://frontend:8080
  - service: http_status:404
```

Add DNS route:
```bash
cloudflared tunnel route dns your-tunnel-name webhook.yourdomain.com
```

Update docker-compose.yml cloudflared service to use the config file.

### 3. Test Auto-Deployment

```bash
echo "# Test" >> README.md
git add README.md
git commit -m "Test auto-deployment"
git push origin main
```

Watch deployment:
```bash
docker compose logs -f webhook
```

## Common Commands

```bash
# View logs
docker compose logs -f [service]

# Restart service
docker compose restart [service]

# Stop all services
docker compose down

# Rebuild and restart
docker compose down && docker compose build && docker compose up -d

# Check health
docker compose ps
```

## Architecture

```
GitHub Push
    ↓
Webhook → Cloudflare Tunnel → Docker Containers
                                     ↓
                           Backend + Frontend + Webhook
```

## File Structure

```
terraforming-mars/
├── backend/              # Go server source
├── frontend/             # React app source
├── infra/                # All infrastructure files
│   ├── docker-compose.yml
│   ├── deploy.sh
│   ├── webhook-server.js
│   ├── Dockerfile.webhook
│   ├── cloudflare-tunnel-setup.sh
│   └── .env.example
├── DEPLOYMENT.md         # Detailed deployment guide
├── WEBHOOK_SETUP.md      # Webhook configuration guide
└── .env                  # Your secrets (not in git)
```

## Troubleshooting

### Services Won't Start

```bash
docker compose logs [service_name]
docker compose ps
```

### Webhook Not Working

```bash
# Check webhook logs
docker compose logs webhook

# Test webhook endpoint
curl -X POST https://webhook.yourdomain.com/webhook
```

### Out of Disk Space

```bash
# Clean up Docker
docker system prune -a -f
docker volume prune -f

# Check disk usage
df -h
```

### Deployment Fails

```bash
# Manual deployment
cd infra
docker compose exec webhook /deploy/infra/deploy.sh

# Check deployment logs
docker compose exec webhook cat /var/log/terraforming-mars-deploy.log
```

## Need More Help?

- **Full Deployment Guide**: See [DEPLOYMENT.md](./DEPLOYMENT.md)
- **Webhook Setup**: See [WEBHOOK_SETUP.md](./WEBHOOK_SETUP.md)
- **Infrastructure Details**: See [infra/README.md](./infra/README.md)

## Support

For issues or questions:
1. Check service logs: `docker compose logs [service]`
2. Verify health: `docker compose ps`
3. Review the detailed documentation
4. Check GitHub webhook delivery status

---

**Next Steps**: After successful deployment, configure the webhook for automatic updates whenever you push to GitHub!
