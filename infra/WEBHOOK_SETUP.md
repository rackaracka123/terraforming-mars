# GitHub Webhook Auto-Deployment Setup

This guide explains how to set up automatic deployment when you push or merge to the main branch using GitHub webhooks with a Docker-based webhook server.

## Overview

When you push to the `main` branch on GitHub:
1. GitHub sends a webhook POST request to your Raspberry Pi
2. Webhook Docker container verifies the request signature
3. Deployment script pulls latest code and rebuilds Docker containers
4. Your application is automatically updated

## Prerequisites

- Raspberry Pi with Docker and Docker Compose installed
- Cloudflare Tunnel configured
- Your domain's DNS managed by Cloudflare

## Architecture

```
GitHub Push → Webhook POST → Cloudflare Tunnel → Webhook Container → Deploy Script → Docker Rebuild
```

The webhook server runs as a Docker container alongside your application, with access to the Docker socket for rebuilding containers.

## Setup Steps

### 1. Generate a Webhook Secret

Generate a secure random secret for webhook authentication:

```bash
openssl rand -hex 32
```

Save this secret - you'll need it for both your `.env` file and GitHub configuration.

### 2. Configure Environment Variables

Edit your `.env` file and add the webhook secret:

```bash
nano .env
```

Add this line:
```env
WEBHOOK_SECRET=your_actual_secret_from_step_1
```

Your `.env` should now have:
```env
TM_LOG_LEVEL=info
TUNNEL_TOKEN=your_tunnel_token
WEBHOOK_SECRET=your_webhook_secret
```

### 3. Configure Cloudflare Tunnel for Webhooks

You need to expose the webhook endpoint so GitHub can reach it. You have two options:

#### Option A: Use Same Domain with Path Routing (Simpler)

Update your `docker-compose.yml` cloudflared service to use a config file:

```yaml
cloudflared:
  image: cloudflare/cloudflared:latest
  container_name: tm-cloudflared
  restart: unless-stopped
  command: tunnel run
  volumes:
    - ./cloudflared-config.yml:/etc/cloudflared/config.yml
  depends_on:
    frontend:
      condition: service_healthy
  networks:
    - tm-network
  extra_hosts:
    - "host.docker.internal:host-gateway"
```

Create `cloudflared-config.yml`:

```yaml
tunnel: your-tunnel-id
credentials-file: /root/.cloudflared/your-tunnel-id.json

ingress:
  # Main application
  - hostname: yourdomain.com
    path: ^/webhook
    service: http://webhook:9000
  - hostname: yourdomain.com
    service: http://frontend:8080
  # Catch-all
  - service: http_status:404
```

Your webhook URL will be: `https://yourdomain.com/webhook`

#### Option B: Use Subdomain (Recommended)

Create `cloudflared-config.yml`:

```yaml
tunnel: your-tunnel-id
credentials-file: /root/.cloudflared/your-tunnel-id.json

ingress:
  # Webhook endpoint on subdomain
  - hostname: webhook.yourdomain.com
    service: http://webhook:9000
  # Main application
  - hostname: yourdomain.com
    service: http://frontend:8080
  # Catch-all
  - service: http_status:404
```

Then add DNS route for the webhook subdomain:

```bash
cloudflared tunnel route dns your-tunnel-name webhook.yourdomain.com
```

Your webhook URL will be: `https://webhook.yourdomain.com/webhook`

### 4. Start the Webhook Server

Rebuild and restart all services:

```bash
docker compose down
docker compose build
docker compose up -d
```

Verify all containers are running:

```bash
docker compose ps
```

You should see:
- `tm-backend` (healthy)
- `tm-frontend` (healthy)
- `tm-cloudflared` (running)
- `tm-webhook` (running)

Check webhook server logs:

```bash
docker compose logs webhook
```

You should see:
```
🎧 Webhook server listening on port 9000
```

### 5. Configure GitHub Webhook

Go to your GitHub repository settings:

1. **Navigate to**: Repository → Settings → Webhooks → Add webhook

2. **Payload URL**:
   - Option A: `https://yourdomain.com/webhook`
   - Option B: `https://webhook.yourdomain.com/webhook`

3. **Content type**: `application/json`

