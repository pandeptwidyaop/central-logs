/**
 * Helper functions to generate test data
 */

export function generateProjectName(prefix = 'Test Project'): string {
  return `${prefix} ${Date.now()}`;
}

export function generateProjectDescription(): string {
  return `Test project created at ${new Date().toISOString()}`;
}

export function generateUserEmail(): string {
  return `test.user.${Date.now()}@example.com`;
}

export function generateUserName(): string {
  return `Test User ${Date.now()}`;
}

/**
 * Wait helper for API calls
 */
export function waitForApiCall(ms = 500): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Retry helper for flaky operations
 */
export async function retry<T>(
  fn: () => Promise<T>,
  maxAttempts = 3,
  delay = 1000
): Promise<T> {
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (attempt === maxAttempts) {
        throw error;
      }
      await waitForApiCall(delay);
    }
  }
  throw new Error('Retry failed');
}
