# Quick Start Guide - E2E Tests

## Setup

### 1. Install Dependencies
```bash
cd /Users/pande/Projects/central-logs/frontend
npm install
```

### 2. Install Playwright Browsers (First Time Only)
```bash
npx playwright install
```

### 3. Start Backend
```bash
# In the project root
cd /Users/pande/Projects/central-logs
make run
# OR
go run cmd/server/main.go
```

### 4. Verify Backend is Running
The backend should be running on `http://localhost:8080`
The frontend will proxy `/api` requests to the backend.

## Running Tests

### Option 1: Let Playwright Start the Dev Server (Recommended)
The Playwright config is set to automatically start the dev server on port 5173.

```bash
# Run all tests (web server starts automatically)
npx playwright test

# Run in headed mode (see browser)
npx playwright test --headed

# Run specific test file
npx playwright test auth.spec.ts
npx playwright test dashboard.spec.ts
npx playwright test projects.spec.ts
npx playwright test logs.spec.ts

# Run in UI mode (interactive)
npx playwright test --ui

# Run in debug mode
npx playwright test --debug
```

### Option 2: Manually Start Dev Server
```bash
# Terminal 1: Start frontend dev server
npm run dev

# Terminal 2: Run tests
npx playwright test
```

## Test Files Overview

| Test File | Description | Test Count |
|-----------|-------------|------------|
| `auth.spec.ts` | Authentication flows, login, logout, session persistence | 13 tests |
| `dashboard.spec.ts` | Dashboard stats, charts, recent logs | 15 tests |
| `projects.spec.ts` | Create, edit, delete projects, API key management | 20+ tests |
| `logs.spec.ts` | View logs, search, filter, pagination | 25+ tests |

**Total: 70+ comprehensive E2E tests**

## Viewing Results

### HTML Report
```bash
# Generate and open HTML report
npx playwright show-report
```

### Screenshots and Traces
- Screenshots on failure: `test-results/`
- Traces on retry: `test-results/`

## Running Specific Tests

### By Test Name
```bash
# Run tests matching a pattern
npx playwright test -g "should login"
npx playwright test -g "should create project"
```

### By File and Browser
```bash
# Run specific file in specific browser
npx playwright test auth.spec.ts --project=chromium
```

### Parallel vs Serial
```bash
# Run in parallel (default in local, faster)
npx playwright test --workers=4

# Run serially (CI mode)
npx playwright test --workers=1
```

## Common Commands

```bash
# Run all tests
npx playwright test

# Run with UI
npx playwright test --ui

# Run in headed mode (see browser)
npx playwright test --headed

# Debug mode (step through)
npx playwright test --debug

# Run specific file
npx playwright test auth.spec.ts

# Run specific test by name
npx playwright test -g "should login successfully"

# Update snapshots (if using visual regression)
npx playwright test --update-snapshots

# Show report
npx playwright show-report

# Run and record video
npx playwright test --video=on

# Run with specific timeout
npx playwright test --timeout=60000
```

## Debugging Tests

### Method 1: Playwright Inspector
```bash
npx playwright test --debug
```
- Step through tests
- Inspect elements
- View console logs
- Record actions

### Method 2: Pause in Test
Add to your test:
```typescript
await page.pause();
```

### Method 3: Headed Mode + Slow Motion
```bash
npx playwright test --headed --slow-mo=1000
```

### Method 4: Screenshots
```typescript
await page.screenshot({ path: 'screenshot.png' });
```

## Test Data

### Default Test User
- Email: `admin@example.com`
- Password: `changeme123`

These credentials are defined in:
- Backend: `/Users/pande/Projects/central-logs/config.yaml`
- Tests: `e2e/helpers/auth.helper.ts`

### Test Projects
Tests create unique project names using timestamps to avoid conflicts:
```typescript
const projectName = `Test Project ${Date.now()}`;
```

## Environment Variables

You can customize the test environment:

```bash
# Custom base URL
BASE_URL=http://localhost:3000 npx playwright test

# Run in CI mode
CI=true npx playwright test
```

## Troubleshooting

### Tests Fail to Start
1. Check backend is running: `curl http://localhost:8080/api/health`
2. Check frontend port 5173 is available
3. Clear browser data: `npx playwright test --global-teardown`

### Login Tests Fail
1. Verify credentials in `config.yaml`
2. Check backend database is seeded with admin user
3. Verify API endpoint `/api/auth/login` works

### Timeout Errors
1. Increase timeout in `playwright.config.ts`
2. Check backend performance
3. Use `--timeout` flag: `npx playwright test --timeout=60000`

### Flaky Tests
1. Add explicit waits where needed
2. Check for race conditions
3. Run with `--retries=3`

## CI/CD Integration

The tests are configured for CI in `playwright.config.ts`:

```typescript
retries: process.env.CI ? 2 : 0,
workers: process.env.CI ? 1 : undefined,
```

### GitHub Actions Example
```yaml
- name: Run E2E Tests
  run: |
    cd frontend
    npx playwright test
  env:
    CI: true
```

## Best Practices

1. Run tests before pushing code
2. Keep tests independent
3. Use unique test data (timestamps)
4. Clean up created resources
5. Check test reports for failures
6. Update page objects when UI changes

## Next Steps

1. Add more test scenarios as needed
2. Integrate with CI/CD pipeline
3. Add visual regression tests
4. Add performance tests
5. Add accessibility tests

## Support

For issues or questions:
1. Check test output and screenshots
2. Review page objects for selector changes
3. Verify backend API is functioning
4. Check Playwright documentation: https://playwright.dev