4. **Secret**: Enter the webhook secret from step 1

5. **Which events**: Select "Just the push event"

6. **Active**: Check this box

7. Click **Add webhook**

### 6. Test the Webhook

GitHub will automatically send a ping event when you create the webhook.

Check if it worked:

```bash
docker compose logs webhook
```

You should see:
```
🏓 Received ping from GitHub - webhook is working!
```

### 7. Test Auto-Deployment

Make a small change and push to the main branch:

```bash
echo "# Test" >> README.md
git add README.md
git commit -m "Test auto-deployment"
git push origin main
```

Watch the deployment in real-time:

```bash
# Terminal 1: Webhook logs
docker compose logs -f webhook

# Terminal 2: All containers
docker compose logs -f
```

You should see:
1. ✅ Webhook received from GitHub
2. 🚀 Deployment script triggered
3. 📥 Git pull
4. 🏗️ Docker containers rebuilt
5. ▶️ Services restarted
6. ✅ Deployment complete

## Management Commands

### View Logs

```bash
# All services
docker compose logs -f

# Webhook server only
docker compose logs -f webhook

# Check deployment logs
docker compose exec webhook tail -f /var/log/terraforming-mars-deploy.log
```

### Restart Webhook Server

```bash
docker compose restart webhook
```

### Rebuild Webhook Server

After making changes to webhook configuration:

```bash
docker compose build webhook
docker compose up -d webhook
```

### Check Webhook Server Status

```bash
docker compose ps webhook
docker compose logs webhook | tail -20
```

## Troubleshooting

### GitHub Says "We couldn't deliver this payload"

**1. Check webhook container is running:**
```bash
docker compose ps webhook
```

**2. Check webhook logs for errors:**
```bash
docker compose logs webhook
```

**3. Verify Cloudflare Tunnel routing:**
```bash
# Test from outside your network
curl -X POST https://webhook.yourdomain.com/webhook

# Should return "Unauthorized" (401) but proves routing works
```

**4. Check internal connectivity:**
```bash
docker compose exec cloudflared ping webhook
docker compose exec cloudflared wget -O- http://webhook:9000/
```

### Webhook Received but Signature Verification Fails

**Check secrets match:**
```bash
# View configured secret (first few chars)
docker compose exec webhook printenv WEBHOOK_SECRET | head -c 10

# Compare with GitHub settings
```

**Update secret:**
```bash
nano .env  # Update WEBHOOK_SECRET
docker compose up -d webhook
```

### Deployment Doesn't Start

**1. Check deployment script permissions:**
```bash
ls -l deploy.sh
```

Should show: `-rwxr-xr-x` (executable)

**2. Test deployment script manually:**
```bash
docker compose exec webhook /deploy/deploy.sh
```

**3. Check Docker socket access:**
```bash
docker compose exec webhook docker ps
```

Should list all running containers.

### Deployment Fails

**Check deployment logs:**
```bash
docker compose exec webhook cat /var/log/terraforming-mars-deploy.log
```

**Check disk space:**
```bash
df -h
```

**Clean up Docker:**
```bash
docker system prune -a -f
docker volume prune -f
```

**Manually test rebuild:**
```bash
cd /path/to/repo
docker compose down
docker compose build
docker compose up -d
```

### Webhook Container Won't Start

**Check Docker socket mount:**
```bash
ls -la /var/run/docker.sock
```

**Check if user is in docker group:**
```bash
groups
docker ps
```

**View webhook container logs:**
```bash
docker compose logs webhook
```

## Security Considerations

1. **Strong Secret**: Always use a cryptographically secure webhook secret
2. **Signature Verification**: Every webhook is verified before processing
3. **HTTPS Only**: All traffic encrypted via Cloudflare Tunnel
4. **Branch Restriction**: Only `main` branch triggers deployment
5. **Docker Socket Access**: Webhook container has controlled access to Docker
6. **No Port Exposure**: Webhook only accessible through Cloudflare Tunnel

## Advanced Configuration

### Deploy on Specific Branches

Edit `webhook-server.js` to support multiple branches:

