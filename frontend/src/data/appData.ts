import {
  Activity,
  Building2,
  ClipboardList,
  LayoutDashboard,
  Settings,
  ShieldCheck,
  Users,
  WalletCards,
} from "lucide-react";
import type {
  ActionRow,
  ActivityItem,
  DistributionItem,
  NavItem,
  PageConfig,
  Role,
  StatCard,
  TableRow,
} from "../types";

export const roleLabels: Record<Role, string> = {
  admin: "Admin",
  manager: "Manager",
  employee: "Employee",
};

export const navByRole: Record<Role, NavItem[]> = {
  admin: [
    { label: "Dashboard", path: "/admin/dashboard", icon: LayoutDashboard },
    { label: "Employees", path: "/admin/employees", icon: Users },
    { label: "Departments", path: "/admin/departments", icon: Building2 },
    { label: "Attendance", path: "/admin/attendance", icon: Activity },
    { label: "Leave", path: "/admin/leave", icon: ClipboardList },
    { label: "Jobs", path: "/admin/jobs", icon: WalletCards },
    { label: "Holidays", path: "/admin/holidays", icon: Settings },
    { label: "Audit Logs", path: "/admin/audit-logs", icon: ShieldCheck },
  ],
  manager: [
    { label: "Dashboard", path: "/manager/dashboard", icon: LayoutDashboard },
    { label: "Team", path: "/manager/team", icon: Users },
    { label: "Attendance", path: "/manager/attendance", icon: Activity },
    { label: "Leave", path: "/manager/leave", icon: ClipboardList },
    { label: "Approvals", path: "/manager/approvals", icon: ClipboardList },
  ],
  employee: [
    { label: "Dashboard", path: "/employee/dashboard", icon: LayoutDashboard },
    { label: "Attendance", path: "/employee/attendance", icon: Activity },
    { label: "Leave", path: "/employee/leave", icon: ClipboardList },
  ],
};

export const bars = [
  { day: "Mon", outerHeight: 112, innerHeight: 64, tone: "teal" },
  { day: "Tue", outerHeight: 120, innerHeight: 72, tone: "teal" },
  { day: "Wed", outerHeight: 126, innerHeight: 84, tone: "deep" },
  { day: "Thu", outerHeight: 120, innerHeight: 78, tone: "soft" },
  { day: "Fri", outerHeight: 128, innerHeight: 88, tone: "gold" },
] as const;

export const activities: ActivityItem[] = [
  { title: "Attendance sync", text: "completed for Jakarta HQ and West Java offices.", meta: "09:10 · system log", tone: "teal" },
  { title: "Leave escalation", text: "triggered because 5 requests crossed manager SLA.", meta: "08:42 · workflow alert", tone: "gold" },
  { title: "Department update", text: "Engineering headcount increased by 3 after approvals.", meta: "08:15 · employee records", tone: "blue" },
];

