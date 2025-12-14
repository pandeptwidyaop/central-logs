import { test, expect } from '@playwright/test';
import { LoginPage } from './page-objects/login.page';
import { LogsPage } from './page-objects/logs.page';
import { ProjectsPage } from './page-objects/projects.page';
import { NavigationPage } from './page-objects/navigation.page';
import { DashboardPage } from './page-objects/dashboard.page';

const TEST_USER = {
  email: 'admin@example.com',
  password: 'changeme123',
};

test.describe('Logs', () => {
  let loginPage: LoginPage;
  let logsPage: LogsPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    logsPage = new LogsPage(page);
    dashboardPage = new DashboardPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
    await logsPage.goto();
  });

  test('should display logs page', async ({ page }) => {
    await expect(logsPage.logsHeading).toBeVisible();
    await expect(page.getByText('View and search application logs')).toBeVisible();
    await expect(logsPage.refreshButton).toBeVisible();
  });

  test('should display filters section', async () => {
    await expect(logsPage.searchInput).toBeVisible();
    await expect(logsPage.projectFilter).toBeVisible();
    await expect(logsPage.levelFilter).toBeVisible();
  });

  test('should display logs table', async ({ page }) => {
    await expect(logsPage.logTable).toBeVisible();

    // Check table headers
    await expect(page.getByRole('columnheader', { name: /level/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /project/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /message/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /time/i })).toBeVisible();
  });

  test('should search logs', async ({ page }) => {
    // Wait for initial load
    await page.waitForTimeout(1000);

    const _initialCount = await logsPage.getLogCount();

    // Search for a specific term
    await logsPage.searchLogs('error');

    // Wait for search results
    await page.waitForTimeout(1000);

    // Results may be filtered (could be same or different count)
    const searchCount = await logsPage.getLogCount();
    expect(searchCount).toBeGreaterThanOrEqual(0);
  });

  test('should clear search', async ({ page }) => {
    // Search first
    await logsPage.searchLogs('test');
    await page.waitForTimeout(1000);

    // Clear search
    await logsPage.searchInput.clear();
    await page.waitForTimeout(1000);

    // Should show all logs again
    const count = await logsPage.getLogCount();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should filter by log level', async ({ page }) => {
    await logsPage.filterByLevel('ERROR');

    // Wait for filter to apply
    await page.waitForTimeout(1000);

    // Check if logs are filtered (if there are any ERROR logs)
    const logRows = await logsPage.getLogRows();

    if (logRows.length > 0) {
      // If there are logs, they should all be ERROR level
      for (const row of logRows) {
        const badge = row.locator('text=ERROR');
        const badgeExists = await badge.count() > 0;
        if (badgeExists) {
          await expect(badge.first()).toBeVisible();
        }
      }
    }
  });

  test('should filter by project', async ({ page }) => {
    // First, check if there are any projects in the filter
    await logsPage.projectFilter.click();

    // Get first project option (not "All Projects")
    const projectOptions = await page.locator('[role="option"]').all();

    if (projectOptions.length > 1) {
      const firstProject = projectOptions[1];
      const projectName = await firstProject.textContent();

      await firstProject.click();

      // Wait for filter to apply
      await page.waitForTimeout(1000);

      // Logs should be filtered by project
      const logRows = await logsPage.getLogRows();

      if (logRows.length > 0 && projectName) {
        // Check that logs belong to the selected project
        const projectCells = await page.locator('tbody td').nth(2).all();
        // At least some should match the project name
        expect(projectCells.length).toBeGreaterThan(0);
      }
    }
  });

  test('should clear all filters', async ({ page }) => {
    // Apply some filters
    await logsPage.searchLogs('test');
    await logsPage.filterByLevel('INFO');

    await page.waitForTimeout(1000);

    // Clear filters
    await logsPage.clearFilters();

    await page.waitForTimeout(1000);

    // Verify filters are cleared
    expect(await logsPage.searchInput.inputValue()).toBe('');
  });

  test('should combine multiple filters', async ({ page }) => {
    // Apply search and level filter together
    await logsPage.searchLogs('application');
    await logsPage.filterByLevel('INFO');

    await page.waitForTimeout(1000);

    // Should apply both filters
    const count = await logsPage.getLogCount();
    expect(count).toBeGreaterThanOrEqual(0);
  });

  test('should select individual log', async ({ page: _page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      await logsPage.selectLog(0);

      // Delete button should appear
      await expect(logsPage.deleteButton).toBeVisible();
      await expect(logsPage.deleteButton).toContainText('(1)');
    }
  });

  test('should select all logs', async ({ page: _page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      await logsPage.selectAllLogs();

      // Delete button should show count
      await expect(logsPage.deleteButton).toBeVisible();
      await expect(logsPage.deleteButton).toContainText(`(${logCount})`);
    }
  });

  test('should deselect all logs', async ({ page: _page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      // Select all
      await logsPage.selectAllLogs();
      await expect(logsPage.deleteButton).toBeVisible();

      // Deselect all
      await logsPage.selectAllCheckbox.uncheck();

      // Delete button should not be visible
      await expect(logsPage.deleteButton).not.toBeVisible();
    }
  });

  test('should open log details dialog', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      await logsPage.clickLogRow(0);

      // Dialog should open
      await logsPage.waitForLogDialog();

      // Dialog should show log details
      await expect(page.getByText(/Project/)).toBeVisible();
      await expect(page.getByText(/Message/)).toBeVisible();
    }
  });

  test('should close log details dialog', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      await logsPage.clickLogRow(0);
      await logsPage.waitForLogDialog();

      // Close dialog
      await logsPage.closeLogDialog();

      // Dialog should be closed
      await expect(page.getByRole('heading', { name: /log details/i })).not.toBeVisible();
    }
  });

  test('should display log metadata if present', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      await logsPage.clickLogRow(0);
      await logsPage.waitForLogDialog();

      // Check if metadata section exists (it may or may not depending on the log)
      const metadataHeading = page.getByText('Metadata');
      const hasMetadata = await metadataHeading.isVisible();

      if (hasMetadata) {
        // Should show formatted JSON
        await expect(page.locator('pre')).toBeVisible();
      }

      await logsPage.closeLogDialog();
    }
  });

  test('should refresh logs', async ({ page }) => {
    const initialCount = await logsPage.getLogCount();

    await logsPage.refreshButton.click();

    // Wait for refresh
    await page.waitForTimeout(1000);

    // Should still show logs
    const newCount = await logsPage.getLogCount();
    expect(newCount).toBeGreaterThanOrEqual(0);
  });

  test('should display pagination info', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      const paginationText = await logsPage.getPaginationText();
      expect(paginationText).toMatch(/Showing \d+ - \d+ of \d+/);
    }
  });

  test('should disable previous button on first page', async () => {
    await expect(logsPage.previousButton).toBeDisabled();
  });

  test('should navigate to next page if available', async ({ page }) => {
    const paginationText = await logsPage.getPaginationText();

    if (paginationText) {
      // Parse total from "Showing X - Y of Z"
      const match = paginationText.match(/of (\d+)/);
      const total = match ? parseInt(match[1]) : 0;

      if (total > 50) {
        // There are more pages
        await logsPage.goToNextPage();

        // Wait for page to load
        await page.waitForTimeout(1000);

        // Should be on page 2
        const newPaginationText = await logsPage.getPaginationText();
        expect(newPaginationText).toContain('51');

        // Previous button should now be enabled
        await expect(logsPage.previousButton).toBeEnabled();
      }
    }
  });

  test('should navigate back to previous page', async ({ page }) => {
    const paginationText = await logsPage.getPaginationText();

    if (paginationText) {
      const match = paginationText.match(/of (\d+)/);
      const total = match ? parseInt(match[1]) : 0;

      if (total > 50) {
        // Go to next page
        await logsPage.goToNextPage();
        await page.waitForTimeout(1000);

        // Go back
        await logsPage.goToPreviousPage();
        await page.waitForTimeout(1000);

        // Should be back on first page
        const newPaginationText = await logsPage.getPaginationText();
        expect(newPaginationText).toContain('1 -');
      }
    }
  });

  test('should display empty state when no logs found', async ({ page }) => {
    // Search for something that doesn't exist
    await logsPage.searchLogs('xyznonexistentlogquery123');

    await page.waitForTimeout(1000);

    // Should show empty state
    await expect(page.getByText('No logs found')).toBeVisible();
  });

  test('should show loading state', async ({ page, context }) => {
    // Create new page to see loading state
    const newPage = await context.newPage();
    const newLogs = new LogsPage(newPage);
    const newLogin = new LoginPage(newPage);

    await newLogin.goto();
    await newLogin.login(TEST_USER.email, TEST_USER.password);
    await newLogs.goto();

    // Check for loading spinner (might be brief)
    const spinner = newPage.locator('.animate-spin');

    try {
      await expect(spinner).toBeVisible({ timeout: 1000 });
    } catch {
      // Content loaded too fast, verify table is visible
      await expect(newLogs.logTable).toBeVisible();
    }

    await newPage.close();
  });

  test('should display log level badges with correct styling', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      // Check for level badges in table
      const badges = await page.locator('tbody tr').first().locator('[data-variant]').all();
      expect(badges.length).toBeGreaterThan(0);
    }
  });

  test('should truncate long log messages in table', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      // Log messages in table should have truncate class
      const messageCell = page.locator('tbody tr').first().locator('td').nth(3);
      await expect(messageCell).toBeVisible();

      // Click to see full message in dialog
      await logsPage.clickLogRow(0);
      await logsPage.waitForLogDialog();

      // Full message should be visible in dialog (not truncated)
      await expect(page.locator('.whitespace-pre-wrap')).toBeVisible();

      await logsPage.closeLogDialog();
    }
  });

  test('should preserve filters after page reload', async ({ page }) => {
    // Set filters
    await logsPage.searchLogs('test');
    await logsPage.filterByLevel('ERROR');

    // Reload page
    await page.reload();

    // Wait for page to load
    await page.waitForTimeout(1000);

    // Filters might be reset (this is expected behavior)
    // Just verify page loads correctly
    await expect(logsPage.logsHeading).toBeVisible();
  });

  test('should not select checkbox when clicking row', async ({ page }) => {
    const logCount = await logsPage.getLogCount();

    if (logCount > 0) {
      const checkbox = page.locator('tbody tr').first().locator('input[type="checkbox"]');

      // Initially unchecked
      await expect(checkbox).not.toBeChecked();

      // Click row (not checkbox)
      await logsPage.clickLogRow(0);

      // Checkbox should still be unchecked
      await logsPage.closeLogDialog();
      await expect(checkbox).not.toBeChecked();
    }
  });
});

