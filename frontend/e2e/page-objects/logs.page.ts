import { Page, Locator } from '@playwright/test';

export class LogsPage {
  readonly page: Page;
  readonly logsHeading: Locator;
  readonly searchInput: Locator;
  readonly projectFilter: Locator;
  readonly levelFilter: Locator;
  readonly refreshButton: Locator;
  readonly deleteButton: Locator;
  readonly selectAllCheckbox: Locator;
  readonly previousButton: Locator;
  readonly nextButton: Locator;
  readonly logTable: Locator;

  constructor(page: Page) {
    this.page = page;
    this.logsHeading = page.getByRole('heading', { name: 'Logs' });
    this.searchInput = page.getByPlaceholder('Search logs...');
    this.projectFilter = page.getByRole('combobox').first();
    this.levelFilter = page.getByRole('combobox').last();
    this.refreshButton = page.getByRole('button', { name: /refresh/i });
    this.deleteButton = page.getByRole('button', { name: /delete/i });
    this.selectAllCheckbox = page.locator('thead input[type="checkbox"]');
    this.previousButton = page.getByRole('button', { name: 'Previous' });
    this.nextButton = page.getByRole('button', { name: 'Next' });
    this.logTable = page.locator('table');
  }

  async goto() {
    await this.page.goto('/logs');
    await this.waitForLogsPage();
  }

  async waitForLogsPage() {
    await this.logsHeading.waitFor();
  }

  async searchLogs(query: string) {
    await this.searchInput.fill(query);
    // Wait for debounce/API call
    await this.page.waitForTimeout(500);
  }

  async filterByProject(projectName: string) {
    await this.projectFilter.click();
    await this.page.getByRole('option', { name: projectName }).click();
  }

  async filterByLevel(level: string) {
    await this.levelFilter.click();
    await this.page.getByRole('option', { name: level }).click();
  }

  async clearFilters() {
    // Clear project filter
    await this.projectFilter.click();
    await this.page.getByRole('option', { name: 'All Projects' }).click();

    // Clear level filter
    await this.levelFilter.click();
    await this.page.getByRole('option', { name: 'All Levels' }).click();

    // Clear search
    await this.searchInput.clear();
  }

  async getLogRows() {
    return await this.page.locator('tbody tr').all();
  }

  async getLogCount(): Promise<number> {
    const rows = await this.getLogRows();
    return rows.length;
  }

  async selectLog(index: number) {
    const rows = await this.getLogRows();
    const checkbox = rows[index].locator('input[type="checkbox"]');
    await checkbox.check();
  }

  async selectAllLogs() {
    await this.selectAllCheckbox.check();
  }

  async clickLogRow(index: number) {
    const rows = await this.getLogRows();
    await rows[index].click();
  }

  async waitForLogDialog() {
    await this.page.getByRole('heading', { name: /log details/i }).waitFor();
  }

  async closeLogDialog() {
    await this.page.keyboard.press('Escape');
  }

  async deleteLogs() {
    this.page.once('dialog', async dialog => {
      await dialog.accept();
    });
    await this.deleteButton.click();
  }

  async goToNextPage() {
    await this.nextButton.click();
  }

  async goToPreviousPage() {
    await this.previousButton.click();
  }

  async getPaginationText(): Promise<string> {
    return await this.page.locator('text=/Showing .* of .*/').textContent() || '';
  }
}
