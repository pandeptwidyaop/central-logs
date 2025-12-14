import { test, expect } from '@playwright/test';

// Mobile device viewports
const mobileDevices = [
  { name: 'iPhone SE', width: 375, height: 667 },
  { name: 'iPhone 12', width: 390, height: 844 },
  { name: 'iPhone 14 Pro Max', width: 430, height: 932 },
  { name: 'Samsung Galaxy S21', width: 360, height: 800 },
  { name: 'Pixel 5', width: 393, height: 851 },
];

test.describe('Mobile Login Page Analysis', () => {
  for (const device of mobileDevices) {
    test(`Login page on ${device.name} (${device.width}x${device.height})`, async ({ page }) => {
      // Set viewport
      await page.setViewportSize({ width: device.width, height: device.height });

      // Go to login page
      await page.goto('/login');

      // Wait for page to load
      await page.waitForSelector('form');

      // Take screenshot
      await page.screenshot({
        path: `test-results/mobile-login-${device.name.replace(/\s+/g, '-').toLowerCase()}.png`,
        fullPage: true
      });

      // Check that mobile logo is visible (lg:hidden means visible on mobile)
      const mobileLogo = page.locator('.lg\\:hidden img[alt="Central Logs"]');
      await expect(mobileLogo).toBeVisible();

      // Check that desktop branding is hidden on mobile
      const desktopBranding = page.locator('.lg\\:flex.lg\\:w-1\\/2');
      await expect(desktopBranding).not.toBeVisible();

      // Check form elements are visible and usable
      const usernameInput = page.locator('#username');
      const passwordInput = page.locator('#password');
      const submitButton = page.locator('button[type="submit"]');

      await expect(usernameInput).toBeVisible();
      await expect(passwordInput).toBeVisible();
      await expect(submitButton).toBeVisible();

      // Check for overflow issues - content should fit within viewport
      const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
      const viewportWidth = device.width;

      if (bodyWidth > viewportWidth) {
        console.log(`WARNING: Horizontal overflow detected on ${device.name}! Body width: ${bodyWidth}, Viewport: ${viewportWidth}`);
      }

      expect(bodyWidth).toBeLessThanOrEqual(viewportWidth + 1); // Allow 1px tolerance

      // Check card is properly sized
      const card = page.locator('.border-0.shadow-lg');
      const cardBox = await card.boundingBox();

      if (cardBox) {
        console.log(`${device.name} - Card dimensions: ${cardBox.width}x${cardBox.height}`);

        // Card should not exceed viewport width minus padding
        expect(cardBox.width).toBeLessThanOrEqual(device.width - 32); // 16px padding on each side
      }
    });
  }

  test('Login page - 2FA step on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    // Go to login page
    await page.goto('/login');

    // Fill in credentials
    await page.fill('#username', 'admin');
    await page.fill('#password', 'changeme123');

    // Submit form
    await page.click('button[type="submit"]');

    // Wait for either redirect or 2FA prompt
    await page.waitForTimeout(2000);

    // Take screenshot of whatever state we're in
    await page.screenshot({
      path: 'test-results/mobile-login-after-submit.png',
      fullPage: true
    });

    // Check if 2FA prompt appeared
    const twoFAInput = page.locator('#twofa_code');
    const is2FAVisible = await twoFAInput.isVisible().catch(() => false);

    if (is2FAVisible) {
      console.log('2FA prompt is visible on mobile');

      // Take screenshot of 2FA step
      await page.screenshot({
        path: 'test-results/mobile-login-2fa-step.png',
        fullPage: true
      });

      // Check 2FA elements
      await expect(twoFAInput).toBeVisible();

      // Check back button is visible
      const backButton = page.locator('button').filter({ has: page.locator('svg.lucide-arrow-left') });
      await expect(backButton).toBeVisible();
    }
  });

  test('Login page - Landscape orientation', async ({ page }) => {
    // Set landscape mobile viewport
    await page.setViewportSize({ width: 667, height: 375 });

    await page.goto('/login');
    await page.waitForSelector('form');

    await page.screenshot({
      path: 'test-results/mobile-login-landscape.png',
      fullPage: true
    });

    // Check form is still usable in landscape
    const usernameInput = page.locator('#username');
    await expect(usernameInput).toBeVisible();
  });
});
