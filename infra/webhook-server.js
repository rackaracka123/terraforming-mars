#!/usr/bin/env node

/**
 * GitHub Webhook Server for Terraforming Mars Auto-Deployment
 *
 * This server listens for GitHub push events and triggers automatic deployment
 * Run as a systemd service for automatic startup on boot
 */

const http = require('http');
const crypto = require('crypto');
const { exec } = require('child_process');
const fs = require('fs');
const path = require('path');

// Configuration
const PORT = process.env.WEBHOOK_PORT;
const SECRET = process.env.WEBHOOK_SECRET;
const DEPLOY_SCRIPT = process.env.DEPLOY_SCRIPT || '/deploy/infra/deploy-host.sh';
const BRANCH = process.env.DEPLOY_BRANCH;
const LOG_FILE = '/var/log/webhook-server.log';

// Logging function
function log(message) {
    const timestamp = new Date().toISOString();
    const logMessage = `[${timestamp}] ${message}\n`;
    console.log(logMessage.trim());

    // Also write to log file
    fs.appendFile(LOG_FILE, logMessage, (err) => {
        if (err && err.code !== 'EACCES') {
            console.error('Failed to write to log file:', err);
        }
    });
}

// Verify GitHub webhook signature
function verifySignature(payload, signature) {
    if (!SECRET) {
        log('⚠️  WARNING: No webhook secret configured - skipping signature verification');
        return true;
    }

    const hmac = crypto.createHmac('sha256', SECRET);
    const digest = 'sha256=' + hmac.update(payload).digest('hex');

    return crypto.timingSafeEqual(
        Buffer.from(signature || ''),
        Buffer.from(digest)
    );
}

// Execute deployment script
function deploy(commit, branch, pusher) {
    log(`🚀 Triggering deployment for commit ${commit} on branch ${branch} by ${pusher}`);

    exec(DEPLOY_SCRIPT, (error, stdout, stderr) => {
        if (error) {
            log(`❌ Deployment failed: ${error.message}`);
            log(`stderr: ${stderr}`);
            return;
        }

        log(`✅ Deployment completed successfully`);
        if (stdout) log(`stdout: ${stdout}`);
    });
}

// Create HTTP server
const server = http.createServer((req, res) => {
    // Only accept POST requests to /webhook
    if (req.method !== 'POST' || req.url !== '/webhook') {
        res.writeHead(404);
        res.end('Not Found');
        return;
    }

    let body = '';

    req.on('data', chunk => {
        body += chunk.toString();
    });

    req.on('end', () => {
        try {
            // Parse payload based on content type
            const contentType = req.headers['content-type'] || '';
            let payload;

            if (contentType.includes('application/x-www-form-urlencoded')) {
                // GitHub sends form-encoded data with payload= prefix
                const params = new URLSearchParams(body);
                const payloadStr = params.get('payload');
                if (!payloadStr) {
                    log('❌ No payload found in form data');
                    res.writeHead(400);
                    res.end('Bad Request');
                    return;
                }
                payload = JSON.parse(payloadStr);
            } else {
                // Assume JSON
                payload = JSON.parse(body);
            }

            // Verify signature (verify against the original body)
            const signature = req.headers['x-hub-signature-256'];
            if (!verifySignature(body, signature)) {
                log('❌ Invalid signature - webhook rejected');
                res.writeHead(401);
                res.end('Unauthorized');
                return;
            }

            const event = req.headers['x-github-event'];

            log(`📨 Received ${event} event from GitHub`);

            // Handle push events
            if (event === 'push') {
                const ref = payload.ref;
                const branch = ref.replace('refs/heads/', '');
                const commit = payload.after.substring(0, 7);
                const pusher = payload.pusher.name;

                log(`📍 Push to branch: ${branch}`);

                // Only deploy for specified branch
                if (branch === BRANCH) {
                    deploy(commit, branch, pusher);
                    res.writeHead(200);
                    res.end('Deployment triggered');
                } else {
                    log(`⏭️  Skipping deployment - branch ${branch} does not match ${BRANCH}`);
                    res.writeHead(200);
                    res.end('Branch ignored');
                }
            } else if (event === 'ping') {
                log('🏓 Received ping from GitHub - webhook is working!');
                res.writeHead(200);
                res.end('Pong!');
            } else {
                log(`⏭️  Ignoring ${event} event`);
                res.writeHead(200);
                res.end('Event ignored');
            }

        } catch (error) {
            log(`❌ Error processing webhook: ${error.message}`);
            res.writeHead(500);
            res.end('Internal Server Error');
        }
    });
});

// Start server
server.listen(PORT, () => {
    log(`🎧 Webhook server listening on port ${PORT}`);
    log(`📂 Deploy script: ${DEPLOY_SCRIPT}`);
    log(`🌿 Target branch: ${BRANCH}`);
    log(`🔐 Secret configured: ${SECRET ? 'Yes' : 'No (⚠️  WARNING: Signatures not verified!)'}`);
});

// Handle graceful shutdown
process.on('SIGTERM', () => {
    log('🛑 Received SIGTERM - shutting down gracefully');
    server.close(() => {
        log('👋 Server closed');
        process.exit(0);
    });
});

process.on('SIGINT', () => {
    log('🛑 Received SIGINT - shutting down gracefully');
    server.close(() => {
        log('👋 Server closed');
        process.exit(0);
    });
});
