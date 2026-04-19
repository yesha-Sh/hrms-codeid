import { expect, test } from "@playwright/test";
import { accounts, fieldControl, futureDate, formatShortDate, leaveRangeLabel, login, openCrudCreate, saveCrudDialog, waitForRow } from "./helpers";

test.describe("Employee HRMS flows", () => {
  test("employee can manage personal attendance and pending leave", async ({ page }) => {
    await login(page, accounts.employee.email, accounts.employee.password);

    await page.goto("/employee/attendance", { waitUntil: "networkidle" });
    await openCrudCreate(page, "Add attendance");
    let dialog = page.getByRole("dialog");
    const attendanceDate = futureDate(72);
    await fieldControl(dialog, "Attendance date", "input").fill(attendanceDate);
    await fieldControl(dialog, "Check in", "input").fill("08:12");
    await fieldControl(dialog, "Check out", "input").fill("17:06");
    await fieldControl(dialog, "Status", "select").selectOption("on time");
    await fieldControl(dialog, "Notes", "textarea").fill("Playwright employee attendance");
    await saveCrudDialog(page);

    const attendanceRow = await waitForRow(page.locator(".table-shell").first(), formatShortDate(attendanceDate));
    await expect(attendanceRow).toContainText("On Time");
    await attendanceRow.getByRole("button", { name: /delete nabila putri/i }).click();
    await page.getByRole("button", { name: /delete record/i }).click();
    await expect(page.locator(".table-row").filter({ hasText: formatShortDate(attendanceDate) })).toHaveCount(0);

    await page.goto("/employee/leave", { waitUntil: "networkidle" });
    await openCrudCreate(page, "Request leave");
    dialog = page.getByRole("dialog");
    const leaveStart = futureDate(78);
    const leaveEnd = futureDate(79);
    await fieldControl(dialog, "Leave type", "select").selectOption({ label: "Annual Leave" });
    await fieldControl(dialog, "Start date", "input").fill(leaveStart);
    await fieldControl(dialog, "End date", "input").fill(leaveEnd);
    await fieldControl(dialog, "Status", "select").selectOption("pending");
    await fieldControl(dialog, "Reason", "textarea").fill("Playwright employee leave");
    await saveCrudDialog(page);

    const leaveRow = await waitForRow(page.locator(".table-shell").first(), leaveRangeLabel(leaveStart, leaveEnd));
    await leaveRow.getByRole("button", { name: /edit nabila putri/i }).click();
    dialog = page.getByRole("dialog");
    await fieldControl(dialog, "Status", "select").selectOption("cancelled");
    await saveCrudDialog(page, /save changes/i);
    await expect(leaveRow).toContainText("Cancelled");
  });

  test("employee finalized leave stays locked and profile reflects database selections", async ({ page }) => {
    await login(page, accounts.employeeFinalized.email, accounts.employeeFinalized.password);

    await page.goto("/employee/leave", { waitUntil: "networkidle" });
    const finalizedRow = page.locator(".table-row").filter({ hasText: "Rejected" }).first();
    await expect(finalizedRow).toBeVisible();
    await expect(finalizedRow.getByRole("button", { name: /edit /i })).toHaveCount(0);

    await page.goto("/employee/profile", { waitUntil: "networkidle" });
    await expect(page.getByRole("button", { name: /^logout$/i }).last()).toBeVisible();
    await page.getByRole("button", { name: /edit profile/i }).click();
    const dialog = page.getByRole("dialog");
    const departmentSelect = fieldControl(dialog, "Department", "select");
    const jobSelect = fieldControl(dialog, "Job title", "select");
    await expect(departmentSelect).toBeDisabled();
    await expect(jobSelect).toBeDisabled();
    await expect(departmentSelect).not.toHaveValue("");
    await expect(jobSelect).not.toHaveValue("");
  });
});
