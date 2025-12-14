import { test, expect } from '@playwright/test';
import { LoginPage } from './page-objects/login.page';
import { DashboardPage } from './page-objects/dashboard.page';
import { NavigationPage } from './page-objects/navigation.page';

const TEST_USER = {
  email: 'admin@example.com',
  password: 'changeme123',
};

test.describe('Authentication', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;
  let navigation: NavigationPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    navigation = new NavigationPage(page);
  });

  test('should display login page', async ({ page }) => {
    await loginPage.goto();
    await expect(loginPage.centralLogsHeading).toBeVisible();
    await expect(page.getByText('Welcome back')).toBeVisible();
    await expect(loginPage.emailInput).toBeVisible();
    await expect(loginPage.passwordInput).toBeVisible();
    await expect(loginPage.signInButton).toBeVisible();
  });

  test('should show error with invalid credentials', async () => {
    await loginPage.goto();
    await loginPage.login('invalid@example.com', 'wrongpassword');

    // Wait for error message
    await expect(loginPage.errorMessage).toBeVisible();
  });

  test('should show error with empty credentials', async () => {
    await loginPage.goto();

    // HTML5 validation should prevent submission
    await expect(loginPage.signInButton).toBeVisible();
    await loginPage.signInButton.click();

    // Email should be required
    await expect(loginPage.emailInput).toBeFocused();
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);

    // Should redirect to dashboard
    await dashboardPage.waitForDashboard();
    await expect(page).toHaveURL('/');
    await expect(dashboardPage.dashboardHeading).toBeVisible();
  });

  test('should persist session after page reload', async ({ page, context: _context }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Verify token exists in localStorage
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeTruthy();

    // Reload page
    await page.reload();

    // Should still be logged in
    await dashboardPage.waitForDashboard();
    await expect(dashboardPage.dashboardHeading).toBeVisible();
    await expect(page).toHaveURL('/');
  });

  test('should persist session in new tab', async ({ context }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Open new tab
    const newPage = await context.newPage();
    const newDashboard = new DashboardPage(newPage);

    await newPage.goto('/');

    // Should be logged in automatically
    await newDashboard.waitForDashboard();
    await expect(newDashboard.dashboardHeading).toBeVisible();

    await newPage.close();
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Logout
    await navigation.logout();

    // Should redirect to login page
    await loginPage.waitForLoginPage();
    await expect(page).toHaveURL('/login');

    // Token should be removed from localStorage
    const token = await page.evaluate(() => localStorage.getItem('token'));
    expect(token).toBeNull();
  });

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await dashboardPage.goto();

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should redirect to dashboard when accessing login page while authenticated', async ({ page }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Try to go to login page
    await loginPage.goto();

    // Should redirect back to dashboard
    await expect(page).toHaveURL('/');
    await expect(dashboardPage.dashboardHeading).toBeVisible();
  });

  test('should handle session expiry gracefully', async ({ page }) => {
    // Login first
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Manually clear token to simulate expiry
    await page.evaluate(() => localStorage.removeItem('token'));

    // Try to navigate to a protected page
    await page.goto('/projects');

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should show loading state while signing in', async () => {
    await loginPage.goto();

    await loginPage.emailInput.fill(TEST_USER.email);
    await loginPage.passwordInput.fill(TEST_USER.password);

    // Click submit and check for loading state
    await loginPage.signInButton.click();

    // Button should show loading text
    await expect(loginPage.signInButton).toHaveText('Signing in...');
  });
});

test.describe('Authentication - Session Management', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
  });

  test('should maintain session across navigation', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    const navigation = new NavigationPage(page);

    // Navigate to different pages
    await navigation.navigateToProjects();
    await expect(page).toHaveURL('/projects');

    await navigation.navigateToLogs();
    await expect(page).toHaveURL('/logs');

    await navigation.navigateToDashboard();
    await expect(page).toHaveURL('/');

    // Should still be logged in after all navigation
    await expect(dashboardPage.dashboardHeading).toBeVisible();
  });

  test('should handle browser back button after logout', async ({ page }) => {
    // Login
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();

    // Logout
    const navigation = new NavigationPage(page);
    await navigation.logout();
    await loginPage.waitForLoginPage();

    // Try to go back
    await page.goBack();

    // Should still be on login page or redirect to login
    await expect(page).toHaveURL('/login');
  });
});
