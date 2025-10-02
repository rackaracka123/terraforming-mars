# Terraforming Mars - Docker Deployment Guide

This guide will help you deploy Terraforming Mars on your Raspberry Pi using Docker Compose with Cloudflare Tunnel for secure domain access.

## Prerequisites

- Raspberry Pi (3/4/5 or any ARM64 device) running Linux
- Docker and Docker Compose installed
- A Cloudflare account
- A domain managed by Cloudflare

## Architecture

The deployment consists of four Docker containers:

1. **Backend** - Go server (port 3001, internal)
2. **Frontend** - React app served by Nginx (port 8080, internal)
3. **Cloudflared** - Cloudflare Tunnel for secure HTTPS access
4. **Webhook** - GitHub webhook server for auto-deployment (port 9000, internal)

```
Internet → Cloudflare Tunnel → Frontend (Nginx) → Backend (Go)
                              → Webhook Server → Deploy Script
```

## Installation Steps

### 1. Install Docker and Docker Compose

If not already installed on your Raspberry Pi:

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Install Docker Compose
sudo apt-get update
sudo apt-get install docker-compose-plugin
```

Verify installation:
```bash
docker --version
docker compose version
```

### 2. Clone or Navigate to the Repository

```bash
cd /path/to/terraforming-mars
```

### 3. Set Up Cloudflare Tunnel

You need to manually create a Cloudflare Tunnel in your Cloudflare dashboard:

1. Go to Cloudflare Zero Trust dashboard
2. Navigate to **Access** → **Tunnels**
3. Create a new tunnel (e.g., `terraforming-mars`)
4. Note the tunnel token
5. Configure tunnel routes:
   - Main app: `yourdomain.com` → `http://frontend:8080`
   - Webhook: `webhook.yourdomain.com` → `http://webhook:9000`

### 4. Configure Environment Variables

Navigate to the infra directory and configure your environment:

```bash
cd infra
cp .env.example .env
nano .env
```

Update the `.env` file with your values:

```env
# Cloudflare tunnel token from step 3
TUNNEL_TOKEN=your_actual_tunnel_token_here

# Backend log level
TM_LOG_LEVEL=info

# GitHub webhook secret (generate with: openssl rand -hex 32)
WEBHOOK_SECRET=your_generated_secret_here
```

### 5. Build and Deploy

From the `infra` directory, build and start all services:

```bash
cd infra
docker compose build
docker compose up -d
```

Check the status:

```bash
docker compose ps
```

View logs:

```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f cloudflared
docker compose logs -f webhook
```

### 6. Verify Deployment

1. **Check all containers are running**:
   ```bash
   docker compose ps
   ```

2. **Domain Access**:
   Visit `https://your-domain.com` in your browser

3. **Test backend health**:
   ```bash
   docker compose exec frontend wget -O- http://backend:3001/api/health
   ```

### 7. Set Up GitHub Webhook (Optional - for auto-deployment)

1. Go to your GitHub repository → Settings → Webhooks
2. Add webhook:
   - **Payload URL**: `https://webhook.yourdomain.com/webhook`
   - **Content type**: `application/json`
   - **Secret**: Use `WEBHOOK_SECRET` from your `.env` file
   - **Events**: Just the push event
3. Save webhook
4. GitHub will send a ping - check webhook logs:
   ```bash
   docker compose logs webhook
   ```

## Management Commands

**Note**: All commands should be run from the `infra/` directory.

### Start Services
```bash
cd infra
docker compose up -d
```

### Stop Services
```bash
docker compose down
```

### Restart Services
```bash
docker compose restart
```

### View Logs
```bash
# All services
docker compose logs -f

# Specific service
docker compose logs -f backend
docker compose logs -f webhook
```

### Update Deployment

#### Manual Update
After making code changes:

```bash
cd infra
docker compose down
docker compose build
docker compose up -d
```

#### Automatic Update (via webhook)
If webhook is configured, simply push to main:

```bash
git add .
git commit -m "Your changes"
git push origin main
```

The webhook will automatically:
1. Pull latest code
2. Rebuild containers
3. Restart services

Watch the deployment:
```bash
docker compose logs -f webhook
```

### Remove Everything
```bash
docker compose down --volumes --rmi all
```

## Troubleshooting

**Note**: Run all commands from the `infra/` directory.

