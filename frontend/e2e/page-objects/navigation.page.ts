import { Page, Locator } from '@playwright/test';

export class NavigationPage {
  readonly page: Page;
  readonly dashboardLink: Locator;
  readonly projectsLink: Locator;
  readonly logsLink: Locator;
  readonly usersLink: Locator;
  readonly settingsLink: Locator;
  readonly logoutButton: Locator;
  readonly userMenu: Locator;

  constructor(page: Page) {
    this.page = page;
    this.dashboardLink = page.getByRole('link', { name: 'Dashboard', exact: true });
    this.projectsLink = page.getByRole('link', { name: 'Projects', exact: true });
    this.logsLink = page.getByRole('link', { name: 'Logs', exact: true });
    this.usersLink = page.getByRole('link', { name: 'Users', exact: true });
    this.settingsLink = page.getByRole('link', { name: 'Settings', exact: true });
    this.logoutButton = page.getByRole('button', { name: /logout/i });
    this.userMenu = page.locator('[role="button"]').filter({ hasText: /@/ });
  }

  async navigateToDashboard() {
    await this.dashboardLink.click();
  }

  async navigateToProjects() {
    await this.projectsLink.click();
  }

  async navigateToLogs() {
    await this.logsLink.click();
  }

  async navigateToUsers() {
    await this.usersLink.click();
  }

  async navigateToSettings() {
    await this.settingsLink.click();
  }

  async logout() {
    await this.logoutButton.click();
  }
}
