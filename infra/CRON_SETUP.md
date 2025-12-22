# Auto-Deploy Cron Job Setup

Simple cron job that checks for new commits every minute and auto-deploys.

## Installation

1. **Make script executable** (already done):

   ```bash
   chmod +x /home/mhm/terraforming-mars/infra/auto-deploy-cron.sh
   ```

2. **Add to crontab**:

   ```bash
   crontab -e
   ```

3. **Add this line** (checks every minute):

   ```
   * * * * * /home/mhm/terraforming-mars/infra/auto-deploy-cron.sh
   ```

   Or check every 5 minutes:

   ```
   */5 * * * * /home/mhm/terraforming-mars/infra/auto-deploy-cron.sh
   ```

## How It Works

1. **Fetch**: Gets latest commits from GitHub
2. **Compare**: Checks if local is behind remote
3. **Deploy**: If behind, pulls code and runs `docker compose up -d --build`
4. **Lock**: Uses lock file to prevent concurrent deployments

## Logs

View deployment logs:

```bash
tail -f /tmp/terraforming-mars-autodeploy.log
```

## Manual Trigger

Run deployment manually:

```bash
/home/mhm/terraforming-mars/infra/auto-deploy-cron.sh
```

## Disable

Remove from crontab:

```bash
crontab -e
# Delete the line with auto-deploy-cron.sh
```
