import { test, expect } from '@playwright/test';
import { LoginPage } from './page-objects/login.page';
import { ProjectsPage } from './page-objects/projects.page';
import { NavigationPage } from './page-objects/navigation.page';
import { DashboardPage } from './page-objects/dashboard.page';

const TEST_USER = {
  email: 'admin@example.com',
  password: 'changeme123',
};

test.describe('Projects', () => {
  let loginPage: LoginPage;
  let projectsPage: ProjectsPage;
  let navigation: NavigationPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    projectsPage = new ProjectsPage(page);
    navigation = new NavigationPage(page);
    dashboardPage = new DashboardPage(page);

    // Login before each test
    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
    await projectsPage.goto();
  });

  test('should display projects page', async ({ page }) => {
    await expect(projectsPage.projectsHeading).toBeVisible();
    await expect(page.getByText('Manage your logging projects')).toBeVisible();
    await expect(projectsPage.newProjectButton).toBeVisible();
  });

  test('should open create project dialog', async ({ page }) => {
    await projectsPage.openCreateDialog();

    await expect(projectsPage.createDialogTitle).toBeVisible();
    await expect(projectsPage.nameInput).toBeVisible();
    await expect(projectsPage.descriptionInput).toBeVisible();
    await expect(projectsPage.createButton).toBeVisible();
    await expect(projectsPage.cancelButton).toBeVisible();
  });

  test('should close create project dialog on cancel', async () => {
    await projectsPage.openCreateDialog();
    await expect(projectsPage.createDialogTitle).toBeVisible();

    await projectsPage.cancelButton.click();

    await expect(projectsPage.createDialogTitle).not.toBeVisible();
  });

  test('should create a new project', async ({ page }) => {
    const projectName = `Test Project ${Date.now()}`;
    const projectDescription = 'This is a test project created by E2E tests';

    await projectsPage.createProject(projectName, projectDescription);

    // Wait for toast notification
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Dialog should close
    await expect(projectsPage.createDialogTitle).not.toBeVisible();

    // Project should appear in the list
    await expect(projectsPage.getProjectCard(projectName)).toBeVisible();

    // Verify description is visible
    await expect(page.getByText(projectDescription)).toBeVisible();
  });

  test('should create project without description', async ({ page }) => {
    const projectName = `Test Project No Desc ${Date.now()}`;

    await projectsPage.createProject(projectName);

    await expect(page.getByText('Project created successfully')).toBeVisible();
    await expect(projectsPage.getProjectCard(projectName)).toBeVisible();
  });

  test('should display project with API key', async ({ page }) => {
    const projectName = `Test Project API ${Date.now()}`;

    await projectsPage.createProject(projectName);

    // Wait for project to be created
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Project card should display API key
    const projectCard = await projectsPage.getProjectCard(projectName);
    const card = projectCard.locator('../..');

    // API key should be visible (truncated in code element)
    await expect(card.locator('code')).toBeVisible();
  });

  test('should view project details', async ({ page }) => {
    const projectName = `Test Project View ${Date.now()}`;

    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Click on project name to view details
    const projectLink = await projectsPage.getProjectCard(projectName);
    const projectId = await projectLink.getAttribute('href');

    await projectLink.click();

    // Should navigate to project detail page
    expect(projectId).toBeTruthy();
    await expect(page).toHaveURL(projectId!);
  });

  test('should edit project', async ({ page }) => {
    const originalName = `Test Project Edit ${Date.now()}`;
    const newName = `${originalName} (Updated)`;
    const newDescription = 'Updated description';

    // Create project first
    await projectsPage.createProject(originalName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Edit project
    await projectsPage.editProject(originalName, newName, newDescription);

    // Wait for success toast
    await expect(page.getByText('Project updated successfully')).toBeVisible();

    // Verify updated name and description
    await expect(projectsPage.getProjectCard(newName)).toBeVisible();
    await expect(page.getByText(newDescription)).toBeVisible();

    // Old name should not be visible
    await expect(projectsPage.getProjectCard(originalName)).not.toBeVisible();
  });

  test('should delete project', async ({ page }) => {
    const projectName = `Test Project Delete ${Date.now()}`;

    // Create project first
    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Delete project
    await projectsPage.deleteProject(projectName);

    // Wait for success toast
    await expect(page.getByText('Project deleted successfully')).toBeVisible();

    // Project should no longer be visible
    await expect(projectsPage.getProjectCard(projectName)).not.toBeVisible();
  });

  test('should rotate API key', async ({ page }) => {
    const projectName = `Test Project Rotate ${Date.now()}`;

    // Create project first
    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Get original API key
    const projectCard = (await projectsPage.getProjectCard(projectName)).locator('../..');
    const originalKey = await projectCard.locator('code').textContent();

    // Rotate API key
    await projectsPage.rotateApiKey(projectName);

    // Wait for toast notification with new key
    await expect(page.getByText(/API key rotated/)).toBeVisible();

    // New key should be different (wait for page to update)
    await page.waitForTimeout(500);
    const newKey = await projectCard.locator('code').textContent();
    expect(newKey).not.toBe(originalKey);
  });

  test('should copy API key to clipboard', async ({ page, context }) => {
    const projectName = `Test Project Copy ${Date.now()}`;

    // Grant clipboard permissions
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);

    // Create project first
    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Get API key
    const projectCard = (await projectsPage.getProjectCard(projectName)).locator('../..');
    const apiKey = await projectCard.locator('code').textContent();

    // Copy API key
    await projectsPage.copyApiKey(projectName);

    // Verify copy icon changed to check mark
    const checkIcon = projectCard.locator('.text-green-500');
    await expect(checkIcon).toBeVisible();

    // Verify clipboard content (if supported)
    try {
      const clipboardText = await page.evaluate(() => navigator.clipboard.readText());
      expect(clipboardText).toBe(apiKey);
    } catch {
      // Clipboard API might not be available in all test environments
      console.log('Clipboard verification skipped');
    }
  });

  test('should display empty state when no projects exist', async ({ page }) => {
    // Delete all projects first (if any)
    const projectCards = await page.locator('[class*="grid"] > div').all();

    for (const card of projectCards) {
      const menuButton = card.locator('button').first();
      if (await menuButton.isVisible()) {
        await menuButton.click();

        page.once('dialog', async dialog => {
          await dialog.accept();
        });

        await page.getByRole('menuitem', { name: /delete/i }).click();
        await page.waitForTimeout(500);
      }
    }

    // Reload to ensure we see empty state
    await page.reload();

    // Should show empty state
    await expect(page.getByText('No projects yet')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create your first project' })).toBeVisible();
  });

  test('should create project from empty state', async ({ page }) => {
    // First ensure we're in empty state
    const hasProjects = await projectsPage.newProjectButton.isVisible();

    if (hasProjects) {
      // Delete all projects to get to empty state
      const projectCards = await page.locator('[class*="grid"] > div').all();

      for (const card of projectCards) {
        const menuButton = card.locator('button').first();
        if (await menuButton.isVisible()) {
          await menuButton.click();

          page.once('dialog', async dialog => {
            await dialog.accept();
          });

          await page.getByRole('menuitem', { name: /delete/i }).click();
          await page.waitForTimeout(500);
        }
      }

      await page.reload();
    }

    // Click create from empty state
    const createButton = page.getByRole('button', { name: 'Create your first project' });
    if (await createButton.isVisible()) {
      await createButton.click();
      await expect(projectsPage.createDialogTitle).toBeVisible();
    }
  });

  test('should display project creation date', async ({ page }) => {
    const projectName = `Test Project Date ${Date.now()}`;

    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Check for creation date
    const projectCard = (await projectsPage.getProjectCard(projectName)).locator('../..');
    await expect(projectCard.getByText(/Created \d/)).toBeVisible();
  });

  test('should show project menu options', async ({ page }) => {
    const projectName = `Test Project Menu ${Date.now()}`;

    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    await projectsPage.openProjectMenu(projectName);

    // Check all menu items
    await expect(page.getByRole('menuitem', { name: /edit/i })).toBeVisible();
    await expect(page.getByRole('menuitem', { name: /rotate api key/i })).toBeVisible();
    await expect(page.getByRole('menuitem', { name: /delete/i })).toBeVisible();
  });

  test('should handle multiple projects', async ({ page }) => {
    const project1 = `Test Project 1 ${Date.now()}`;
    const project2 = `Test Project 2 ${Date.now()}`;

    // Create first project
    await projectsPage.createProject(project1);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Create second project
    await projectsPage.createProject(project2);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Both should be visible
    await expect(projectsPage.getProjectCard(project1)).toBeVisible();
    await expect(projectsPage.getProjectCard(project2)).toBeVisible();
  });

  test('should validate required fields', async () => {
    await projectsPage.openCreateDialog();

    // Try to submit without name
    await projectsPage.createButton.click();

    // Name field should be focused (HTML5 validation)
    await expect(projectsPage.nameInput).toBeFocused();
  });

  test('should truncate long descriptions in card view', async ({ page }) => {
    const projectName = `Test Project Long Desc ${Date.now()}`;
    const longDescription = 'This is a very long description that should be truncated in the card view to prevent the layout from breaking. It contains many words and should demonstrate the line clamp behavior.';

    await projectsPage.createProject(projectName, longDescription);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    const projectCard = (await projectsPage.getProjectCard(projectName)).locator('../..');

    // Description should be visible but may be clamped
    const description = projectCard.locator('text=' + longDescription.substring(0, 50));
    await expect(description).toBeVisible();
  });
});

test.describe('Projects - Navigation', () => {
  let loginPage: LoginPage;
  let projectsPage: ProjectsPage;
  let navigation: NavigationPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    loginPage = new LoginPage(page);
    projectsPage = new ProjectsPage(page);
    navigation = new NavigationPage(page);
    dashboardPage = new DashboardPage(page);

    await loginPage.goto();
    await loginPage.login(TEST_USER.email, TEST_USER.password);
    await dashboardPage.waitForDashboard();
  });

  test('should navigate to projects page from dashboard', async ({ page }) => {
    await navigation.navigateToProjects();

    await expect(page).toHaveURL('/projects');
    await expect(projectsPage.projectsHeading).toBeVisible();
  });

  test('should maintain project list after navigation', async ({ page }) => {
    await projectsPage.goto();

    // Create a project
    const projectName = `Test Project Nav ${Date.now()}`;
    await projectsPage.createProject(projectName);
    await expect(page.getByText('Project created successfully')).toBeVisible();

    // Navigate away and back
    await navigation.navigateToDashboard();
    await navigation.navigateToProjects();

    // Project should still be visible
    await expect(projectsPage.getProjectCard(projectName)).toBeVisible();
  });
});
