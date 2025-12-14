import { Page, Locator } from '@playwright/test';

export class DashboardPage {
  readonly page: Page;
  readonly dashboardHeading: Locator;
  readonly totalProjectsCard: Locator;
  readonly totalLogsCard: Locator;
  readonly logsTodayCard: Locator;
  readonly errorsTodayCard: Locator;
  readonly logsByLevelCard: Locator;
  readonly recentLogsCard: Locator;
  readonly viewAllLogsLink: Locator;

  constructor(page: Page) {
    this.page = page;
    this.dashboardHeading = page.getByRole('heading', { name: 'Dashboard' });
    this.totalProjectsCard = page.getByText('Total Projects').locator('..');
    this.totalLogsCard = page.getByText('Total Logs').first().locator('..');
    this.logsTodayCard = page.getByText('Logs Today').locator('..');
    this.errorsTodayCard = page.getByText('Errors Today').locator('..');
    this.logsByLevelCard = page.getByRole('heading', { name: 'Logs by Level' });
    this.recentLogsCard = page.getByRole('heading', { name: 'Recent Logs' });
    this.viewAllLogsLink = page.getByRole('link', { name: 'View all' });
  }

  async goto() {
    await this.page.goto('/');
  }

  async waitForDashboard() {
    await this.dashboardHeading.waitFor();
  }

  async getStatValue(cardLocator: Locator): Promise<string> {
    return await cardLocator.locator('.text-2xl').textContent() || '';
  }

  async getLevelBadges() {
    return await this.page.locator('[data-variant]').all();
  }
}
