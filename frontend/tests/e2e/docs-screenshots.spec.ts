import path from "node:path";
import { expect, test } from "@playwright/test";
import { accounts, login, waitForRow } from "./helpers";

const outputDir = path.resolve(process.cwd(), "..", "docs", "screenshots");

test.describe.configure({ mode: "serial" });
test.skip(process.env.DOC_SCREENSHOTS !== "1", "Set DOC_SCREENSHOTS=1 to generate documentation screenshots.");

test("generate documentation screenshots", async ({ page }) => {
  await login(page, accounts.admin.email, accounts.admin.password);

  await page.goto("/admin/employees", { waitUntil: "networkidle" });
  await expect(page.locator(".table-shell").first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "admin-employees-page.png"), fullPage: true });

  await page.getByPlaceholder("Search employee, email, or employee code").fill("CID-006");
  const nabilaRow = await waitForRow(page.locator(".table-shell").first(), "Nabila Putri");
  await nabilaRow.getByRole("link", { name: /^open$/i }).click();
  await page.waitForURL(/\/admin\/employees\/.+$/);
  await expect(page.getByText("Secondary assignments", { exact: false }).first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "admin-employee-detail-assignments.png"), fullPage: true });

  await page.goto("/admin/departments", { waitUntil: "networkidle" });
  await expect(page.locator(".table-shell").first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "admin-departments-page.png"), fullPage: true });

  await page.goto("/admin/holidays", { waitUntil: "networkidle" });
  await expect(page.locator(".table-shell").first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "admin-holidays-page.png"), fullPage: true });

  await page.locator("aside").getByRole("button", { name: /^logout$/i }).click();
  await expect(page).toHaveURL(/\/login$/);

  await login(page, accounts.teamManager.email, accounts.teamManager.password);
  await page.goto("/manager/team", { waitUntil: "networkidle" });
  await expect(page.locator(".table-shell").first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "manager-team-page.png"), fullPage: true });

  await page.goto("/manager/approvals", { waitUntil: "networkidle" });
  await expect(page.locator(".table-shell").first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "manager-approvals-page.png"), fullPage: true });

  await page.locator("aside").getByRole("button", { name: /^logout$/i }).click();
  await expect(page).toHaveURL(/\/login$/);

  await login(page, accounts.employee.email, accounts.employee.password);
  await page.goto("/employee/leave", { waitUntil: "networkidle" });
  await expect(page.locator(".table-shell").first()).toBeVisible();
  await page.screenshot({ path: path.join(outputDir, "employee-leave-page.png"), fullPage: true });
});
