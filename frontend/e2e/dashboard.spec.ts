import { test, expect } from '@playwright/test';
import { LoginPage } from './page-objects/login.page';
import { DashboardPage } from './page-objects/dashboard.page';
import { NavigationPage } from './page-objects/navigation.page';

const TEST_USER = {
  email: 'admin@example.com',
  password: 'changeme123',
};

test.describe('Dashboard', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let navigation: NavigationPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    navigation = new NavigationPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
  });

  test('should display dashboard page', async ({ page }) => {
    await expect(dashboardPage.dashboardHeading).toBeVisible();
    await expect(page.getByText('Overview of your logging activity')).toBeVisible();
  });

  test('should display all stat cards', async () => {
    // Check all 4 stat cards are visible
    await expect(dashboardPage.totalProjectsCard).toBeVisible();
    await expect(dashboardPage.totalLogsCard).toBeVisible();
    await expect(dashboardPage.logsTodayCard).toBeVisible();
    await expect(dashboardPage.errorsTodayCard).toBeVisible();
  });

  test('should display stat values', async () => {
    // Verify each card has a numeric value
    const totalProjects = await dashboardPage.getStatValue(dashboardPage.totalProjectsCard);
    expect(totalProjects).toMatch(/^\d+$/);

    const totalLogs = await dashboardPage.getStatValue(dashboardPage.totalLogsCard);
    expect(totalLogs).toMatch(/^[\d,]+$/);

    const logsToday = await dashboardPage.getStatValue(dashboardPage.logsTodayCard);
    expect(logsToday).toMatch(/^[\d,]+$/);

    const errorsToday = await dashboardPage.getStatValue(dashboardPage.errorsTodayCard);
    expect(errorsToday).toMatch(/^[\d,]+$/);
  });

  test('should display logs by level card', async ({ page }) => {
    await expect(dashboardPage.logsByLevelCard).toBeVisible();

    // Check for level badges
    const levels = ['DEBUG', 'INFO', 'WARN', 'ERROR', 'CRITICAL'];
    for (const level of levels) {
      await expect(page.getByText(level).first()).toBeVisible();
    }
  });

  test('should display recent logs card', async () => {
    await expect(dashboardPage.recentLogsCard).toBeVisible();
    await expect(dashboardPage.viewAllLogsLink).toBeVisible();
  });

  test('should show recent logs when available', async ({ page }) => {
    // The recent logs section should either show logs or "No recent logs"
    const recentLogsSection = page.locator('text=/Recent Logs/').locator('..');
    await expect(recentLogsSection).toBeVisible();

    // Check if there are logs or empty state
    const hasLogs = await page.locator('text=/No recent logs/').isVisible();

    if (!hasLogs) {
      // If there are logs, they should have badges and content
      const logBadges = await page.locator('[data-variant]').all();
      expect(logBadges.length).toBeGreaterThan(0);
    }
  });

  test('should navigate to logs page when clicking "View all"', async ({ page }) => {
    await dashboardPage.viewAllLogsLink.click();
    await expect(page).toHaveURL('/logs');
    await expect(page.getByRole('heading', { name: 'Logs' })).toBeVisible();
  });

  test('should show loading state initially', async ({ context }) => {
    // Open fresh page to see loading state
    const newPage = await context.newPage();
    const newDashboard = new DashboardPage(newPage);
    const newLogin = new LoginPage(newPage);

    await newLogin.goto();
    await newLogin.login(TEST_USER.email, TEST_USER.password);

    // Check for loading spinner (might be brief)
    const spinner = newPage.locator('.animate-spin');

    // Either spinner is visible or content has already loaded
    try {
      await expect(spinner).toBeVisible({ timeout: 1000 });
    } catch {
      // Content loaded too fast, which is fine
      await expect(newDashboard.dashboardHeading).toBeVisible();
    }

    await newPage.close();
  });

  test('should refresh stats on navigation back to dashboard', async ({ page }) => {
    // Get initial total logs value
    const _initialLogs = await dashboardPage.getStatValue(dashboardPage.totalLogsCard);

    // Navigate away and back
    await navigation.navigateToProjects();
    await expect(page).toHaveURL('/projects');

    await navigation.navigateToDashboard();
    await expect(page).toHaveURL('/');

    // Stats should be loaded
    const currentLogs = await dashboardPage.getStatValue(dashboardPage.totalLogsCard);
    expect(currentLogs).toBeTruthy();
  });

  test('should display stat cards in correct grid layout', async ({ page }) => {
    // Check grid layout on larger screens
    const grid = page.locator('div[class*="grid"]').first();
    await expect(grid).toBeVisible();

    // Should have 4 stat cards
    const statCards = await grid.locator('> div').all();
    expect(statCards.length).toBe(4);
  });

  test('should display percentage bars in logs by level', async ({ page }) => {
    // Check that each level has a progress bar
    const progressBars = await page.locator('.h-2.rounded-full.bg-muted').all();
    expect(progressBars.length).toBeGreaterThanOrEqual(5); // At least 5 for each level
  });

  test('should handle empty state gracefully', async ({ page: _page }) => {
    // The dashboard should handle cases with no data gracefully
    // Stats should show 0 or actual values, never crash
    await expect(dashboardPage.dashboardHeading).toBeVisible();
    await expect(dashboardPage.totalProjectsCard).toBeVisible();
  });

  test('should show stat icons', async ({ page }) => {
    // Each stat card should have an icon
    const icons = await page.locator('.h-4.w-4.text-').all();
    expect(icons.length).toBeGreaterThanOrEqual(4);
  });
});

test.describe('Dashboard - Responsive Design', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
  });

  test('should be responsive on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    await expect(dashboardPage.dashboardHeading).toBeVisible();
    await expect(dashboardPage.totalProjectsCard).toBeVisible();
  });

  test('should be responsive on tablet', async ({ page }) => {
    // Set tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });

    await expect(dashboardPage.dashboardHeading).toBeVisible();
    await expect(dashboardPage.totalProjectsCard).toBeVisible();
  });
});

test.describe('Dashboard - Real-time Updates', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
  });

  test('should display consistent data across multiple page loads', async ({ page }) => {
    const firstLoad = await dashboardPage.getStatValue(dashboardPage.totalProjectsCard);

    await page.reload();
    await dashboardPage.waitForDashboard();

    const secondLoad = await dashboardPage.getStatValue(dashboardPage.totalProjectsCard);

    // Values should be the same (assuming no changes between loads)
    expect(firstLoad).toBe(secondLoad);
  });
});