### Check Container Status

```bash
docker compose ps
```

All containers should show "Up" status.

### Backend Issues

```bash
# Check backend logs
docker compose logs backend

# Test backend health endpoint
docker compose exec frontend wget -O- http://backend:3001/api/health
```

### Frontend Issues

```bash
# Check frontend logs
docker compose logs frontend

# Check Nginx configuration
docker compose exec frontend cat /etc/nginx/conf.d/default.conf
```

### Cloudflare Tunnel Issues

```bash
# Check tunnel logs
docker compose logs cloudflared

# Verify tunnel status
docker compose exec cloudflared cloudflared tunnel info
```

### WebSocket Connection Issues

WebSocket connections are proxied through Nginx to the backend. If game state updates aren't working:

1. Check backend logs for WebSocket connections
2. Check browser console for connection errors
3. Verify Nginx is properly proxying WebSocket connections

```bash
docker compose logs backend | grep -i websocket
```

### Domain Not Accessible

1. Verify Cloudflare Tunnel is running:
   ```bash
   docker compose logs cloudflared
   ```

2. Check DNS configuration in Cloudflare dashboard:
   - Go to your domain's DNS settings
   - Verify CNAME record exists for your subdomain
   - Ensure proxy is enabled (orange cloud)

3. Wait a few minutes for DNS propagation

### Rebuild from Scratch

If you encounter build issues:

```bash
# Remove all containers and images
docker compose down --rmi all

# Clean Docker cache
docker system prune -a

# Rebuild
docker compose build --no-cache
docker compose up -d
```

## Performance Optimization

### For Raspberry Pi 3/4

If running on older Raspberry Pi models, you may want to:

1. **Reduce Logging**:
   ```env
   TM_LOG_LEVEL=warn
   ```

2. **Monitor Resources**:
   ```bash
   docker stats
   ```

3. **Limit Container Resources** (add to docker-compose.yml):
   ```yaml
   services:
     backend:
       deploy:
         resources:
           limits:
             memory: 512M
   ```

## Security Considerations

1. **Cloudflare Tunnel** provides automatic HTTPS and DDoS protection
2. **No exposed ports** - All traffic goes through Cloudflare
3. **Non-root containers** - Both frontend and backend run as non-root users
4. **Security headers** - Nginx adds security headers to all responses

## Architecture Details

### Network Flow

```
Internet
  ↓
Cloudflare CDN & Security
  ↓
Cloudflare Tunnel (cloudflared container)
  ↓
Frontend (Nginx on port 8080)
  ↓
  ├─ Static files (React build)
  ├─ /api/* → Backend (port 3001)
  └─ /socket.io/* → Backend WebSocket
```

### Container Communication

All containers communicate via the `tm-network` Docker bridge network:
- Frontend can reach backend at `http://backend:3001`
- Cloudflared can reach frontend at `http://frontend:8080`
- External access only through Cloudflare Tunnel (no exposed ports)

## Updating Cloudflare Tunnel

To change your domain or update tunnel configuration:

1. Stop services:
   ```bash
   cd infra
   docker compose down
   ```

2. Update tunnel in Cloudflare dashboard:
   - Modify routes/domains as needed
   - Get new tunnel token if recreating

3. Update `.env` with new token (if changed)

4. Restart services:
   ```bash
   docker compose up -d
   ```

## Backup and Restore

Since the game uses in-memory storage, there's no database to back up. If you add persistent storage later:

```bash
# Backup volumes
docker compose down
tar -czf backup.tar.gz -C /var/lib/docker/volumes .

# Restore
tar -xzf backup.tar.gz -C /var/lib/docker/volumes
docker compose up -d
```

## Support

For issues or questions:
1. Check container logs: `docker compose logs`
2. Verify health status: `docker compose ps`
3. Review this deployment guide
4. Check the main repository documentation

## Advanced Configuration

### Custom Domain SSL

Cloudflare Tunnel automatically handles SSL certificates. No additional configuration needed!

### Multiple Domains

To serve the app on multiple domains, add additional routes in your Cloudflare Tunnel dashboard configuration.

### Local Development Alongside Docker

The Docker setup is separate from local development:
- Local dev: `make run` (ports 3000/3001)
- Docker: `docker compose up -d` (internal networking + Cloudflare Tunnel)

Both can run simultaneously without conflict.
