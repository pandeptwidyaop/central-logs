/* eslint-disable react-hooks/rules-of-hooks */
import { test as base } from '@playwright/test';
import { LoginPage } from '../page-objects/login.page';
import { DashboardPage } from '../page-objects/dashboard.page';
import { ProjectsPage } from '../page-objects/projects.page';
import { LogsPage } from '../page-objects/logs.page';
import { NavigationPage } from '../page-objects/navigation.page';

type PageObjects = {
  loginPage: LoginPage;
  dashboardPage: DashboardPage;
  projectsPage: ProjectsPage;
  logsPage: LogsPage;
  navigation: NavigationPage;
};

/**
 * Extended test fixture with page objects
 * Usage: import { test, expect } from './fixtures/auth.fixture';
 */
export const test = base.extend<PageObjects>({
  loginPage: async ({ page }, use) => {
    await use(new LoginPage(page));
  },
  dashboardPage: async ({ page }, use) => {
    await use(new DashboardPage(page));
  },
  projectsPage: async ({ page }, use) => {
    await use(new ProjectsPage(page));
  },
  logsPage: async ({ page }, use) => {
    await use(new LogsPage(page));
  },
  navigation: async ({ page }, use) => {
    await use(new NavigationPage(page));
  },
});

export { expect } from '@playwright/test';