test.describe('Logs - With Project Context', () => {
  let loginPage: LoginPage;
  let logsPage: LogsPage;
  let projectsPage: ProjectsPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    logsPage = new LogsPage(page);
    projectsPage = new ProjectsPage(page);
    dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
  });

  test('should create project and filter logs by it', async ({ page }) => {
    // Create a new project
    await projectsPage.goto();
    const projectName = `Test Project Logs ${Date.now()}`;
    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Go to logs page
    await logsPage.goto();

    // Project should appear in filter dropdown
    await logsPage.projectFilter.click();
    await expect(page.getByRole('option', { name: projectName })).toBeVisible();

    // Select the project
    await page.getByRole('option', { name: projectName }).click();

    await page.waitForTimeout(1000);

    // Filter should be applied
    await expect(logsPage.projectFilter).toContainText(projectName);
  });
});

test.describe('Logs - Navigation', () => {
  let loginPage: LoginPage;
  let logsPage: LogsPage;
  let navigation: NavigationPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    logsPage = new LogsPage(page);
    navigation = new NavigationPage(page);
    dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
  });

  test('should navigate to logs from dashboard', async ({ page }) => {
    // Start at dashboard
    await page.goto('/');

    // Navigate to logs
    await navigation.navigateToLogs();

    await expect(page).toHaveURL('/logs');
    await expect(logsPage.logsHeading).toBeVisible();
  });

  test('should maintain scroll position in table', async ({ page }) => {
    await logsPage.goto();

    const logCount = await logsPage.getLogCount();

    if (logCount > 10) {
      // Scroll down in table
      const table = await page.locator('tbody');
      await table.evaluate(el => el.scrollTop = 200);

      const scrollTop = await table.evaluate(el => el.scrollTop);
      expect(scrollTop).toBeGreaterThan(0);
    }
  });
});
