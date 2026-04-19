import { expect, test } from "@playwright/test";
import { accounts, login, logoutFromProfile } from "./helpers";

test.describe("Auth flows", () => {
  test("admin can sign in and sign out", async ({ page }) => {
    await login(page, accounts.admin.email, accounts.admin.password);
    await expect(page).toHaveURL(/\/admin\/dashboard$/);
    await expect(page.getByText("PT. CODEID", { exact: true }).first()).toBeVisible();
    await logoutFromProfile(page, "/admin/profile");
  });

  test("manager can sign in and sign out", async ({ page }) => {
    await login(page, accounts.teamManager.email, accounts.teamManager.password);
    await expect(page).toHaveURL(/\/manager\/dashboard$/);
    await logoutFromProfile(page, "/manager/profile");
  });

  test("employee can sign in and sign out", async ({ page }) => {
    await login(page, accounts.employee.email, accounts.employee.password);
    await expect(page).toHaveURL(/\/employee\/dashboard$/);
    await logoutFromProfile(page, "/employee/profile");
  });
});
