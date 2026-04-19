import { expect, test } from "@playwright/test";
import { accounts, fieldControl, login, openCrudCreate, saveCrudDialog, uniqueEmail, uniqueEmployeeCode, waitForRow } from "./helpers";

test.describe("Admin HRMS flows", () => {
  test("admin can filter by management scope and manage employees", async ({ page }) => {
    await login(page, accounts.admin.email, accounts.admin.password);
    await page.goto("/admin/employees", { waitUntil: "networkidle" });

    const scopeSelect = page.locator(".toolbar-select select");
    await scopeSelect.selectOption("team_manager");
    await expect(page.getByText("Showing 2 employee records")).toBeVisible();

    await scopeSelect.selectOption("all");
    await openCrudCreate(page, "Add employee");

    const email = uniqueEmail("automation.admin");
    const code = uniqueEmployeeCode("CID-AUTO");
    const dialog = page.getByRole("dialog");
    await fieldControl(dialog, "First name", "input").fill("Automation");
    await fieldControl(dialog, "Last name", "input").fill("Candidate");
    await fieldControl(dialog, "Employee code", "input").fill(code);
    await fieldControl(dialog, "Email", "input").fill(email);
    await fieldControl(dialog, "Phone number", "input").fill("+62-811-7000");
    await fieldControl(dialog, "Hire date", "input").fill("2026-04-09");
    await fieldControl(dialog, "Salary", "input").fill("15000000");
    await fieldControl(dialog, "Employee type", "select").selectOption({ label: "Full-time" });
    await fieldControl(dialog, "Employment status", "select").selectOption({ label: "Active" });
    await fieldControl(dialog, "Department", "select").selectOption({ label: "System Development" });
    await fieldControl(dialog, "Primary job", "select").selectOption({ label: "System Developer" });
    await fieldControl(dialog, "Location", "select").selectOption({ label: "Jakarta Headquarters" });
    await fieldControl(dialog, "Work mode", "select").selectOption("hybrid");
    await fieldControl(dialog, "Management scope", "select").selectOption("individual_contributor");
    await saveCrudDialog(page);

    await page.getByPlaceholder("Search employee, email, or employee code").fill(code);
    const tableShell = page.locator(".table-shell").first();
    const row = await waitForRow(tableShell, code);
    await expect(row).toContainText("Automation Candidate");

    await row.getByRole("button", { name: /delete automation candidate/i }).click();
    await page.getByRole("button", { name: /delete record/i }).click();
    await expect(tableShell.getByText(code)).toHaveCount(0);
  });

  test("admin employee table export works", async ({ page }) => {
    await login(page, accounts.admin.email, accounts.admin.password);
    await page.goto("/admin/employees", { waitUntil: "networkidle" });

    const downloadPromise = page.waitForEvent("download");
    await page.getByRole("button", { name: /export csv/i }).click();
    const download = await downloadPromise;
    await expect(download.suggestedFilename()).toContain("employees");
  });
});
