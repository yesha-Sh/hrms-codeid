const { chromium, devices } = require('@playwright/test');
const path = require('path');
(async () => {
  const browser = await chromium.launch({ headless: true });
  async function login(page, email, password) {
    await page.goto('http://localhost:5173/', { waitUntil: 'domcontentloaded' });
    await page.locator('input[type="email"]').fill(email);
    await page.locator('input[type="password"]').fill(password);
    await page.getByRole('button', { name: /continue/i }).click();
    await page.waitForLoadState('networkidle');
  }
  const desktop = await browser.newContext({ viewport: { width: 1440, height: 1100 } });
  const adminPage = await desktop.newPage();
  await login(adminPage, 'admin@northstar.id', 'ChangeMe123!');
  await adminPage.goto('http://localhost:5173/admin/employees', { waitUntil: 'networkidle' });
  await adminPage.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'admin-employees-desktop.png'), fullPage: true });
  const employeePage = await desktop.newPage();
  await login(employeePage, 'nabila.putri@codeid.co.id', 'Employee123!');
  await employeePage.goto('http://localhost:5173/employee/leave', { waitUntil: 'networkidle' });
  await employeePage.screenshot({ path: path.join('C:/laragon/www/HRMS/tmp-qa', 'employee-leave-desktop.png'), fullPage: true });
  await browser.close();
})();
