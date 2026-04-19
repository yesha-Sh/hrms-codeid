import { expect, test } from "@playwright/test";
import { accounts, fieldControl, formatShortDate, futureDate, leaveRangeLabel, login, openCrudCreate, saveCrudDialog, waitForRow } from "./helpers";

test.describe("Manager HRMS flows", () => {
  test("team manager can manage team membership, self attendance, and self leave", async ({ page }) => {
    await login(page, accounts.teamManager.email, accounts.teamManager.password);

    await page.goto("/manager/team", { waitUntil: "networkidle" });
    await page.getByRole("button", { name: "Client Solutions Squad" }).click();

    const tables = page.locator(".table-shell");
    const membersTable = tables.nth(0);
    const availableTable = tables.nth(1);
    const availableRow = availableTable.locator(".table-row").first();
    const employeeName = await availableRow.locator(".person-cell__copy strong").innerText();
    await availableRow.getByRole("button", { name: /add to team/i }).click();

    await waitForRow(membersTable, employeeName);

    await page.goto("/manager/attendance", { waitUntil: "networkidle" });
    const attendanceDate = futureDate(41);
    await fieldControl(page.locator("form"), "Attendance date", "input").fill(attendanceDate);
    await fieldControl(page.locator("form"), "Check in", "input").fill("08:03");
    await fieldControl(page.locator("form"), "Check out", "input").fill("17:11");
    await fieldControl(page.locator("form"), "Notes", "textarea").fill("Playwright manager self attendance");
    await page.getByRole("button", { name: /add my attendance/i }).click();
    await expect(page.locator(".table-row").filter({ hasText: formatShortDate(attendanceDate) })).toBeVisible();

    await page.goto("/manager/leave", { waitUntil: "networkidle" });
    await openCrudCreate(page, "Request leave");
    const leaveStart = futureDate(52);
    const leaveEnd = futureDate(53);
    const dialog = page.getByRole("dialog");
    await fieldControl(dialog, "Leave type", "select").selectOption({ label: "Annual Leave" });
    await fieldControl(dialog, "Start date", "input").fill(leaveStart);
    await fieldControl(dialog, "End date", "input").fill(leaveEnd);
    await fieldControl(dialog, "Status", "select").selectOption("pending");
    await fieldControl(dialog, "Reason", "textarea").fill("Playwright manager self leave");
    await saveCrudDialog(page);
    await expect(page.locator(".table-row").filter({ hasText: leaveRangeLabel(leaveStart, leaveEnd) })).toBeVisible();
  });

  test("subdepartment manager can approve team leave through the UI", async ({ page }) => {
    const leaveStart = futureDate(64);
    const leaveEnd = futureDate(65);

    await login(page, accounts.employeeApprovalFlow.email, accounts.employeeApprovalFlow.password);
    await page.goto("/employee/leave", { waitUntil: "networkidle" });
    await openCrudCreate(page, "Request leave");
    let dialog = page.getByRole("dialog");
    await fieldControl(dialog, "Leave type", "select").selectOption({ label: "Annual Leave" });
    await fieldControl(dialog, "Start date", "input").fill(leaveStart);
    await fieldControl(dialog, "End date", "input").fill(leaveEnd);
    await fieldControl(dialog, "Status", "select").selectOption("pending");
    await fieldControl(dialog, "Reason", "textarea").fill("Playwright approval flow");
    await saveCrudDialog(page);
    await page.goto("/employee/profile", { waitUntil: "networkidle" });
    await page.locator(".quick-actions").getByRole("button", { name: /^logout$/i }).click();
    await expect(page).toHaveURL(/\/login$/);

    await login(page, accounts.hrManager.email, accounts.hrManager.password);
    await page.goto("/manager/approvals", { waitUntil: "networkidle" });
    const approvalRow = await waitForRow(page.locator(".table-shell").first(), leaveRangeLabel(leaveStart, leaveEnd));
    await approvalRow.getByRole("button", { name: /^approve$/i }).click();
    await expect(approvalRow).toContainText("Approved");

    await page.goto("/manager/profile", { waitUntil: "networkidle" });
    await page.locator(".quick-actions").getByRole("button", { name: /^logout$/i }).click();
    await expect(page).toHaveURL(/\/login$/);

    await login(page, accounts.employeeApprovalFlow.email, accounts.employeeApprovalFlow.password);
    await page.goto("/employee/leave", { waitUntil: "networkidle" });
    const finalizedRow = await waitForRow(page.locator(".table-shell").first(), leaveRangeLabel(leaveStart, leaveEnd));
    await expect(finalizedRow).toContainText("Approved");
    await expect(finalizedRow.getByRole("button", { name: /edit /i })).toHaveCount(0);
  });
});