export const adminConfigs: Record<string, PageConfig> = {
  dashboard: {
    sectionLabel: "Admin Dashboard",
    badge: "Live snapshot",
    rangeLabel: "Q2 2026 · All regions",
    searchLabel: "Search employees, teams, or logs",
    insight: {
      label: "Workforce health",
      value: "94.8%",
      detail: "PT. CODEID workforce coverage is stable across the seeded division structure and current attendance cycle.",
      statusLabel: "Stable",
      progressA: { label: "Leave approvals", value: "82%", width: "82%" },
      progressB: { label: "Audit coverage", value: "97%", width: "97%" },
    },
  },
  employees: {
    sectionLabel: "Admin Employees",
    badge: "Directory view",
    rangeLabel: "April 2026 · All offices",
    searchLabel: "Search name, ID, email, or department",
    insight: {
      label: "Record quality",
      value: "92.1%",
      detail: "Most PT. CODEID employee records are complete, with a smaller review queue for manager links and staffing data.",
      statusLabel: "Healthy",
      progressA: { label: "Verified documents", value: "84%", width: "84%" },
      progressB: { label: "Payroll mapped", value: "96%", width: "96%" },
    },
  },
  departments: {
    sectionLabel: "Admin Departments",
    badge: "Structure view",
    rangeLabel: "April 2026 · Org structure",
    searchLabel: "Search department, manager, or location",
    insight: {
      label: "Structure health",
      value: "96.4%",
      detail: "The PT. CODEID department tree is mapped, with only a few reporting and manager assignments still needing cleanup.",
      statusLabel: "Organized",
      progressA: { label: "Manager coverage", value: "83%", width: "83%" },
      progressB: { label: "Location mapped", value: "94%", width: "94%" },
    },
  },
  attendance: {
    sectionLabel: "Admin Attendance",
    badge: "Live monitor",
    rangeLabel: "Today · 7 Apr 2026",
    searchLabel: "Search employee, team, or exception",
    insight: {
      label: "Attendance health",
      value: "94.8%",
      detail: "Most attendance signals are within normal range, with a small cluster of late check-ins and unresolved mismatches to review.",
      statusLabel: "Stable",
      progressA: { label: "Check-out completion", value: "88%", width: "88%" },
      progressB: { label: "Exception resolved", value: "76%", width: "76%" },
    },
  },
  leave: {
    sectionLabel: "Admin Leave",
    badge: "Approval center",
    rangeLabel: "April 2026 · All requests",
    searchLabel: "Search employee, leave type, or status",
    insight: {
      label: "Leave health",
      value: "81.6%",
      detail: "The leave queue is stable, though several overdue approvals need manager follow-up today.",
      statusLabel: "Reviewing",
      progressA: { label: "Approved within SLA", value: "78%", width: "78%" },
      progressB: { label: "Balance synced", value: "93%", width: "93%" },
    },
  },
  jobs: {
    sectionLabel: "Admin Jobs",
    badge: "Role catalog",
    rangeLabel: "April 2026 · Active roles",
    searchLabel: "Search title, grade, or salary band",
    insight: {
      label: "Catalog health",
      value: "89.4%",
      detail: "Most PT. CODEID job records are aligned to departments and levels, with a small set still under review.",
      statusLabel: "Aligned",
      progressA: { label: "Mapped to departments", value: "91%", width: "91%" },
      progressB: { label: "Salary bands reviewed", value: "74%", width: "74%" },
    },
  },
  audit: {
    sectionLabel: "Admin Audit Logs",
    badge: "Compliance stream",
    rangeLabel: "Today · All entities",
    searchLabel: "Search actor, entity, or action",
    insight: {
      label: "Audit integrity",
      value: "97.2%",
      detail: "System logging is healthy, with only a very small set of retries still pending after write bursts.",
      statusLabel: "Protected",
      progressA: { label: "Write success", value: "99%", width: "99%" },
      progressB: { label: "Entity coverage", value: "95%", width: "95%" },
    },
  },
  holidays: {
    sectionLabel: "Admin Holidays",
    badge: "Calendar management",
    rangeLabel: "2026 · PT. CODEID",
    searchLabel: "Search holiday, year, or location",
    insight: {
      label: "Calendar readiness",
      value: "95.2%",
      detail: "Indonesia holiday coverage is seeded for the current year, with room to add future-year company and location-specific dates.",
      statusLabel: "Current",
      progressA: { label: "2026 coverage", value: "100%", width: "100%" },
      progressB: { label: "Branch-specific entries", value: "74%", width: "74%" },
    },
  },
};

export const managerConfig: PageConfig = {
  sectionLabel: "Manager Dashboard",
  badge: "Team snapshot",
  rangeLabel: "This week · Team view",
  searchLabel: "Search team member or request",
  insight: {
    label: "Team attendance",
    value: "92.6%",
    detail: "Your team is tracking close to target, with a few leave approvals and attendance exceptions still pending.",
    statusLabel: "Steady",
    progressA: { label: "On-time arrivals", value: "89%", width: "89%" },
    progressB: { label: "Approvals closed", value: "81%", width: "81%" },
  },
};

export const employeeConfig: PageConfig = {
  sectionLabel: "Employee Dashboard",
  badge: "Personal overview",
  rangeLabel: "This month · My data",
  searchLabel: "Search leave, policy, or attendance",
  insight: {
    label: "Attendance record",
    value: "96.1%",
    detail: "Your attendance record is strong this month, with leave balance and profile data fully synced.",
    statusLabel: "On track",
    progressA: { label: "On-time check-ins", value: "94%", width: "94%" },
    progressB: { label: "Leave balance synced", value: "100%", width: "100%" },
  },
};

export const dashboardStats: StatCard[] = [
  { label: "Total employees", value: "248", detail: "+12 new hires this quarter" },
  { label: "Today attendance", value: "231", detail: "17 late arrivals flagged automatically", tone: "accent" },
  { label: "Pending leave requests", value: "14", detail: "5 require manager action today", tone: "warning" },
  { label: "Total departments", value: "18", detail: "6 active regions with assigned teams", tone: "cool" },
];

export const employeeStats: StatCard[] = [
  { label: "Active employees", value: "241", detail: "7 employees currently inactive or on long leave" },
  { label: "New hires this month", value: "12", detail: "Mostly Engineering and Operations.", tone: "accent" },
  { label: "Incomplete profiles", value: "19", detail: "Documents or payroll details still missing.", tone: "warning" },
  { label: "Pending changes", value: "8", detail: "Role, salary, and department changes awaiting final review", tone: "cool" },
];

