import { test, expect } from '@playwright/test';

test.describe('Game Flow Debugging', () => {
  test('should complete full game creation and start flow', async ({ page }) => {
    // Navigate to landing page
    await page.goto('http://localhost:3000');
    await page.waitForLoadState('networkidle');

    console.log('📍 Step 1: Landing page loaded');
    await page.screenshot({ path: 'test-results/01-landing-page.png' });

    // Check for create/join game buttons
    const createButton = page.getByRole('button', { name: /create/i });
    const joinButton = page.getByRole('button', { name: /join/i });

    await expect(createButton).toBeVisible();
    await expect(joinButton).toBeVisible();
    console.log('✅ Create and Join buttons visible');

    // Click Create Game
    await createButton.click();
    await page.waitForTimeout(1000);

    console.log('📍 Step 2: Create game clicked');
    await page.screenshot({ path: 'test-results/02-after-create-click.png' });

    // Should navigate to waiting room/lobby
    await page.waitForURL(/.*/, { timeout: 5000 });
    const currentUrl = page.url();
    console.log('📍 Current URL:', currentUrl);

    // Check for lobby/waiting room elements
    const bodyText = await page.textContent('body');
    console.log('📍 Page content preview:', bodyText?.substring(0, 500));

    await page.screenshot({ path: 'test-results/03-lobby-page.png' });

    // Look for start game button (should be visible to host)
    const startGameButton = page.getByRole('button', { name: /start/i });

    if (await startGameButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      console.log('✅ Start Game button found');
      await page.screenshot({ path: 'test-results/04-start-button-visible.png' });

      // Click Start Game
      await startGameButton.click();
      await page.waitForTimeout(1000);

      console.log('📍 Step 3: Start game clicked');
      await page.screenshot({ path: 'test-results/05-after-start-click.png' });

      // Wait for game to start - should show corporation selection or game view
      await page.waitForTimeout(2000);

      const newBodyText = await page.textContent('body');
      console.log('📍 After start - page content:', newBodyText?.substring(0, 500));

      await page.screenshot({ path: 'test-results/06-game-started.png' });

      // Check for corporation selection or game elements
      const hasCorporation = await page.getByText(/corporation/i).count();
      const hasResources = await page.getByText(/credit|steel|titanium/i).count();

      console.log('📊 Elements found:', {
        corporationMentions: hasCorporation,
        resourceMentions: hasResources,
      });

    } else {
      console.log('❌ Start Game button not found');
    }

    // Print console logs from browser
    page.on('console', msg => console.log('🖥️ Browser console:', msg.text()));

    // Print any errors
    page.on('pageerror', error => console.error('❌ Page error:', error.message));
  });

  test('should show WebSocket connection status', async ({ page }) => {
    await page.goto('http://localhost:3000');

    // Wait for page load
    await page.waitForLoadState('networkidle');

    // Listen for WebSocket connections
    const wsMessages: string[] = [];
    page.on('websocket', ws => {
      console.log('🔌 WebSocket opened:', ws.url());
      ws.on('framesent', event => {
        console.log('📤 WS Send:', event.payload);
        wsMessages.push(`SENT: ${event.payload}`);
      });
      ws.on('framereceived', event => {
        console.log('📥 WS Receive:', event.payload);
        wsMessages.push(`RECEIVED: ${event.payload}`);
      });
      ws.on('close', () => console.log('🔌 WebSocket closed'));
    });

    // Click create game to trigger WebSocket
    await page.getByRole('button', { name: /create/i }).click();
    await page.waitForTimeout(2000);

    console.log('📊 Total WebSocket messages:', wsMessages.length);
    console.log('📝 Messages:', wsMessages);
  });

  test('should inspect game state in Debug panel', async ({ page }) => {
    await page.goto('http://localhost:3000');
    await page.waitForLoadState('networkidle');

    // Create game
    await page.getByRole('button', { name: /create/i }).click();
    await page.waitForTimeout(1000);

    // Check if Debug panel exists
    const debugPanel = page.locator('[class*="debug" i]');
    const debugPanelExists = await debugPanel.count();

    console.log('🔍 Debug panels found:', debugPanelExists);

    if (debugPanelExists > 0) {
      await page.screenshot({ path: 'test-results/debug-panel.png' });

      // Try to extract game state from debug panel
      const debugText = await debugPanel.first().textContent();
      console.log('🎮 Debug panel content:', debugText);
    }

    // Check localStorage for game data
    const gameData = await page.evaluate(() => {
      return localStorage.getItem('terraformingMarsGameData');
    });

    console.log('💾 LocalStorage game data:', gameData);

    // Check for any game ID in the page
    const pageContent = await page.content();
    const gameIdMatch = pageContent.match(/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i);

    if (gameIdMatch) {
      console.log('🎯 Found game ID:', gameIdMatch[0]);
    }
  });
});
