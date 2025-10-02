# Terraforming Mars - Docker Deployment Guide

This guide will help you deploy Terraforming Mars on your Raspberry Pi using Docker Compose with Cloudflare Tunnel for secure domain access.

## Prerequisites

- Raspberry Pi (3/4/5 or any ARM64 device) running Linux
- Docker and Docker Compose installed
- A Cloudflare account
- A domain managed by Cloudflare

## Architecture

The deployment consists of three Docker containers:

1. **Backend** - Go server (port 3001)
2. **Frontend** - React app served by Nginx (port 8080)
3. **Cloudflared** - Cloudflare Tunnel for secure HTTPS access

```
Internet → Cloudflare Tunnel → Frontend (Nginx) → Backend (Go)
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

Run the automated setup script:

```bash
./cloudflare-tunnel-setup.sh
```

This script will:
1. Install `cloudflared` CLI (if not already installed)
2. Authenticate with your Cloudflare account
3. Create a new tunnel
4. Configure DNS for your domain
5. Generate a tunnel token

Follow the prompts:
- Choose a tunnel name (e.g., `terraforming-mars`)
- Enter your domain (e.g., `example.com`)
- Enter a subdomain (e.g., `tm` for `tm.example.com`) or leave empty for root domain

### 4. Configure Environment Variables

Create a `.env` file from the example:

```bash
cp .env.example .env
```

Edit `.env` and add your tunnel token:

```bash
nano .env
```

```env
TM_LOG_LEVEL=info
TUNNEL_TOKEN=your_actual_tunnel_token_here
```

### 5. Build and Deploy

Build the Docker images:

```bash
docker compose build
```

Start the services:

```bash
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
```

### 6. Verify Deployment

1. **Local Access** (optional, if you exposed port 8080):
   ```bash
   curl http://localhost:8080/health
   ```

2. **Domain Access**:
   Visit `https://your-domain.com` in your browser

3. **Check Container Health**:
   ```bash
   docker compose ps
   ```
   All containers should show "healthy" status.

## Management Commands

### Start Services
```bash
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
```

### Update Deployment

After making code changes:

```bash
# Rebuild and restart
docker compose down
docker compose build
docker compose up -d
```

### Remove Everything
```bash
docker compose down --volumes --rmi all
```

## Troubleshooting

### Check Container Health

```bash
docker compose ps
```

Expected output shows all services as "healthy".

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

To change your domain or recreate the tunnel:

1. Stop services:
   ```bash
   docker compose down
   ```

2. Delete existing tunnel:
   ```bash
   cloudflared tunnel delete terraforming-mars
   ```

3. Re-run setup:
   ```bash
   ./cloudflare-tunnel-setup.sh
   ```

4. Update `.env` with new token

5. Restart services:
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

To serve the app on multiple domains:

```bash
# Add additional DNS routes
cloudflared tunnel route dns terraforming-mars second-domain.com
```

### Local Development Alongside Docker

The Docker setup is separate from local development:
- Local dev: `make run` (ports 3000/3001)
- Docker: `docker compose up -d` (internal networking + Cloudflare Tunnel)

Both can run simultaneously without conflict.