export const departmentStats: StatCard[] = [
  { label: "Total departments", value: "18", detail: "Including business units, regional teams, and support functions" },
  { label: "Assigned managers", value: "15", detail: "3 departments still require final manager assignment", tone: "accent" },
  { label: "Sub-departments", value: "7", detail: "Grouped under Operations and Sales.", tone: "warning" },
  { label: "Location coverage", value: "6", detail: "Departments currently distributed across 6 office or regional locations", tone: "cool" },
];

export const attendanceStats: StatCard[] = [
  { label: "Present today", value: "231", detail: "Out of 248 employees tracked for the current shift" },
  { label: "Late arrivals", value: "17", detail: "Most were concentrated in Sales and West Java field teams", tone: "accent" },
  { label: "Remote check-ins", value: "9", detail: "Verified remote status across approved teams and locations", tone: "warning" },
  { label: "Exceptions open", value: "6", detail: "Missing check-outs, mismatches, or manual attendance corrections", tone: "cool" },
];

export const leaveStats: StatCard[] = [
  { label: "Open requests", value: "14", detail: "5 requests are beyond the manager SLA" },
  { label: "Approved today", value: "9", detail: "Mostly annual leave and sick leave", tone: "accent" },
  { label: "Rejected", value: "3", detail: "Conflicts with staffing coverage or balance", tone: "warning" },
  { label: "Balance sync", value: "98%", detail: "Most leave balances are current across employee records", tone: "cool" },
];

export const jobStats: StatCard[] = [
  { label: "Active jobs", value: "26", detail: "Across operations, engineering, HR, and field teams" },
  { label: "Open alignments", value: "7", detail: "Need updated band or department mapping", tone: "accent" },
  { label: "Salary bands", value: "11", detail: "Current range library used across active roles", tone: "warning" },
  { label: "Recently updated", value: "4", detail: "Titles revised this week with hiring scope changes", tone: "cool" },
];

export const auditStats: StatCard[] = [
  { label: "Events today", value: "1,284", detail: "Across HR records, approvals, and attendance writes" },
  { label: "High priority", value: "6", detail: "Sensitive changes affecting compensation or manager scope", tone: "accent" },
  { label: "Retrying logs", value: "3", detail: "Minor transient failures currently being replayed", tone: "warning" },
  { label: "Coverage", value: "95%", detail: "Core entities are fully tracked with entity and actor metadata", tone: "cool" },
];

export const holidayStats: StatCard[] = [
  { label: "2026 holidays", value: "9", detail: "Indonesia and PT. CODEID calendar dates currently seeded" },
  { label: "Company-wide", value: "7", detail: "Shared across Jakarta, Bandung, and remote teams", tone: "accent" },
  { label: "Location scoped", value: "2", detail: "Branch-specific or operational calendar entries", tone: "warning" },
  { label: "Future planning", value: "2027", detail: "Ready for next-year holiday additions and revisions", tone: "cool" },
];

export const managerStats: StatCard[] = [
  { label: "Team members", value: "18", detail: "Across Engineering Platform and Support" },
  { label: "Today present", value: "16", detail: "2 exceptions still need attention", tone: "accent" },
  { label: "Pending approvals", value: "5", detail: "Leave and schedule adjustments awaiting you", tone: "warning" },
  { label: "Performance notes", value: "7", detail: "Recent feedback entries captured this month", tone: "cool" },
];

export const selfStats: StatCard[] = [
  { label: "Attendance this month", value: "96%", detail: "On-time across 22 of 23 tracked days" },
  { label: "Leave balance", value: "8 days", detail: "Annual leave remaining this year", tone: "accent" },
  { label: "Remote days", value: "3", detail: "Approved remote work days this month", tone: "warning" },
  { label: "Profile status", value: "100%", detail: "Personal and payroll details are complete", tone: "cool" },
];

export const distribution: DistributionItem[] = [
  { label: "Operations", value: "64", width: "82%", tone: "teal" },
  { label: "Engineering", value: "52", width: "68%", tone: "blue" },
  { label: "People & Culture", value: "37", width: "46%", tone: "gold" },
];

export const locations: DistributionItem[] = [
  { label: "Jakarta HQ", value: "8", width: "78%", tone: "teal" },
  { label: "West Java", value: "4", width: "44%", tone: "blue" },
  { label: "Surabaya", value: "3", width: "30%", tone: "gold" },
];

export const hotspots: DistributionItem[] = [
  { label: "Sales", value: "7", width: "70%", tone: "gold" },
  { label: "Field Ops", value: "5", width: "52%", tone: "teal" },
  { label: "Support", value: "3", width: "31%", tone: "blue" },
];

