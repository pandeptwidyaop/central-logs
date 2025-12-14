# E2E Test Architecture

## Directory Structure

```
frontend/e2e/
├── README.md                    # Comprehensive test documentation
├── QUICK_START.md              # Quick start guide for running tests
├── TEST_SUMMARY.md             # Summary of all tests and coverage
├── ARCHITECTURE.md             # This file - architecture overview
│
├── page-objects/               # Page Object Model (POM) classes
│   ├── login.page.ts           # Login page interactions
│   ├── dashboard.page.ts       # Dashboard page interactions
│   ├── projects.page.ts        # Projects page interactions
│   ├── logs.page.ts            # Logs page interactions
│   └── navigation.page.ts      # Navigation/sidebar interactions
│
├── helpers/                    # Reusable test utilities
│   ├── auth.helper.ts          # Authentication helpers
│   └── test-data.helper.ts     # Test data generation helpers
│
├── fixtures/                   # Playwright fixtures
│   └── auth.fixture.ts         # Page object fixtures for DI
│
├── auth.spec.ts                # Authentication test suite (13 tests)
├── dashboard.spec.ts           # Dashboard test suite (15 tests)
├── projects.spec.ts            # Projects test suite (22 tests)
├── logs.spec.ts                # Logs test suite (28 tests)
└── example.spec.ts             # Example tests showing best practices
```

## Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│                     Test Specifications                      │
│  (auth.spec.ts, dashboard.spec.ts, projects.spec.ts, etc.)  │
│                                                              │
│  - Test scenarios and assertions                            │
│  - Business logic validation                                │
│  - User flow testing                                        │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                     Fixtures Layer                           │
│                  (auth.fixture.ts)                           │
│                                                              │
│  - Dependency injection                                     │
│  - Shared page object instances                            │
│  - Test setup/teardown                                     │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                   Page Objects Layer                         │
│         (login.page.ts, dashboard.page.ts, etc.)            │
│                                                              │
│  - Encapsulate page-specific logic                         │
│  - Define locators and actions                             │
│  - Provide high-level methods                              │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                     Helpers Layer                            │
│         (auth.helper.ts, test-data.helper.ts)               │
│                                                              │
│  - Reusable utility functions                              │
│  - Test data generation                                    │
│  - Common operations                                       │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                  Playwright Core                             │
│              (Page, Browser, Context)                        │
│                                                              │
│  - Browser automation                                       │
│  - Element interaction                                      │
│  - Assertions                                              │
└─────────────────────────────────────────────────────────────┘
```

## Test Flow Example

### Traditional Approach
```
Test File
    ↓
Import Page Objects
    ↓
Instantiate Page Objects
    ↓
Call Page Object Methods
    ↓
Assertions
```

### Using Fixtures (Recommended)
```
Test File with Fixtures
    ↓
Fixtures Auto-inject Page Objects
    ↓
Use Pre-instantiated Objects
    ↓
Call Page Object Methods
    ↓
Assertions
```

## Component Interaction Diagram

```
┌────────────────┐
│   auth.spec.ts │
│   ┌──────────┐ │
│   │ Test 1   │ │──┐
│   └──────────┘ │  │
│   ┌──────────┐ │  │
│   │ Test 2   │ │──┤
│   └──────────┘ │  │
└────────────────┘  │
                    │
                    ▼
┌─────────────────────────────┐
│   auth.fixture.ts           │
│   ┌───────────────────────┐ │
│   │ loginPage fixture     │ │
│   │ dashboardPage fixture │ │
│   │ navigation fixture    │ │
│   └───────────────────────┘ │
└─────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────┐
│   login.page.ts             │
│   ┌───────────────────────┐ │
│   │ emailInput: Locator   │ │
│   │ passwordInput: Locator│ │
│   │ signInButton: Locator │ │
│   │                       │ │
│   │ login(email, pass)    │ │──┐
│   │ goto()                │ │  │
│   └───────────────────────┘ │  │
└─────────────────────────────┘  │
                                 │
                                 ▼
┌─────────────────────────────────────┐
│   Application Under Test            │
│   http://localhost:5173              │
│   ┌───────────────────────────────┐ │
│   │ Login Page                    │ │
│   │ - Email Input                 │ │
│   │ - Password Input              │ │
│   │ - Sign In Button              │ │
│   └───────────────────────────────┘ │
└─────────────────────────────────────┘
```

## Page Object Pattern Details

### LoginPage Example
```typescript
class LoginPage {
  // Locators (what)
  readonly emailInput: Locator;
  readonly passwordInput: Locator;
  readonly signInButton: Locator;

  // Constructor
  constructor(page: Page) {
    this.emailInput = page.locator('#email');
    this.passwordInput = page.locator('#password');
    this.signInButton = page.getByRole('button', { name: /sign in/i });
  }

  // Actions (how)
  async login(email: string, password: string) {
    await this.emailInput.fill(email);
    await this.passwordInput.fill(password);
    await this.signInButton.click();
  }

  async goto() {
    await this.page.goto('/login');
  }
}
```

### Test Using Page Object
```typescript
test('should login successfully', async ({ page }) => {
  const loginPage = new LoginPage(page);

  // High-level, readable actions
  await loginPage.goto();
  await loginPage.login('admin@example.com', 'changeme123');

  // Assertions
  await expect(page).toHaveURL('/');
});
```

## Test Data Flow

```
Test Spec
    │
    │ needs unique project name
    │
    ▼
