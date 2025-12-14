import { test, expect } from '@playwright/test';

// Mobile viewport
const mobileViewport = { width: 375, height: 667 };

// Helper to login - returns true if successful
async function tryLogin(page: any): Promise<boolean> {
  await page.goto('/login');
  await page.fill('#username', 'admin');
  await page.fill('#password', 'changeme123');
  await page.click('button[type="submit"]');

  // Wait a bit for response
  await page.waitForTimeout(1000);

  // Check if 2FA is required
  const twoFAInput = page.locator('#twofa_code');
  const is2FA = await twoFAInput.isVisible().catch(() => false);

  if (is2FA) {
    console.log('2FA required - cannot complete login automatically');
    return false;
  }

  // Check if we're redirected to dashboard
  const currentUrl = page.url();
  if (currentUrl.includes('/login')) {
    return false;
  }

  return true;
}

test.describe('Mobile UI Analysis - All Pages', () => {
  test('Dashboard page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);

    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      // Take screenshot of login/2FA page
      await page.screenshot({
        path: 'test-results/mobile-dashboard-needs-auth.png',
        fullPage: true
      });
      console.log('Skipping dashboard test - needs manual 2FA');
      test.skip();
      return;
    }

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    await page.screenshot({
      path: 'test-results/mobile-dashboard.png',
      fullPage: true
    });

    await expect(page.locator('h1')).toContainText('Dashboard');

    // Check for horizontal overflow
    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    expect(bodyWidth).toBeLessThanOrEqual(mobileViewport.width + 1);
  });

  test('Logs page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      test.skip();
      return;
    }

    await page.goto('/logs');
    await page.waitForLoadState('networkidle');

    await page.screenshot({
      path: 'test-results/mobile-logs.png',
      fullPage: true
    });

    await expect(page.locator('h1')).toContainText('Logs');

    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    if (bodyWidth > mobileViewport.width) {
      console.log(`WARNING: Logs page overflow! Width: ${bodyWidth}`);
    }
  });

  test('Projects page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      test.skip();
      return;
    }

    await page.goto('/projects');
    await page.waitForLoadState('networkidle');

    await page.screenshot({
      path: 'test-results/mobile-projects.png',
      fullPage: true
    });

    await expect(page.locator('h1')).toContainText('Projects');

    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    expect(bodyWidth).toBeLessThanOrEqual(mobileViewport.width + 1);
  });

  test('Users page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      test.skip();
      return;
    }

    await page.goto('/users');
    await page.waitForLoadState('networkidle');

    await page.screenshot({
      path: 'test-results/mobile-users.png',
      fullPage: true
    });

    await expect(page.locator('h1')).toContainText('Users');

    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    if (bodyWidth > mobileViewport.width) {
      console.log(`WARNING: Users page overflow! Width: ${bodyWidth}`);
    }
  });

  test('Settings page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      test.skip();
      return;
    }

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');

    await page.screenshot({
      path: 'test-results/mobile-settings.png',
      fullPage: true
    });

    await expect(page.locator('h1')).toContainText('Settings');
    await expect(page.locator('text=Profile').first()).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Two-Factor Authentication' })).toBeVisible();

    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    expect(bodyWidth).toBeLessThanOrEqual(mobileViewport.width + 1);
  });

  test('Sidebar/Navigation on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      test.skip();
      return;
    }

    await page.goto('/');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(500); // Brief wait for content

    // Check sidebar
    const sidebar = page.locator('aside').first();
    const sidebarVisible = await sidebar.isVisible().catch(() => false);

    if (sidebarVisible) {
      const sidebarBox = await sidebar.boundingBox().catch(() => null);
      if (sidebarBox) {
        console.log(`Sidebar: ${sidebarBox.width}x${sidebarBox.height} at (${sidebarBox.x}, ${sidebarBox.y})`);
      }
    } else {
      console.log('Sidebar not visible on mobile (expected for collapsed/hidden sidebar)');
    }

    await page.screenshot({
      path: 'test-results/mobile-sidebar.png',
      fullPage: true
    });
  });

  test('Project Detail page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    const loggedIn = await tryLogin(page);

    if (!loggedIn) {
      test.skip();
      return;
    }

    await page.goto('/projects');
    await page.waitForLoadState('networkidle');

    const projectLink = page.locator('a[href^="/projects/"]').first();
    const hasProject = await projectLink.isVisible().catch(() => false);

    if (hasProject) {
      await projectLink.click();
      await page.waitForLoadState('networkidle');

      await page.screenshot({
        path: 'test-results/mobile-project-detail.png',
        fullPage: true
      });

      const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
      if (bodyWidth > mobileViewport.width) {
        console.log(`WARNING: Project detail overflow! Width: ${bodyWidth}`);
      }
    } else {
      console.log('No projects to test');
    }
  });
});

// Login page tests don't need auth
test.describe('Mobile UI - Login Page', () => {
  test('Login page on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    await page.goto('/login');

    await page.screenshot({
      path: 'test-results/mobile-login-page.png',
      fullPage: true
    });

    await expect(page.locator('#username')).toBeVisible();
    await expect(page.locator('#password')).toBeVisible();

    const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
    expect(bodyWidth).toBeLessThanOrEqual(mobileViewport.width + 1);
  });

  test('2FA step on mobile', async ({ page }) => {
    await page.setViewportSize(mobileViewport);
    await page.goto('/login');

    await page.fill('#username', 'admin');
    await page.fill('#password', 'changeme123');
    await page.click('button[type="submit"]');

    await page.waitForTimeout(1500);

    // Check if 2FA appeared
    const twoFAInput = page.locator('#twofa_code');
    const is2FA = await twoFAInput.isVisible().catch(() => false);

    if (is2FA) {
      await page.screenshot({
        path: 'test-results/mobile-2fa-step.png',
        fullPage: true
      });

      await expect(twoFAInput).toBeVisible();
      await expect(page.locator('text=Two-Factor Authentication')).toBeVisible();

      // Check back button
      const backBtn = page.locator('button svg.lucide-arrow-left').first();
      await expect(backBtn).toBeVisible();
    }
  });
});
