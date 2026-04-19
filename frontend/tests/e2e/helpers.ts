import { expect, type Locator, type Page } from "@playwright/test";

export const accounts = {
  admin: { email: "admin@northstar.id", password: "ChangeMe123!" },
  subdepartmentManager: { email: "salma.nuraini@codeid.co.id", password: "Manager123!" },
  teamManager: { email: "fajar.maulana@codeid.co.id", password: "Manager123!" },
  hrManager: { email: "dimas.kusuma@codeid.co.id", password: "Manager123!" },
  employee: { email: "nabila.putri@codeid.co.id", password: "Employee123!" },
  employeeFinalized: { email: "keisha.anindya@codeid.co.id", password: "Employee123!" },
  employeeApprovalFlow: { email: "yoga.prasetyo@codeid.co.id", password: "Employee123!" },
} as const;

const runDayOffset = Math.floor(Date.now() / 1000) % 17;

export async function login(page: Page, email: string, password: string) {
  await page.goto("/login", { waitUntil: "domcontentloaded" });
  await page.getByRole("textbox", { name: "Email" }).fill(email);
  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: /continue/i }).click();
  await page.waitForLoadState("networkidle");
}

export async function logoutFromProfile(page: Page, profilePath: string) {
  await page.goto(profilePath, { waitUntil: "networkidle" });
  const actionsCard = page.locator(".quick-actions");
  await actionsCard.getByRole("button", { name: /^logout$/i }).click();
  await expect(page).toHaveURL(/\/login$/);
}

export function uniqueEmail(prefix: string) {
  return `${prefix}.${Date.now()}@codeid.co.id`;
}

export function uniqueEmployeeCode(prefix: string) {
  return `${prefix}-${Date.now().toString().slice(-6)}`;
}

export function futureDate(offsetDays: number) {
  const value = new Date();
  value.setUTCDate(value.getUTCDate() + offsetDays + runDayOffset);
  return value.toISOString().slice(0, 10);
}

export function formatShortDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short", timeZone: "UTC" }).format(new Date(`${value}T00:00:00Z`));
}

export function formatFullDate(value: string) {
  return new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short", year: "numeric", timeZone: "UTC" }).format(new Date(`${value}T00:00:00Z`));
}

export function leaveRangeLabel(start: string, end: string) {
  if (start === end) return formatFullDate(start);
  return `${formatFullDate(start)} - ${formatFullDate(end)}`;
}

export async function openCrudCreate(page: Page, label: string) {
  await page.getByRole("button", { name: label }).click();
  await expect(page.getByRole("dialog")).toBeVisible();
}

export async function saveCrudDialog(page: Page, buttonName = /create record|save changes/i) {
  const dialog = page.getByRole("dialog");
  await dialog.getByRole("button", { name: buttonName }).click();
  await expect(dialog).toBeHidden();
}

export function fieldControl(container: Locator, label: string, selector = "input, select, textarea") {
  return container.locator(".field-block").filter({ hasText: label }).locator(selector).first();
}

export async function rowByText(tableShell: Locator, text: string) {
  return tableShell.locator(".table-row").filter({ hasText: text }).first();
}

export async function waitForRow(tableShell: Locator, text: string) {
  const row = await rowByText(tableShell, text);
  await expect(row).toBeVisible();
  return row;
}