┌─────────────────────────┐
│ test-data.helper.ts     │
│                         │
│ generateProjectName()   │──────► "Test Project 1702345678123"
│   └─ Uses Date.now()    │
└─────────────────────────┘
    │
    │
    ▼
Page Object Method
    │
    │ createProject(name)
    │
    ▼
Application API
    │
    │ POST /api/admin/projects
    │
    ▼
Backend Database
```

## Authentication Flow

```
┌─────────────┐
│ Test Start  │
└──────┬──────┘
       │
       ▼
┌────────────────────────┐
│ loginAsAdmin(page)     │◄─── auth.helper.ts
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ LoginPage.login()      │◄─── login.page.ts
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Fill email/password    │
│ Click sign in          │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ POST /api/auth/login   │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Receive JWT token      │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Store in localStorage  │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Redirect to dashboard  │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Test assertions        │
└────────────────────────┘
```

## Test Execution Flow

```
Developer runs: npx playwright test
           │
           ▼
┌────────────────────────┐
│ Playwright Config      │
│ - Load test files      │
│ - Start web server     │
│ - Setup browser        │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ For each test file:    │
│ - auth.spec.ts         │
│ - dashboard.spec.ts    │
│ - projects.spec.ts     │
│ - logs.spec.ts         │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ For each test:         │
│ - Setup fixtures       │
│ - Run test             │
│ - Assertions           │
│ - Cleanup              │
└──────┬─────────────────┘
       │
       ▼
┌────────────────────────┐
│ Generate report:       │
│ - HTML report          │
│ - Screenshots (fails)  │
│ - Traces (retries)     │
└────────────────────────┘
```

## Locator Strategy

### Priority Order (Recommended by Playwright)

1. **User-facing attributes** (Best)
   ```typescript
   page.getByRole('button', { name: 'Sign in' })
   page.getByLabel('Email')
   page.getByPlaceholder('Search logs...')
   page.getByText('Dashboard')
   ```

2. **Test IDs** (Good for dynamic content)
   ```typescript
   page.getByTestId('submit-button')
   ```

3. **CSS/XPath** (Last resort)
   ```typescript
   page.locator('#email')
   page.locator('.submit-btn')
   ```

### Our Implementation
- Primary: Role-based selectors (`getByRole`, `getByLabel`)
- Secondary: Text content (`getByText`)
- Fallback: ID selectors (`#email`, `#password`)
- Avoid: Complex CSS selectors

## Test Independence

Each test follows the **AAA Pattern**:

```
┌─────────────────┐
│ Arrange         │  Setup: Login, navigate, create test data
├─────────────────┤
│ Act             │  Action: Perform the operation being tested
├─────────────────┤
│ Assert          │  Verify: Check expected outcomes
└─────────────────┘
```

**Example:**
```typescript
test('should create project', async ({ page }) => {
  // Arrange
  await loginPage.goto();
  await loginPage.login(email, password);
  await projectsPage.goto();
  const projectName = generateProjectName();

  // Act
  await projectsPage.createProject(projectName);

  // Assert
  await expect(projectsPage.getProjectCard(projectName)).toBeVisible();
});
```

## Error Handling & Retries

```
Test Execution
    │
    ├─ Success ──────────► Pass
    │
    └─ Failure
         │
         ├─ CI Mode?
         │   │
         │   ├─ Yes ──► Retry (up to 2 times)
         │   │            │
         │   │            ├─ Success ──► Pass (with warning)
         │   │            └─ Fail ─────► Fail (with trace)
         │   │
         │   └─ No ───► Fail immediately
```

## Configuration

```
playwright.config.ts
    │
    ├── testDir: './e2e'
    ├── baseURL: 'http://localhost:5173'
    ├── retries: CI ? 2 : 0
    ├── workers: CI ? 1 : undefined
    ├── reporter: 'html'
    │
    └── webServer
        ├── command: 'npm run dev'
        ├── url: 'http://localhost:5173'
        └── timeout: 120s
```

## Best Practices Applied

1. **Separation of Concerns**
   - Tests: Define scenarios
   - Page Objects: Define interactions
   - Helpers: Define utilities

2. **DRY (Don't Repeat Yourself)**
   - Reusable page objects
   - Shared helpers
   - Fixtures for common setup

3. **Single Responsibility**
   - Each page object represents one page
   - Each helper has one purpose
   - Each test validates one scenario

4. **Maintainability**
   - When UI changes, update page objects only
   - Tests remain stable
   - Clear, descriptive names

5. **Readability**
   - Tests read like user stories
   - Clear arrange-act-assert structure
   - Meaningful variable names

## Scaling Strategy

As the application grows:

1. **Add New Page Objects**
   - Settings page
   - User management page
   - Channel configuration page

2. **Add New Test Suites**
   - settings.spec.ts
   - users.spec.ts
   - channels.spec.ts

3. **Add New Helpers**
   - API helpers for direct API testing
   - Mock data helpers
   - Database helpers

4. **Add New Fixtures**
   - API client fixture
   - Mock server fixture
   - Test data fixture

## Performance Considerations

- Tests run in parallel locally (faster)
- Tests run serially in CI (more stable)
- Auto-waiting reduces flakiness
- Strategic use of explicit waits
- Reuse browser contexts where possible

## Maintenance Checklist

- [ ] Update page objects when UI changes
- [ ] Add tests for new features
- [ ] Review and fix flaky tests
- [ ] Update documentation
- [ ] Check test coverage
- [ ] Optimize slow tests
- [ ] Update dependencies
- [ ] Review test results in CI
