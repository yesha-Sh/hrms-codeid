import { useEffect, useMemo, useRef, useState } from "react";
import { getDashboard, getProfile, type DashboardResponse, type ProfileResponse } from "../api/client";
import {
  listAssignments,
  listAttendances,
  listAuditLogs,
  listDepartments,
  listEmployees,
  listHolidays,
  listJobs,
  listLeaveRequests,
  type AssignmentResource,
  type AttendanceResource,
  type AuditLogResource,
  type DepartmentResource,
  type EmployeeResource,
  type HolidayResource,
  type JobResource,
  type LeaveResource,
  type PaginatedResponse,
  type ResourceListQuery,
} from "../api/resources";
import { useAuth } from "../auth/AuthContext";
import type { ActionRow, ActivityItem, DistributionItem, MiniStatItem, StatCard } from "../types";

type BarItem = {
  day: string;
  outerHeight: number;
  innerHeight: number;
  tone: "teal" | "deep" | "soft" | "gold";
};

type AdminDashboardInsights = {
  stats: StatCard[];
  attendanceBars: BarItem[];
  attendanceMetrics: MiniStatItem[];
  recentActivity: ActivityItem[];
  priorityApprovals: ActionRow[];
  loading: boolean;
  error: string | null;
};

type ManagerDashboardInsights = {
  stats: StatCard[];
  overview: MiniStatItem[];
  approvals: MiniStatItem[];
  recentActivity: ActivityItem[];
  loading: boolean;
  error: string | null;
};

type EmployeeDashboardInsights = {
  stats: StatCard[];
  overview: MiniStatItem[];
  balance: MiniStatItem[];
  loading: boolean;
  error: string | null;
};

type PageSummaryInsights = {
  stats: StatCard[];
  miniStats: MiniStatItem[];
  distribution?: DistributionItem[];
  activities?: ActivityItem[];
  actions?: ActionRow[];
  notice?: { text: string; action: string };
  loading: boolean;
  error: string | null;
};

const emptyStats: StatCard[] = [
  { label: "Loading", value: "--", detail: "Pulling live data." },
  { label: "Loading", value: "--", detail: "Pulling live data.", tone: "accent" },
  { label: "Loading", value: "--", detail: "Pulling live data.", tone: "warning" },
  { label: "Loading", value: "--", detail: "Pulling live data.", tone: "cool" },
];

async function collectAllPages<TResource>(
  loader: (accessToken: string, query?: ResourceListQuery) => Promise<PaginatedResponse<TResource>>,
  accessToken: string,
  query?: ResourceListQuery,
) {
  const firstPage = await loader(accessToken, { ...query, page: 1, limit: 100 });
  const items = [...firstPage.items];
  const pageCount = Math.max(1, Math.ceil(firstPage.meta.total / firstPage.meta.limit));

  for (let currentPage = 2; currentPage <= pageCount; currentPage += 1) {
    const response = await loader(accessToken, { ...query, page: currentPage, limit: 100 });
    items.push(...response.items);
  }

  return items;
}

