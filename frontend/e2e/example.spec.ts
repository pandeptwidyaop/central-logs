/**
 * Example test file demonstrating best practices and fixture usage
 *
 * This file shows:
 * - How to use fixtures for cleaner tests
 * - How to use helpers
 * - Page object pattern
 * - Test organization
 */

import { test, expect } from './fixtures/auth.fixture';
import { TEST_USER } from './helpers/auth.helper';
import { generateProjectName } from './helpers/test-data.helper';

test.describe('Example Tests - Using Fixtures', () => {

  test('example: login and navigate using fixtures', async ({
    page,
    loginPage,
    dashboardPage,
    navigation
  }) => {
    // Login is easy with page objects from fixtures
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);

    // Wait for dashboard
    await dashboardPage.waitForDashboard();
    await expect(page).toHaveURL('/');

    // Navigate using navigation page object
    await navigation.navigateToProjects();
    await expect(page).toHaveURL('/projects');
  });

  test('example: create project with helpers', async ({
    page,
    loginPage,
    projectsPage,
    dashboardPage
  }) => {
    // Login
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Go to projects
    await projectsPage.goto();

    // Use helper to generate unique project name
    const projectName = generateProjectName('Example');

    // Create project using page object method
    await projectsPage.createProject(projectName, 'Example project using helpers');

    // Verify creation
    await expect(page.getByText('Project created successfully')).toBeVisible();
    await expect(projectsPage.getProjectCard(projectName)).toBeVisible();
  });

  test('example: search logs with filters', async ({
    page,
    loginPage,
    logsPage,
    dashboardPage
  }) => {
    // Login
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Go to logs
    await logsPage.goto();

    // Use page object methods for filtering
    await logsPage.searchLogs('error');
    await logsPage.filterByLevel('ERROR');

    // Wait for results
    await page.waitForTimeout(1000);

    // Verify table is visible
    await expect(logsPage.logTable).toBeVisible();
  });
});

test.describe('Example Tests - Traditional Style', () => {

  test('example: traditional approach without fixtures', async ({ page }) => {
    // You can still use page objects directly if preferred
    const { LoginPage } = await import('./page-objects/login.page');
    const { DashboardPage } = await import('./page-objects/dashboard.page');

    const loginPage = new LoginPage(page);
    const dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    await expect(page).toHaveURL('/');
  });
});

/**
 * This example file can be deleted after reviewing.
 * It's included to show different approaches to writing tests.
 */