export const employeeRows: TableRow[] = [
  { initials: "ML", name: "Maya Lestari", meta: "EMP-1024 · maya@codeid.co.id", cols: ["Finance", "Active", "96%", "2h ago"], pillIndex: 1, pillTone: "teal", tone: "default" },
  { initials: "RK", name: "Rafi Kurniawan", meta: "EMP-1091 · rafi@codeid.co.id", cols: ["Engineering", "Active", "88%", "4h ago"], pillIndex: 1, pillTone: "teal", tone: "cool" },
  { initials: "NH", name: "Nadia Hartono", meta: "EMP-1107 · nadia@codeid.co.id", cols: ["Operations", "Review", "73%", "6h ago"], pillIndex: 1, pillTone: "warning", tone: "warning" },
];

export const departmentRows: TableRow[] = [
  { initials: "OP", name: "Operations", meta: "Parent dept · 2 child teams", cols: ["Nadia Hartono", "Jakarta HQ", "64", "Healthy", "2h ago"], pillIndex: 3, pillTone: "teal", tone: "default" },
  { initials: "FS", name: "Field Support", meta: "Child of Operations", cols: ["Rafi Kurniawan", "West Java", "21", "Healthy", "5h ago"], pillIndex: 3, pillTone: "teal", tone: "soft" },
  { initials: "PC", name: "People & Culture", meta: "Parent dept · manager pending", cols: ["Unassigned", "Jakarta HQ", "37", "Review", "1d ago"], pillIndex: 3, pillTone: "warning", tone: "warning" },
];

export const attendanceRows: TableRow[] = [
  { initials: "ML", name: "Maya Lestari", meta: "Finance · EMP-1024", cols: ["08:01", "17:06", "On time", "Jakarta HQ", "Synced"], pillIndex: 2, pillTone: "teal", tone: "default" },
  { initials: "AS", name: "Arif Saputra", meta: "Sales · EMP-0958", cols: ["08:17", "--", "Late", "West Java", "Needs review"], pillIndex: 2, pillTone: "warning", tone: "warning" },
  { initials: "RK", name: "Rafi Kurniawan", meta: "Engineering · EMP-1091", cols: ["08:04", "--", "Remote", "Remote", "Approved"], pillIndex: 2, pillTone: "cool", tone: "cool" },
];

export const leaveRows: TableRow[] = [
  { initials: "NH", name: "Nadia Hartono", meta: "Operations · Annual leave", cols: ["7 Apr - 10 Apr", "Pending", "Rafi Kurniawan", "SLA risk"], pillIndex: 1, pillTone: "warning", tone: "warning" },
  { initials: "ML", name: "Maya Lestari", meta: "Finance · Sick leave", cols: ["6 Apr", "Approved", "Ariana Reyes", "Closed"], pillIndex: 1, pillTone: "teal", tone: "default" },
  { initials: "AS", name: "Arif Saputra", meta: "Sales · Emergency leave", cols: ["5 Apr", "Review", "Ariana Reyes", "Need note"], pillIndex: 1, pillTone: "cool", tone: "cool" },
];

export const jobRows: TableRow[] = [
  { initials: "SB", name: "Senior Backend Engineer", meta: "Engineering Platform", cols: ["Band 5", "IDR 24M - 31M", "Engineering", "Updated today"], tone: "cool" },
  { initials: "OA", name: "Operations Analyst", meta: "Field Operations", cols: ["Band 3", "IDR 12M - 16M", "Operations", "Needs review"], tone: "default" },
  { initials: "HR", name: "HR Business Partner", meta: "People & Culture", cols: ["Band 4", "IDR 17M - 22M", "People", "Aligned"], tone: "warning" },
];

export const auditRows: TableRow[] = [
  { initials: "AR", name: "Ariana Reyes", meta: "Compensation update", cols: ["employees", "update", "EMP-1024", "09:11", "High"], pillIndex: 4, pillTone: "warning", tone: "warning" },
  { initials: "SY", name: "System", meta: "Attendance sync completed", cols: ["attendances", "sync", "248 records", "09:10", "Normal"], pillIndex: 4, pillTone: "teal", tone: "default" },
  { initials: "RK", name: "Rafi Kurniawan", meta: "Leave approval action", cols: ["leaves", "approve", "REQ-1102", "08:42", "Normal"], pillIndex: 4, pillTone: "cool", tone: "cool" },
];


export const approvals: ActionRow[] = [
  { initials: "NH", name: "Nadia Hartono", detail: "Annual leave · Operations", status: "SLA risk", tone: "warning" },
  { initials: "RK", name: "Rafi Kurniawan", detail: "Job change · Engineering", status: "Review", tone: "cool" },
];

export const managerApprovals: ActionRow[] = [
  { initials: "AS", name: "Arif Saputra", detail: "Leave request · Sales", status: "Pending", tone: "warning" },
  { initials: "DN", name: "Dina Novita", detail: "Schedule swap · Support", status: "Review", tone: "cool" },
];
