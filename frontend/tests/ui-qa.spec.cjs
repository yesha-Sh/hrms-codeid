const { test, devices } = require('@playwright/test');
const path = require('path');

test.setTimeout(120000);
test.use({ browserName: 'chromium' });

async function login(page, email, password) {
  await page.goto('http://localhost:5173/', { waitUntil: 'domcontentloaded' });
  await page.locator('input[type="email"]').fill(email);
  await page.locator('input[type="password"]').fill(password);
  await page.getByRole('button', { name: /continue/i }).click();
  await page.waitForLoadState('networkidle');
}

test('capture admin employees desktop', async ({ browser }) => {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1100 } });
  const page = await context.newPage();
  await login(page, 'admin@northstar.id', 'ChangeMe123!');
  await page.goto('http://localhost:5173/admin/employees', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'admin-employees-desktop.png'), fullPage: true });
});

test('capture manager pages desktop', async ({ browser }) => {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1100 } });
  const page = await context.newPage();
  await login(page, 'aria.pratama@codeid.co.id', 'Manager123!');
  await page.goto('http://localhost:5173/manager/team', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'manager-team-desktop.png'), fullPage: true });
  await page.goto('http://localhost:5173/manager/attendance', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'manager-attendance-desktop.png'), fullPage: true });
});

test('capture employee pages desktop', async ({ browser }) => {
  const context = await browser.newContext({ viewport: { width: 1440, height: 1100 } });
  const page = await context.newPage();
  await login(page, 'nabila.putri@codeid.co.id', 'Employee123!');
  await page.goto('http://localhost:5173/employee/leave', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'employee-leave-desktop.png'), fullPage: true });
  await page.goto('http://localhost:5173/employee/profile', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'employee-profile-desktop.png'), fullPage: true });
});

test('capture admin employees mobile', async ({ browser }) => {
  const context = await browser.newContext({ ...devices['iPhone 13'] });
  const page = await context.newPage();
  await login(page, 'admin@northstar.id', 'ChangeMe123!');
  await page.goto('http://localhost:5173/admin/employees', { waitUntil: 'networkidle' });
  await page.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'admin-employees-mobile.png'), fullPage: true });
});
