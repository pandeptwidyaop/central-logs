import { Page } from '@playwright/test';
import { LoginPage } from '../page-objects/login.page';

export const TEST_USER = {
  email: 'admin@example.com',
  password: 'changeme123',
};

/**
 * Helper function to login with default test user
 */
export async function loginAsAdmin(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.login(TEST_USER.email, TEST_USER.password);
}

/**
 * Helper function to logout
 */
export async function logout(page: Page) {
  const logoutButton = page.getByRole('button', { name: /logout/i });
  await logoutButton.click();
}

/**
 * Helper function to check if user is authenticated
 */
export async function isAuthenticated(page: Page): Promise<boolean> {
  const token = await page.evaluate(() => localStorage.getItem('token'));
  return token !== null;
}

/**
 * Helper function to clear authentication
 */
export async function clearAuth(page: Page) {
  await page.evaluate(() => localStorage.removeItem('token'));
}

/**
 * Helper function to set authentication token
 */
export async function setAuthToken(page: Page, token: string) {
  await page.evaluate((t) => localStorage.setItem('token', t), token);
}