```javascript
const BRANCH_CONFIGS = {
    'main': '/deploy/deploy.sh',
    'staging': '/deploy/deploy-staging.sh'
};

// In the push event handler:
const deployScript = BRANCH_CONFIGS[branch] || null;
if (deployScript) {
    deploy(commit, branch, pusher, deployScript);
}
```

### Add Deployment Notifications

Add to the end of `deploy.sh`:

```bash
# Notify via Discord/Slack webhook
curl -X POST -H 'Content-type: application/json' \
  --data "{\"text\":\"Deployment complete for commit $COMMIT_HASH\"}" \
  YOUR_WEBHOOK_URL
```

### Custom Deploy Script

Create environment-specific deploy scripts:

```bash
# deploy-staging.sh
docker compose -f docker-compose.staging.yml down
docker compose -f docker-compose.staging.yml build
docker compose -f docker-compose.staging.yml up -d
```

### Deployment with Zero Downtime

Modify `deploy.sh` to use rolling updates:

```bash
# Build new images first
docker compose build

# Start new containers before stopping old ones
docker compose up -d --no-deps --scale backend=2 backend
sleep 5
docker compose up -d --no-deps --scale backend=1 backend
```

## Monitoring and Maintenance

### Log Rotation

Logs are stored inside the webhook container. To prevent unbounded growth:

```bash
# Check log sizes
docker compose exec webhook du -h /var/log/

# Clear logs
docker compose exec webhook sh -c 'echo "" > /var/log/webhook-server.log'
docker compose exec webhook sh -c 'echo "" > /var/log/terraforming-mars-deploy.log'
```

Or add log rotation to the Dockerfile.webhook:

```dockerfile
RUN apk add --no-cache logrotate
COPY logrotate.conf /etc/logrotate.d/webhook
```

### Health Monitoring

Add a health check endpoint to webhook-server.js:

```javascript
// Add to server route handler
if (req.url === '/health') {
    res.writeHead(200);
    res.end('OK');
    return;
}
```

Then add to docker-compose.yml:

```yaml
webhook:
  # ... existing config ...
  healthcheck:
    test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:9000/health"]
    interval: 30s
    timeout: 3s
    retries: 3
```

### Deployment History

Track deployments:

```bash
# View recent deployments
docker compose exec webhook cat /var/log/terraforming-mars-deploy.log | grep "Deployment complete"

# Export deployment history
docker compose exec webhook cat /var/log/terraforming-mars-deploy.log > deployment-history.log
```

## Rollback Procedure

If a deployment breaks your application:

```bash
# 1. Check recent commits
cd /path/to/repo
git log --oneline -5

# 2. Revert to previous commit
git reset --hard <previous-commit-hash>

# 3. Manually trigger rebuild
docker compose exec webhook /deploy/deploy.sh

# Or trigger from GitHub by pushing the rollback
git push --force origin main
```

## Complete Architecture

```
Developer
    ↓
Git Push to GitHub
    ↓
GitHub Webhook → https://webhook.yourdomain.com/webhook
    ↓
Cloudflare Tunnel (cloudflared container)
    ↓
Webhook Container (Node.js server on port 9000)
    ↓ (verifies signature)
    ↓
Deploy Script (/deploy/deploy.sh)
    ↓
Docker Socket (/var/run/docker.sock)
    ↓
Rebuild & Restart Containers
    ↓
Application Updated! ✨
```

## Files Overview

- **webhook-server.js**: Node.js HTTP server that receives webhooks
- **deploy.sh**: Bash script that pulls code and rebuilds containers
- **Dockerfile.webhook**: Docker image for webhook server
- **docker-compose.yml**: Includes webhook service configuration
- **.env**: Contains WEBHOOK_SECRET and other secrets

## Support

If you encounter issues:

1. **Check webhook logs**: `docker compose logs webhook`
2. **Check deployment logs**: `docker compose exec webhook cat /var/log/terraforming-mars-deploy.log`
3. **Test GitHub webhook delivery**: GitHub repo → Settings → Webhooks → Recent Deliveries
4. **Test manual deployment**: `docker compose exec webhook /deploy/deploy.sh`
5. **Check Docker access**: `docker compose exec webhook docker ps`
6. **Verify secrets**: Check `.env` file and GitHub webhook settings match
