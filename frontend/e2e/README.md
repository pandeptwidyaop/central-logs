# E2E Tests for Central Logs Frontend

This directory contains end-to-end tests for the Central Logs React frontend application using Playwright.

## Structure

```
e2e/
├── page-objects/          # Page Object Model implementations
│   ├── login.page.ts      # Login page interactions
│   ├── dashboard.page.ts  # Dashboard page interactions
│   ├── projects.page.ts   # Projects page interactions
│   ├── logs.page.ts       # Logs page interactions
│   └── navigation.page.ts # Navigation/sidebar interactions
├── helpers/               # Test helper functions
│   ├── auth.helper.ts     # Authentication utilities
│   └── test-data.helper.ts # Test data generation
├── auth.spec.ts           # Authentication tests
├── dashboard.spec.ts      # Dashboard tests
├── projects.spec.ts       # Projects tests
└── logs.spec.ts           # Logs tests
```

## Test Coverage

### Authentication Tests (`auth.spec.ts`)
- Login page display
- Invalid credentials handling
- Successful login flow
- Session persistence across page reloads
- Session persistence across tabs
- Logout functionality
- Protected route access control
- Session expiry handling
- Loading states

### Dashboard Tests (`dashboard.spec.ts`)
- Dashboard page display
- Stat cards rendering
- Logs by level visualization
- Recent logs display
- Navigation to logs page
- Responsive design
- Loading states

### Projects Tests (`projects.spec.ts`)
- Projects page display
- Create project dialog
- Project creation (with/without description)
- Project editing
- Project deletion
- API key rotation
- API key copying
- Empty state handling
- Multiple projects management
- Form validation

### Logs Tests (`logs.spec.ts`)
- Logs page display
- Search functionality
- Filter by level
- Filter by project
- Multiple filter combinations
- Log selection (individual and all)
- Log details dialog
- Pagination
- Refresh functionality
- Empty state
- Loading states

## Running Tests

### Prerequisites

1. Ensure the backend is running and accessible
2. Ensure the frontend dev server is running on `http://localhost:5173`
3. Default admin credentials are available: `admin@example.com` / `changeme123`

### Commands

```bash
# Install dependencies (if not already installed)
npm install

# Run all tests
npm run test:e2e

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run specific test file
npx playwright test auth.spec.ts

# Run tests in debug mode
npx playwright test --debug

# Run tests in specific browser
npx playwright test --project=chromium

# Generate test report
npx playwright show-report
```

### Interactive Mode

```bash
# Open Playwright UI for interactive testing
npx playwright test --ui
```

## Configuration

Tests are configured in `/Users/pande/Projects/central-logs/frontend/playwright.config.ts`:

- Base URL: `http://localhost:5173`
- Test directory: `./e2e`
- Browser: Chromium (desktop)
- Screenshots: On failure
- Traces: On first retry
- Web server auto-start: `npm run dev`

## Page Object Pattern

Tests use the Page Object Model (POM) pattern for better maintainability and reusability:

- **Page Objects** encapsulate page-specific locators and interactions
- **Test files** focus on test scenarios and assertions
- **Helpers** provide reusable utilities

Example:
```typescript
// Page Object
export class LoginPage {
  readonly emailInput: Locator;
  readonly passwordInput: Locator;

  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.signInButton.click();
  }
}

// Test
test('should login successfully', async ({ page }) => {
  const loginPage = new LoginPage(page);
  await loginPage.login('admin@example.com', 'changeme123');
  await expect(page).toHaveURL('/');
});
```

## Test Independence

All tests are designed to be independent:

- Each test can run in isolation
- Tests don't rely on execution order
- Test data is generated uniquely (using timestamps)
- Authentication state is established per test
- Cleanup is handled appropriately

## Best Practices

1. **Use Page Objects**: Encapsulate page interactions in page objects
2. **Unique Test Data**: Generate unique names/emails to avoid conflicts
3. **Wait Strategies**: Use Playwright's auto-waiting, add explicit waits only when needed
4. **Assertions**: Use meaningful assertions with clear failure messages
5. **Selectors**: Prefer user-facing selectors (roles, labels) over CSS classes
6. **Clean Up**: Handle test data cleanup (especially for delete operations)
7. **Screenshots**: Automatic on failure, helps with debugging

## Troubleshooting

### Tests Failing Intermittently
- Check for race conditions
- Increase timeout for slow operations
- Verify backend is running and responding

### Element Not Found
- Check if selectors have changed
- Verify page has fully loaded
- Check for dynamic content loading

### Authentication Issues
- Verify credentials are correct in config
- Check backend API is accessible
- Clear browser storage between tests if needed

### Debugging Tips
1. Run tests in headed mode: `npx playwright test --headed`
2. Use debug mode: `npx playwright test --debug`
3. Add `await page.pause()` in test to pause execution
4. Check screenshots in `test-results/` folder
5. View traces in Playwright report for failed tests

## CI/CD Integration

Tests are configured for CI/CD with:
- 2 retries on failure
- 1 worker (serial execution)
- HTML reporter
- Automatic web server startup

## Contributing

When adding new tests:

1. Create appropriate page objects in `page-objects/`
2. Use existing helpers or create new ones in `helpers/`
3. Follow existing naming conventions
4. Ensure tests are independent
5. Add meaningful test descriptions
6. Group related tests with `test.describe()`
7. Update this README if adding new test files
