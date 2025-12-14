import { Page, Locator } from '@playwright/test';

export class ProjectsPage {
  readonly page: Page;
  readonly projectsHeading: Locator;
  readonly newProjectButton: Locator;
  readonly createDialogTitle: Locator;
  readonly nameInput: Locator;
  readonly descriptionInput: Locator;
  readonly createButton: Locator;
  readonly cancelButton: Locator;
  readonly editNameInput: Locator;
  readonly editDescriptionInput: Locator;
  readonly saveButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.projectsHeading = page.getByRole('heading', { name: 'Projects' });
    this.newProjectButton = page.getByRole('button', { name: /new project/i });
    this.createDialogTitle = page.getByRole('heading', { name: 'Create Project' });
    this.nameInput = page.locator('#name');
    this.descriptionInput = page.locator('#description');
    this.createButton = page.getByRole('button', { name: 'Create' });
    this.cancelButton = page.getByRole('button', { name: 'Cancel' });
    this.editNameInput = page.locator('#edit-name');
    this.editDescriptionInput = page.locator('#edit-description');
    this.saveButton = page.getByRole('button', { name: 'Save' });
  }

  async goto() {
    await this.page.goto('/projects');
    await this.waitForProjectsPage();
  }

  async waitForProjectsPage() {
    await this.projectsHeading.waitFor();
  }

  async openCreateDialog() {
    await this.newProjectButton.click();
    await this.createDialogTitle.waitFor();
  }

  async createProject(name: string, description?: string) {
    await this.openCreateDialog();
    await this.nameInput.fill(name);
    if (description) {
      await this.descriptionInput.fill(description);
    }
    await this.createButton.click();
  }

  async getProjectCard(projectName: string) {
    return this.page.getByRole('link', { name: projectName });
  }

  async openProjectMenu(projectName: string) {
    const projectCard = await this.getProjectCard(projectName);
    const card = projectCard.locator('../..');
    await card.getByRole('button').first().click();
  }

  async editProject(oldName: string, newName: string, newDescription?: string) {
    await this.openProjectMenu(oldName);
    await this.page.getByRole('menuitem', { name: /edit/i }).click();
    await this.editNameInput.fill(newName);
    if (newDescription !== undefined) {
      await this.editDescriptionInput.fill(newDescription);
    }
    await this.saveButton.click();
  }

  async deleteProject(projectName: string) {
    await this.openProjectMenu(projectName);

    // Set up dialog handler before clicking delete
    this.page.once('dialog', async dialog => {
      await dialog.accept();
    });

    await this.page.getByRole('menuitem', { name: /delete/i }).click();
  }

  async rotateApiKey(projectName: string) {
    await this.openProjectMenu(projectName);
    await this.page.getByRole('menuitem', { name: /rotate api key/i }).click();
  }

  async copyApiKey(projectName: string) {
    const projectCard = await this.getProjectCard(projectName);
    const card = projectCard.locator('../..');
    const copyButton = card.getByRole('button').last();
    await copyButton.click();
  }

  async getProjectCount(): Promise<number> {
    const projectCards = await this.page.locator('[class*="grid"] > div[class*="rounded"]').all();
    return projectCards.length;
  }
}
