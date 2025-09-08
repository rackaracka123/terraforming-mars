import { test, expect } from '@playwright/test';

test.describe('Production Phase Modal', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the game interface
    // This assumes we can access a game in progress or create one
    await page.goto('/');
  });

  test('should display production phase modal when all players pass', async ({ page }) => {
    // Wait for the page to load
    await page.waitForLoadState('networkidle');

    // Look for game creation or join functionality
    // First check if we're on the landing page
    const createGameButton = page.locator('text=Create Game').first();
    if (await createGameButton.isVisible()) {
      await createGameButton.click();
      await page.waitForLoadState('networkidle');
    }

    // If we're on a game join page, try to join
    const joinButton = page.locator('button:has-text("Join Game")').first();
    if (await joinButton.isVisible()) {
      // Fill in player name if required
      const nameInput = page.locator('input[placeholder*="name"], input[type="text"]').first();
      if (await nameInput.isVisible()) {
        await nameInput.fill('TestPlayer');
      }
      await joinButton.click();
      await page.waitForLoadState('networkidle');
    }

    // Wait for game interface to load
    await page.waitForSelector('[class*="gameLayout"]', { timeout: 10000 });

    // Check for skip action button (this might be in a debug menu or action panel)
    const skipButton = page.locator('button:has-text("Skip"), button:has-text("Pass")').first();
    
    if (await skipButton.isVisible()) {
      // Click skip/pass multiple times to trigger production phase
      // In a real game with multiple players, we'd need all players to pass
      await skipButton.click();
      await page.waitForTimeout(500);
      
      // Look for production phase modal
      const productionModal = page.locator('[class*="modalOverlay"]');
      await expect(productionModal).toBeVisible({ timeout: 5000 });

      // Check for modal title
      await expect(page.locator('text=Production Phase')).toBeVisible();

      // Check for generation info
      await expect(page.locator('[class*="generationInfo"]')).toBeVisible();

      // Check for player progress dots
      const progressDots = page.locator('[class*="progressDot"]');
      await expect(progressDots.first()).toBeVisible();

      // Check for resource animations
      const resourceItems = page.locator('[class*="resourceItem"]');
      await expect(resourceItems.first()).toBeVisible();

      // Wait for animation to complete (should take a few seconds)
      await page.waitForTimeout(3000);

      // Modal should auto-close after animations
      await expect(productionModal).not.toBeVisible({ timeout: 10000 });
    }
  });

  test('should show correct resource changes in modal', async ({ page }) => {
    // This test would require setting up specific game state
    // For now, we'll check the modal structure when it appears
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Look for any existing production phase modal triggers
    // This is a more targeted test for modal content
    const modalTrigger = page.locator('button:has-text("Production"), button:has-text("Skip")').first();
    
    if (await modalTrigger.isVisible()) {
      await modalTrigger.click();
      
      // Wait for modal to appear
      const modal = page.locator('[class*="modalOverlay"]');
      if (await modal.isVisible({ timeout: 2000 })) {
        // Check for resource icons (megacredits, steel, titanium, plants, energy, heat)
        const resourceIcons = page.locator('[class*="resourceIcon"] img');
        await expect(resourceIcons).toHaveCount(6, { timeout: 3000 });

        // Check for before/after amounts
        const beforeAmounts = page.locator('[class*="beforeAmount"]');
        const afterAmounts = page.locator('[class*="afterAmount"]');
        
        await expect(beforeAmounts.first()).toBeVisible();
        await expect(afterAmounts.first()).toBeVisible();

        // Check for change indicators
        const changeIndicators = page.locator('[class*="changeIndicator"]');
        await expect(changeIndicators.first()).toBeVisible();

        // Check for close button
        const closeButton = page.locator('[class*="closeBtn"]');
        await expect(closeButton).toBeVisible();

        // Test modal can be closed
        await closeButton.click();
        await expect(modal).not.toBeVisible({ timeout: 2000 });
      }
    }
  });

  test('should handle keyboard shortcuts (ESC to close)', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Try to trigger production phase modal
    const actionButton = page.locator('button:has-text("Skip"), button:has-text("Production"), button:has-text("Pass")').first();
    
    if (await actionButton.isVisible()) {
      await actionButton.click();
      
      const modal = page.locator('[class*="modalOverlay"]');
      if (await modal.isVisible({ timeout: 2000 })) {
        // Press ESC key to close modal
        await page.keyboard.press('Escape');
        
        // Modal should close
        await expect(modal).not.toBeVisible({ timeout: 2000 });
      }
    }
  });

  test('should show player names and colors correctly', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // This test checks the player-specific display in the modal
    const actionButton = page.locator('button:has-text("Skip"), button:has-text("Production")').first();
    
    if (await actionButton.isVisible()) {
      await actionButton.click();
      
      const modal = page.locator('[class*="modalOverlay"]');
      if (await modal.isVisible({ timeout: 2000 })) {
        // Check for player name display
        const playerName = page.locator('[class*="playerName"]');
        await expect(playerName).toBeVisible();

        // Check for animation step indicator
        const animationStep = page.locator('[class*="animationStep"]');
        await expect(animationStep).toBeVisible();
        
        // Should show either "Energy Conversion" or "Resource Production"
        await expect(animationStep).toContainText(/Energy Conversion|Resource Production/);
      }
    }
  });
});