function useAuthorizedLoader<T>(loader: (accessToken: string) => Promise<T>, initialValue: T) {
  const { accessToken, loading: authLoading } = useAuth();
  const [data, setData] = useState<T>(initialValue);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const loaderRef = useRef(loader);

  loaderRef.current = loader;

  useEffect(() => {
    if (authLoading) return;
    if (!accessToken) {
      setData(initialValue);
      setLoading(false);
      return;
    }

    let active = true;
    setLoading(true);
    setError(null);

    void loaderRef.current(accessToken)
      .then((nextData) => {
        if (active) setData(nextData);
      })
      .catch((nextError) => {
        if (!active) return;
        setError(nextError instanceof Error ? nextError.message : "Could not load dashboard data");
        setData(initialValue);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [accessToken, authLoading, initialValue]);

  return { data, loading, error };
}

function initials(value: string) {
  return value
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("");
}

function titleCase(value?: string | null) {
  return (value ?? "")
    .split(/[\s_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatShortDate(value?: string | null) {
  if (!value) return "--";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short" }).format(parsed);
}

function formatDateTime(value?: string | null) {
  if (!value) return "--";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("en-GB", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(parsed);
}

function formatRange(start?: string | null, end?: string | null) {
  if (!start && !end) return "--";
  if (!end || start === end) return formatShortDate(start);
  return `${formatShortDate(start)} - ${formatShortDate(end)}`;
}

function countBy<T>(items: T[], keyGetter: (item: T) => string | undefined | null) {
  const counts = new Map<string, number>();
  items.forEach((item) => {
    const key = keyGetter(item)?.trim();
    if (!key) return;
    counts.set(key, (counts.get(key) ?? 0) + 1);
  });
  return counts;
}

function topDistributions(counts: Map<string, number>, maxItems = 4): DistributionItem[] {
  const sorted = [...counts.entries()].sort((left, right) => right[1] - left[1]).slice(0, maxItems);
  const max = Math.max(1, ...sorted.map((entry) => entry[1]));
  const tones: DistributionItem["tone"][] = ["teal", "blue", "gold", "soft"];

  return sorted.map(([label, value], index) => ({
    label,
    value: `${value}`,
    width: `${Math.max(18, Math.round((value / max) * 100))}%`,
    tone: tones[index % tones.length],
  }));
}

function leaveTone(status?: string): ActionRow["tone"] {
  const normalized = (status ?? "").toLowerCase();
  if (["pending", "review", "late"].includes(normalized)) return "warning";
  if (["approved", "remote"].includes(normalized)) return "cool";
  return "default";
}

function auditTone(action?: string): ActivityItem["tone"] {
  const normalized = (action ?? "").toLowerCase();
  if (normalized.includes("delete")) return "gold";
  if (normalized.includes("create")) return "teal";
  return "blue";
}

function attendanceTone(status?: string): ActivityItem["tone"] {
  const normalized = (status ?? "").toLowerCase();
  if (normalized === "late") return "gold";
  if (normalized === "remote") return "blue";
  return "teal";
}

function toAdminStats(data: DashboardResponse): StatCard[] {
  return [
    { label: "Total employees", value: `${data.total_employees ?? "--"}`, detail: "Live org count across PT. CODEID." },
    { label: "Today attendance", value: `${data.today_attendance ?? "--"}`, detail: "Attendance entries recorded for today.", tone: "accent" },
    { label: "Pending leave requests", value: `${data.pending_leave_requests ?? "--"}`, detail: "Requests waiting for review.", tone: "warning" },
    { label: "Total departments", value: `${data.total_departments ?? "--"}`, detail: "Current department structure count.", tone: "cool" },
  ];
}

function toManagerStats(data: DashboardResponse): StatCard[] {
  return [
    { label: "Team members", value: `${data.team_members ?? "--"}`, detail: "People currently in your scope." },
    { label: "Today present", value: `${data.today_present ?? "--"}`, detail: "Attendance entries recorded today.", tone: "accent" },
    { label: "Pending approvals", value: `${data.pending_approvals ?? "--"}`, detail: "Items needing your review.", tone: "warning" },
    { label: "Secondary assignments", value: `${data.active_secondary_assignments ?? "--"}`, detail: "Active flexible staffing roles.", tone: "cool" },
  ];
}

function toEmployeeStats(data: DashboardResponse): StatCard[] {
  return [
    { label: "Attendance entries", value: `${data.attendance_entries ?? "--"}`, detail: "Your recorded attendance history." },
    { label: "Pending leaves", value: `${data.pending_leaves ?? "--"}`, detail: "Leave requests still in progress.", tone: "accent" },
    { label: "Secondary assignments", value: `${data.active_secondary_assignments ?? "--"}`, detail: "Additional active role assignments.", tone: "warning" },
    { label: "Job level", value: `${data.job_level_name ?? "--"}`, detail: "Your current primary job level.", tone: "cool" },
  ];
}

function attendanceBarsFromRecords(items: AttendanceResource[]): BarItem[] {
  const grouped = new Map<string, AttendanceResource[]>();
  items.forEach((item) => {
    const key = item.attendance_date;
    const current = grouped.get(key) ?? [];
    current.push(item);
    grouped.set(key, current);
  });

  const dates = [...grouped.keys()].sort().slice(-5);
  const values = dates.map((date) => {
    const entries = grouped.get(date) ?? [];
    const total = entries.length;
    const healthy = entries.filter((entry) => ["on time", "present", "remote"].includes(entry.status.toLowerCase())).length;
    const late = entries.filter((entry) => entry.status.toLowerCase() === "late").length;
    return { date, total, healthy, late };
  });

  const maxTotal = Math.max(1, ...values.map((entry) => entry.total));
  return values.map((entry) => ({
    day: new Intl.DateTimeFormat("en-GB", { weekday: "short" }).format(new Date(entry.date)),
    outerHeight: Math.max(48, Math.round((entry.total / maxTotal) * 128)),
    innerHeight: Math.max(18, Math.round((entry.healthy / maxTotal) * 96)),
    tone: entry.late > 0 ? "gold" : entry.healthy === entry.total ? "teal" : "soft",
  }));
}

function auditActivities(items: AuditLogResource[]): ActivityItem[] {
  return items.slice(0, 4).map((item) => ({
    title: `${titleCase(item.entity)} ${titleCase(item.action)}`,
    text: item.actor_email ? `recorded by ${item.actor_email}.` : "recorded by the system.",
    meta: formatDateTime(item.created_at),
    tone: auditTone(item.action),
  }));
}

function attendanceActivities(items: AttendanceResource[]): ActivityItem[] {
  return items.slice(0, 4).map((item) => ({
    title: `${item.employee_name}`,
    text: `${titleCase(item.status)} attendance on ${formatShortDate(item.attendance_date)}.`,
    meta: `${item.department_name ?? "Department not set"} · ${item.location_name ?? "Location not set"}`,
    tone: attendanceTone(item.status),
  }));
}

function leaveActivities(items: LeaveResource[]): ActivityItem[] {
  return items.slice(0, 4).map((item) => ({
    title: item.employee_name,
    text: `${item.leave_type_name ?? "Leave"} is ${titleCase(item.status).toLowerCase()} for ${formatRange(item.start_date, item.end_date)}.`,
    meta: item.approver_name ? `Approver: ${item.approver_name}` : "Awaiting approver assignment",
    tone: auditTone(item.status),
  }));
}

function leaveActions(items: LeaveResource[]): ActionRow[] {
  return items.slice(0, 4).map((item) => ({
    initials: initials(item.employee_name),
    name: item.employee_name,
    detail: `${item.leave_type_name ?? "Leave"} · ${formatRange(item.start_date, item.end_date)}`,
    status: titleCase(item.status),
    tone: leaveTone(item.status),
  }));
}

export function useAdminDashboardInsights(): AdminDashboardInsights {
  const initialValue = useMemo(() => ({
    dashboard: null as DashboardResponse | null,
    attendances: [] as AttendanceResource[],
    leaves: [] as LeaveResource[],
    audits: [] as AuditLogResource[],
  }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => {
    const [dashboard, attendances, leaves, audits] = await Promise.all([
      getDashboard(accessToken, "admin"),
      collectAllPages(listAttendances, accessToken, { sort: "attendance_date", order: "desc" }),
      collectAllPages(listLeaveRequests, accessToken, { sort: "created_at", order: "desc" }),
      collectAllPages(listAuditLogs, accessToken, { sort: "created_at", order: "desc" }),
    ]);
    return { dashboard, attendances, leaves, audits };
  }, initialValue);

  return useMemo(() => {
    const latestAttendances = data.attendances.slice(0, 40);
    const lateDepartments = new Set(latestAttendances.filter((item) => item.status.toLowerCase() === "late").map((item) => item.department_name).filter(Boolean));
    const remoteCount = latestAttendances.filter((item) => item.status.toLowerCase() === "remote").length;
    const onTimeCount = latestAttendances.filter((item) => ["on time", "present"].includes(item.status.toLowerCase())).length;

    return {
      stats: data.dashboard ? toAdminStats(data.dashboard) : emptyStats,
      attendanceBars: attendanceBarsFromRecords(latestAttendances),
      attendanceMetrics: [
        { label: "On-time entries", value: `${onTimeCount}` },
        { label: "Late departments", value: `${lateDepartments.size}` },
        { label: "Remote check-ins", value: `${remoteCount}` },
        { label: "Pending approvals", value: `${data.leaves.filter((item) => ["pending", "review"].includes(item.status.toLowerCase())).length}` },
      ],
      recentActivity: auditActivities(data.audits),
      priorityApprovals: leaveActions(data.leaves.filter((item) => ["pending", "review"].includes(item.status.toLowerCase()))),
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useManagerDashboardInsights(): ManagerDashboardInsights {
  const initialValue = useMemo(() => ({
    dashboard: null as DashboardResponse | null,
    employees: [] as EmployeeResource[],
    attendances: [] as AttendanceResource[],
    leaves: [] as LeaveResource[],
  }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => {
    const [dashboard, employees, attendances, leaves] = await Promise.all([
      getDashboard(accessToken, "manager"),
      collectAllPages(listEmployees, accessToken, { sort: "created_at", order: "desc" }),
      collectAllPages(listAttendances, accessToken, { sort: "attendance_date", order: "desc" }),
      collectAllPages(listLeaveRequests, accessToken, { sort: "created_at", order: "desc" }),
    ]);
    return { dashboard, employees, attendances, leaves };
  }, initialValue);

  return useMemo(() => {
    const lateCount = data.attendances.filter((item) => item.status.toLowerCase() === "late").length;
    const remoteCount = data.attendances.filter((item) => item.status.toLowerCase() === "remote").length;
    const activeDepartments = new Set(data.employees.map((item) => item.department_name).filter(Boolean)).size;
    const pending = data.leaves.filter((item) => item.status.toLowerCase() === "pending").length;
    const review = data.leaves.filter((item) => item.status.toLowerCase() === "review").length;
    const approved = data.leaves.filter((item) => item.status.toLowerCase() === "approved").length;

    return {
      stats: data.dashboard ? toManagerStats(data.dashboard) : emptyStats,
      overview: [
        { label: "Present today", value: `${data.dashboard?.today_present ?? "--"} / ${data.dashboard?.team_members ?? "--"}` },
        { label: "Late arrivals", value: `${lateCount}` },
        { label: "Remote active", value: `${remoteCount}` },
        { label: "Active departments", value: `${activeDepartments}` },
      ],
      approvals: [
        { label: "Pending", value: `${pending}` },
        { label: "Review", value: `${review}` },
        { label: "Approved", value: `${approved}` },
        { label: "Coverage watch", value: `${new Set(data.leaves.filter((item) => ["pending", "review"].includes(item.status.toLowerCase())).map((item) => item.department_name).filter(Boolean)).size}` },
      ],
      recentActivity: [...leaveActivities(data.leaves.slice(0, 2)), ...attendanceActivities(data.attendances.slice(0, 2))].slice(0, 4),
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useEmployeeDashboardInsights(): EmployeeDashboardInsights {
  const initialValue = useMemo(() => ({
    dashboard: null as DashboardResponse | null,
    profile: null as ProfileResponse | null,
    attendances: [] as AttendanceResource[],
    leaves: [] as LeaveResource[],
  }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => {
    const [dashboard, profile, attendances, leaves] = await Promise.all([
      getDashboard(accessToken, "employee"),
      getProfile(accessToken),
      collectAllPages(listAttendances, accessToken, { sort: "attendance_date", order: "desc" }),
      collectAllPages(listLeaveRequests, accessToken, { sort: "created_at", order: "desc" }),
    ]);
    return { dashboard, profile, attendances, leaves };
  }, initialValue);

  return useMemo(() => {
    const employee = data.profile?.employee;
    const latestAttendance = data.attendances[0];
    const approvedLeaves = data.leaves.filter((item) => item.status.toLowerCase() === "approved").length;
    const pendingLeaves = data.leaves.filter((item) => ["pending", "review"].includes(item.status.toLowerCase())).length;
    const healthyAttendance = data.attendances.filter((item) => ["on time", "present", "remote"].includes(item.status.toLowerCase())).length;

    return {
      stats: data.dashboard ? toEmployeeStats(data.dashboard) : emptyStats,
      overview: [
        { label: "Latest attendance", value: latestAttendance ? `${formatShortDate(latestAttendance.attendance_date)} · ${titleCase(latestAttendance.status)}` : "--" },
        { label: "Manager", value: employee?.manager_name ?? "--" },
        { label: "Office", value: employee?.location_name ?? "--" },
        { label: "Primary role", value: employee?.job_title ?? "--" },
      ],
      balance: [
        { label: "Approved leaves", value: `${approvedLeaves}` },
        { label: "Pending approvals", value: `${pendingLeaves}` },
        { label: "Healthy entries", value: `${healthyAttendance}` },
        { label: "Secondary roles", value: `${data.dashboard?.active_secondary_assignments ?? 0}` },
      ],
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminEmployeesInsights(): PageSummaryInsights {
  const initialValue = useMemo(() => ({
    employees: [] as EmployeeResource[],
    assignments: [] as AssignmentResource[],
  }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => {
    const [employees, assignments] = await Promise.all([
      collectAllPages(listEmployees, accessToken, { sort: "created_at", order: "desc" }),
      collectAllPages(listAssignments, accessToken, { sort: "created_at", order: "desc" }),
    ]);
    return { employees, assignments };
  }, initialValue);

  return useMemo(() => {
    const active = data.employees.filter((item) => item.employment_status_name?.toLowerCase() === "active").length;
    const hiredThisYear = data.employees.filter((item) => new Date(item.hire_date).getFullYear() === new Date().getFullYear()).length;
    const probation = data.employees.filter((item) => item.employment_status_name?.toLowerCase() === "probation").length;
    const locationCoverage = new Set(data.employees.map((item) => item.location_name).filter(Boolean)).size;
    const managerMapped = data.employees.filter((item) => item.manager_name).length;

    return {
      stats: [
        { label: "Active employees", value: `${active}`, detail: "Employees currently active in PT. CODEID." },
        { label: "Hired this year", value: `${hiredThisYear}`, detail: "Newly hired records this calendar year.", tone: "accent" },
        { label: "Secondary assignments", value: `${data.assignments.length}`, detail: "Flexible staffing roles currently tracked.", tone: "warning" },
        { label: "Locations represented", value: `${locationCoverage}`, detail: "Distinct work locations across employee records.", tone: "cool" },
      ],
      miniStats: [
        { label: "Primary roles", value: `${data.employees.length}` },
        { label: "Secondary roles", value: `${data.assignments.length}` },
        { label: "Probation", value: `${probation}` },
        { label: "Manager links", value: `${managerMapped}/${data.employees.length || 0}` },
      ],
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminDepartmentsInsights(): PageSummaryInsights {
  const initialValue = useMemo(() => ({ departments: [] as DepartmentResource[] }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => ({
    departments: await collectAllPages(listDepartments, accessToken, { sort: "name", order: "asc" }),
  }), initialValue);

  return useMemo(() => {
    const managerCoverage = data.departments.filter((item) => item.manager_name).length;
    const levelTwo = data.departments.filter((item) => item.level === 2).length;
    const levelOne = data.departments.filter((item) => item.level === 1).length;
    const locationCoverage = new Set(data.departments.map((item) => item.location_name).filter(Boolean)).size;
    const root = data.departments.find((item) => item.level === 0)?.name ?? "PT. CODEID";

    return {
      stats: [
        { label: "Total departments", value: `${data.departments.length}`, detail: "Active PT. CODEID department nodes." },
        { label: "Assigned managers", value: `${managerCoverage}`, detail: "Departments with a mapped manager.", tone: "accent" },
        { label: "Sub-departments", value: `${levelTwo}`, detail: "Level 2 units currently tracked.", tone: "warning" },
        { label: "Location coverage", value: `${locationCoverage}`, detail: "Locations represented in the org tree.", tone: "cool" },
      ],
      miniStats: [
        { label: "Root node", value: root },
        { label: "Level 1", value: `${levelOne}` },
        { label: "Level 2", value: `${levelTwo}` },
        { label: "Managers mapped", value: `${managerCoverage}/${data.departments.length || 0}` },
      ],
      distribution: topDistributions(countBy(data.departments, (item) => item.location_name), 4),
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminAttendanceInsights(): PageSummaryInsights & { attendanceBars: BarItem[] } {
  const initialValue = useMemo(() => ({ attendances: [] as AttendanceResource[] }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => ({
    attendances: await collectAllPages(listAttendances, accessToken, { sort: "attendance_date", order: "desc" }),
  }), initialValue);

  return useMemo(() => {
    const latestDate = [...data.attendances.map((item) => item.attendance_date)].sort().at(-1);
    const currentEntries = latestDate ? data.attendances.filter((item) => item.attendance_date === latestDate) : data.attendances;
    const late = currentEntries.filter((item) => item.status.toLowerCase() === "late").length;
    const remote = currentEntries.filter((item) => item.status.toLowerCase() === "remote").length;
    const openIssues = currentEntries.filter((item) => item.status.toLowerCase() === "late" || !item.check_out_at).length;
    const locations = new Set(currentEntries.map((item) => item.location_name).filter(Boolean)).size;

    return {
      stats: [
        { label: "Present today", value: `${currentEntries.length}`, detail: "Attendance entries recorded for the latest tracked workday." },
        { label: "Late arrivals", value: `${late}`, detail: "Late check-ins needing review.", tone: "accent" },
        { label: "Remote check-ins", value: `${remote}`, detail: "Remote attendance entries recorded.", tone: "warning" },
        { label: "Exceptions open", value: `${openIssues}`, detail: "Late or incomplete attendance records.", tone: "cool" },
      ],
      miniStats: [
        { label: "Late", value: `${late}` },
        { label: "Remote", value: `${remote}` },
        { label: "Missing checkout", value: `${currentEntries.filter((item) => !item.check_out_at).length}` },
        { label: "Locations active", value: `${locations}` },
      ],
      activities: attendanceActivities(data.attendances),
      attendanceBars: attendanceBarsFromRecords(data.attendances),
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminLeaveInsights(): PageSummaryInsights {
  const initialValue = useMemo(() => ({ leaves: [] as LeaveResource[] }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => ({
    leaves: await collectAllPages(listLeaveRequests, accessToken, { sort: "created_at", order: "desc" }),
  }), initialValue);

  return useMemo(() => {
    const pending = data.leaves.filter((item) => ["pending", "review"].includes(item.status.toLowerCase())).length;
    const approved = data.leaves.filter((item) => item.status.toLowerCase() === "approved").length;
    const rejected = data.leaves.filter((item) => ["rejected", "cancelled"].includes(item.status.toLowerCase())).length;

    return {
      stats: [
        { label: "Open requests", value: `${pending}`, detail: "Leave requests awaiting action." },
        { label: "Approved", value: `${approved}`, detail: "Requests approved in current dataset.", tone: "accent" },
        { label: "Closed negative", value: `${rejected}`, detail: "Rejected or cancelled requests.", tone: "warning" },
        { label: "Leave types used", value: `${new Set(data.leaves.map((item) => item.leave_type_name).filter(Boolean)).size}`, detail: "Leave types actively appearing in requests.", tone: "cool" },
      ],
      miniStats: [
        { label: "Pending", value: `${data.leaves.filter((item) => item.status.toLowerCase() === "pending").length}` },
        { label: "Review", value: `${data.leaves.filter((item) => item.status.toLowerCase() === "review").length}` },
        { label: "Approved", value: `${approved}` },
        { label: "Rejected", value: `${rejected}` },
      ],
      actions: leaveActions(data.leaves.filter((item) => ["pending", "review"].includes(item.status.toLowerCase()))),
      distribution: topDistributions(countBy(data.leaves, (item) => item.leave_type_name), 4),
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminJobsInsights(): PageSummaryInsights {
  const initialValue = useMemo(() => ({ jobs: [] as JobResource[] }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => ({
    jobs: await collectAllPages(listJobs, accessToken, { sort: "title", order: "asc" }),
  }), initialValue);

  return useMemo(() => {
    const salaryBands = data.jobs.filter((item) => item.min_salary != null && item.max_salary != null).length;

    return {
      stats: [
        { label: "Active jobs", value: `${data.jobs.length}`, detail: "Current PT. CODEID job catalog entries." },
        { label: "Departments linked", value: `${new Set(data.jobs.map((item) => item.primary_department_name).filter(Boolean)).size}`, detail: "Departments owning active roles.", tone: "accent" },
        { label: "Levels used", value: `${new Set(data.jobs.map((item) => item.job_level_name).filter(Boolean)).size}`, detail: "Distinct job levels in active use.", tone: "warning" },
        { label: "Salary bands set", value: `${salaryBands}`, detail: "Roles with both minimum and maximum salary mapped.", tone: "cool" },
      ],
      miniStats: [
        { label: "Roles with descriptions", value: `${data.jobs.filter((item) => (item.job_description ?? "").trim() !== "").length}` },
        { label: "Departments", value: `${new Set(data.jobs.map((item) => item.primary_department_name).filter(Boolean)).size}` },
        { label: "Levels", value: `${new Set(data.jobs.map((item) => item.job_level_name).filter(Boolean)).size}` },
        { label: "Salary bands", value: `${salaryBands}` },
      ],
      distribution: topDistributions(countBy(data.jobs, (item) => item.primary_department_name), 4),
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminHolidayInsights(): PageSummaryInsights {
  const initialValue = useMemo(() => ({ holidays: [] as HolidayResource[] }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => ({
    holidays: await collectAllPages(listHolidays, accessToken, { sort: "holiday_date", order: "asc" }),
  }), initialValue);

  return useMemo(() => {
    const currentYear = new Date().getFullYear();
    const currentYearHolidays = data.holidays.filter((item) => item.year === currentYear);
    const nextHoliday = [...data.holidays]
      .filter((item) => new Date(item.holiday_date).getTime() >= Date.now())
      .sort((left, right) => new Date(left.holiday_date).getTime() - new Date(right.holiday_date).getTime())[0];

    return {
      stats: [
        { label: `${currentYear} holidays`, value: `${currentYearHolidays.length}`, detail: "Holiday records for the current year." },
        { label: "Company-wide", value: `${data.holidays.filter((item) => !item.location_name).length}`, detail: "Applies across PT. CODEID.", tone: "accent" },
        { label: "Location scoped", value: `${data.holidays.filter((item) => item.location_name).length}`, detail: "Branch or location-specific holiday entries.", tone: "warning" },
        { label: "Years planned", value: `${new Set(data.holidays.map((item) => item.year)).size}`, detail: "Distinct calendar years currently planned.", tone: "cool" },
      ],
      miniStats: [
        { label: "Current year", value: `${currentYearHolidays.length}` },
        { label: "Company-wide", value: `${data.holidays.filter((item) => !item.location_name).length}` },
        { label: "Location scoped", value: `${data.holidays.filter((item) => item.location_name).length}` },
        { label: "Next holiday", value: nextHoliday ? formatShortDate(nextHoliday.holiday_date) : "--" },
      ],
      notice: nextHoliday
        ? { text: `${nextHoliday.name} is the next upcoming holiday on ${formatShortDate(nextHoliday.holiday_date)}.`, action: nextHoliday.location_name ?? "Company-wide" }
        : { text: "Add future-year holidays as PT. CODEID plans the next calendar cycle.", action: "Planning ready" },
      loading,
      error,
    };
  }, [data, error, loading]);
}

export function useAdminAuditInsights(): PageSummaryInsights {
  const initialValue = useMemo(() => ({ logs: [] as AuditLogResource[] }), []);

  const { data, loading, error } = useAuthorizedLoader(async (accessToken) => ({
    logs: await collectAllPages(listAuditLogs, accessToken, { sort: "created_at", order: "desc" }),
  }), initialValue);

  return useMemo(() => {
    const highPriority = data.logs.filter((item) => item.action.toLowerCase().includes("delete")).length;
    return {
      stats: [
        { label: "Events tracked", value: `${data.logs.length}`, detail: "Audit events currently returned from the backend." },
        { label: "High priority", value: `${highPriority}`, detail: "Delete actions and other sensitive writes.", tone: "accent" },
        { label: "Actors seen", value: `${new Set(data.logs.map((item) => item.actor_email).filter(Boolean)).size}`, detail: "Distinct user actors in the audit stream.", tone: "warning" },
        { label: "Entities covered", value: `${new Set(data.logs.map((item) => item.entity).filter(Boolean)).size}`, detail: "Tracked resource entities in the stream.", tone: "cool" },
      ],
      miniStats: [
        { label: "Latest event", value: data.logs[0] ? formatDateTime(data.logs[0].created_at) : "--" },
        { label: "High priority", value: `${highPriority}` },
        { label: "Actors", value: `${new Set(data.logs.map((item) => item.actor_email).filter(Boolean)).size}` },
        { label: "Entities", value: `${new Set(data.logs.map((item) => item.entity).filter(Boolean)).size}` },
      ],
      distribution: topDistributions(countBy(data.logs, (item) => item.entity), 4),
      activities: auditActivities(data.logs),
      loading,
      error,
    };
  }, [data, error, loading]);
}
