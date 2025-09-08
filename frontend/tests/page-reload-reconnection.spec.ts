import { test, expect } from '@playwright/test';

test.describe('Page Reload Reconnection', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the main page
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should reconnect successfully after page reload during active game', async ({ page }) => {
    // Step 1: Create or join a game
    console.log('🎮 Test: Setting up game...');
    
    // Look for game creation functionality
    const createGameButton = page.locator('text=Create Game').first();
    if (await createGameButton.isVisible()) {
      await createGameButton.click();
      await page.waitForLoadState('networkidle');
    }

    // Fill in player name if required
    const nameInput = page.locator('input[placeholder*="name"], input[type="text"]').first();
    if (await nameInput.isVisible()) {
      await nameInput.fill('TestPlayerReconnect');
    }

    // Join the game
    const joinButton = page.locator('button:has-text("Join Game"), button:has-text("Join")').first();
    if (await joinButton.isVisible()) {
      await joinButton.click();
      await page.waitForLoadState('networkidle');
    }

    // Step 2: Wait for game interface to load and start game if necessary
    await page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 10000 });
    console.log('✅ Test: Game interface loaded');

    // Start game if we're in lobby phase
    const startGameButton = page.locator('button:has-text("Start Game")').first();
    if (await startGameButton.isVisible()) {
      await startGameButton.click();
      await page.waitForTimeout(2000); // Wait for game to start
      console.log('✅ Test: Game started');
    }

    // Step 3: Verify we're in an active game (not lobby)
    // Check that waiting room overlay is not visible
    const waitingRoomOverlay = page.locator('[class*="waitingRoom"], text="Waiting for players"');
    await expect(waitingRoomOverlay).not.toBeVisible({ timeout: 5000 });
    console.log('✅ Test: Confirmed active game state');

    // Step 4: Verify localStorage contains game data
    const gameData = await page.evaluate(() => {
      return localStorage.getItem('terraforming-mars-game');
    });
    expect(gameData).toBeTruthy();
    console.log('✅ Test: Game data found in localStorage');

    // Step 5: Perform page reload
    console.log('🔄 Test: Performing page reload...');
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Step 6: Verify reconnection flow
    // The page should first redirect to /reconnecting
    console.log('🔄 Test: Checking reconnection flow...');
    
    // Wait for either reconnecting page or direct game interface
    await Promise.race([
      page.waitForSelector('text="Reconnecting to game"', { timeout: 5000 }).catch(() => null),
      page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 5000 })
    ]);

    // If we see the reconnecting page, wait for it to complete
    const reconnectingText = page.locator('text="Reconnecting to game"');
    if (await reconnectingText.isVisible()) {
      console.log('🔄 Test: On reconnecting page, waiting for completion...');
      await expect(reconnectingText).not.toBeVisible({ timeout: 15000 });
    }

    // Step 7: Verify we're back in the active game (not lobby)
    await page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 10000 });
    console.log('✅ Test: Back on game interface');

    // Critical check: Ensure we're NOT stuck on "Waiting for players"
    const waitingOverlayAfterReconnect = page.locator('[class*="waitingRoom"], text="Waiting for players"');
    await expect(waitingOverlayAfterReconnect).not.toBeVisible({ timeout: 3000 });
    console.log('✅ Test: Confirmed NOT stuck on waiting for players screen');

    // Step 8: Verify game functionality is working
    // Look for game elements that indicate active gameplay
    const gameElements = page.locator('[class*="resourceBar"], [class*="playerArea"], [class*="gameView"]');
    await expect(gameElements.first()).toBeVisible({ timeout: 5000 });
    console.log('✅ Test: Game elements are visible and functional');
  });

  test('should handle reconnection during lobby phase correctly', async ({ page }) => {
    // Step 1: Create a game but don't start it (stay in lobby)
    console.log('🎮 Test: Setting up lobby game...');
    
    const createGameButton = page.locator('text=Create Game').first();
    if (await createGameButton.isVisible()) {
      await createGameButton.click();
      await page.waitForLoadState('networkidle');
    }

    // Fill in player name
    const nameInput = page.locator('input[placeholder*="name"], input[type="text"]').first();
    if (await nameInput.isVisible()) {
      await nameInput.fill('TestPlayerLobby');
    }

    // Join but don't start the game
    const joinButton = page.locator('button:has-text("Join Game"), button:has-text("Join")').first();
    if (await joinButton.isVisible()) {
      await joinButton.click();
      await page.waitForLoadState('networkidle');
    }

    // Step 2: Verify we're in lobby phase
    await page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 10000 });
    const waitingRoomOverlay = page.locator('[class*="waitingRoom"], text="Waiting for players"');
    await expect(waitingRoomOverlay).toBeVisible({ timeout: 5000 });
    console.log('✅ Test: Confirmed lobby state');

    // Step 3: Perform page reload
    console.log('🔄 Test: Performing page reload in lobby...');
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Step 4: Verify we reconnect to lobby correctly
    await Promise.race([
      page.waitForSelector('text="Reconnecting to game"', { timeout: 5000 }).catch(() => null),
      page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 5000 })
    ]);

    // Wait for reconnection to complete
    const reconnectingText = page.locator('text="Reconnecting to game"');
    if (await reconnectingText.isVisible()) {
      await expect(reconnectingText).not.toBeVisible({ timeout: 15000 });
    }

    // Step 5: Verify we're back in lobby (this should show waiting room)
    await page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 10000 });
    const waitingOverlayAfterReconnect = page.locator('[class*="waitingRoom"], text="Waiting for players"');
    await expect(waitingOverlayAfterReconnect).toBeVisible({ timeout: 5000 });
    console.log('✅ Test: Successfully reconnected to lobby phase');
  });

  test('should handle multiple tab conflict correctly', async ({ page, context }) => {
    // Step 1: Set up a game in first tab
    console.log('🎮 Test: Setting up game in first tab...');
    
    const createGameButton = page.locator('text=Create Game').first();
    if (await createGameButton.isVisible()) {
      await createGameButton.click();
      await page.waitForLoadState('networkidle');
    }

    const nameInput = page.locator('input[placeholder*="name"], input[type="text"]').first();
    if (await nameInput.isVisible()) {
      await nameInput.fill('TestPlayerMultiTab');
    }

    const joinButton = page.locator('button:has-text("Join Game"), button:has-text("Join")').first();
    if (await joinButton.isVisible()) {
      await joinButton.click();
      await page.waitForLoadState('networkidle');
    }

    await page.waitForSelector('[class*="gameLayout"], [class*="gameInterface"]', { timeout: 10000 });
    console.log('✅ Test: First tab game ready');

    // Step 2: Open second tab with same game
    const secondPage = await context.newPage();
    await secondPage.goto(page.url());
    await secondPage.waitForLoadState('networkidle');

    // Step 3: Check for tab conflict handling
    const tabConflictOverlay = secondPage.locator('[class*="tabConflict"], text="Another tab"');
    await expect(tabConflictOverlay).toBeVisible({ timeout: 10000 });
    console.log('✅ Test: Tab conflict detected correctly');

    // Clean up
    await secondPage.close();
  });

  test('should handle reconnection failure gracefully', async ({ page }) => {
    // Step 1: Manually add invalid localStorage data
    await page.evaluate(() => {
      localStorage.setItem('terraforming-mars-game', JSON.stringify({
        gameId: 'invalid-game-id',
        playerId: 'invalid-player-id',
        playerName: 'TestPlayerInvalid',
        timestamp: Date.now()
      }));
    });

    // Step 2: Navigate to game interface (should trigger reconnection)
    await page.goto('/game');
    await page.waitForLoadState('networkidle');

    // Step 3: Should redirect to reconnecting page
    await page.waitForSelector('text="Reconnecting to game"', { timeout: 5000 });
    console.log('✅ Test: Reconnecting page shown for invalid data');

    // Step 4: Should show error after failed reconnection
    const errorMessage = page.locator('[class*="error"], text="Reconnection Failed"');
    await expect(errorMessage).toBeVisible({ timeout: 15000 });
    console.log('✅ Test: Reconnection failure handled gracefully');

    // Step 5: Should provide option to return to main menu
    const returnButton = page.locator('button:has-text("Return to Main Menu"), button:has-text("Main Menu")');
    await expect(returnButton).toBeVisible({ timeout: 5000 });
    console.log('✅ Test: Return to main menu option available');
  });
});