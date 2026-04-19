import { apiFetch } from "./client";
import type { LookupKey } from "../types";

export type PaginationMeta = {
  page: number;
  limit: number;
  total: number;
};

export type PaginatedResponse<T> = {
  items: T[];
  meta: PaginationMeta;
};

export type ResourceListQuery = {
  page?: number;
  limit?: number;
  search?: string;
  sort?: string;
  order?: "asc" | "desc";
  filters?: Record<string, string | number | undefined | null>;
};

export type LookupItem = {
  id: string;
  label: string;
  meta?: string;
  context?: Record<string, string | number | boolean | null>;
};

export type SecondaryAssignmentSummary = {
  id: string;
  job_id: string;
  job_title: string;
  job_level_name?: string;
  department_id: string;
  department_name: string;
  estimated_hours_per_week: number;
  start_date: string;
  end_date?: string | null;
  assignment_status?: string;
  notes?: string;
};

export type EmployeeResource = {
  id: string;
  organization_id: string;
  employee_code: string;
  first_name: string;
  last_name: string;
  full_name: string;
  email: string;
  phone_number?: string;
  hire_date: string;
  salary?: number | null;
  employee_type_id: string;
  employee_type_name?: string;
  employment_status_id: string;
  employment_status_name?: string;
  department_id: string;
  department_name?: string;
  job_id: string;
  job_title?: string;
  job_level_id?: string;
  job_level_name?: string;
  location_id: string;
  location_name?: string;
  work_mode?: string;
  management_scope?: string;
  manager_employee_id?: string | null;
  manager_name?: string;
  secondary_assignments?: SecondaryAssignmentSummary[];
  created_at: string;
  updated_at: string;
};

export type DepartmentResource = {
  id: string;
  organization_id: string;
  name: string;
  level: number;
  location_id: string;
  location_name?: string;
  parent_department_id?: string | null;
  parent_department_name?: string;
  manager_employee_id?: string | null;
  manager_name?: string;
  created_at: string;
  updated_at: string;
};

export type JobResource = {
  id: string;
  organization_id: string;
  title: string;
  primary_department_id: string;
  primary_department_name?: string;
  job_level_id: string;
  job_level_name?: string;
  min_salary?: number | null;
  max_salary?: number | null;
  job_description?: string;
  created_at: string;
  updated_at: string;
};

export type AttendanceResource = {
  id: string;
  organization_id: string;
  employee_id: string;
  employee_name: string;
  employee_code?: string;
  department_id?: string;
  department_name?: string;
  location_id?: string;
  location_name?: string;
  attendance_date: string;
  check_in_at?: string | null;
  check_out_at?: string | null;
  status: string;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type LeaveResource = {
  id: string;
  organization_id: string;
  employee_id: string;
  employee_name: string;
  employee_code?: string;
  department_name?: string;
  approver_employee_id?: string | null;
  approver_name?: string;
  start_date: string;
  end_date: string;
  leave_type_id: string;
  leave_type_name?: string;
  reason?: string;
  status: string;
  approved_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type AssignmentResource = {
  id: string;
  organization_id: string;
  employee_id: string;
  employee_name: string;
  job_id: string;
  job_title: string;
  job_level_name?: string;
  department_id: string;
  department_name: string;
  estimated_hours_per_week: number;
  start_date: string;
  end_date?: string | null;
  assignment_status?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type HolidayResource = {
  id: string;
  organization_id: string;
  name: string;
  holiday_date: string;
  year: number;
  location_id?: string | null;
  location_name?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type AuditLogResource = {
  id: string;
  action: string;
  entity: string;
  entity_id?: string | null;
  actor_user_id?: string | null;
  actor_email?: string;
  metadata?: unknown;
  created_at: string;
};

export type EmployeePayload = {
  employee_code: string;
  first_name: string;
  last_name: string;
  email: string;
  phone_number?: string;
  hire_date: string;
  salary?: number | null;
  employee_type_id: string;
  employment_status_id: string;
  department_id: string;
  job_id: string;
  location_id: string;
  work_mode?: string;
  management_scope?: string;
  manager_employee_id?: string | null;
};

export type TeamResource = {
  id: string;
  name: string;
  department_id?: string | null;
  department_name?: string;
  manager_employee_id: string;
  manager_name?: string;
  is_cross_department: boolean;
  focus_area?: string;
  member_count: number;
};

export type TeamMemberResource = {
  id: string;
  team_id: string;
  employee_id: string;
  employee_code: string;
  employee_name: string;
  department_name?: string;
  job_title?: string;
  job_level_name?: string;
  location_name?: string;
  work_mode?: string;
  role_name?: string;
  start_date?: string;
  end_date?: string | null;
  notes?: string;
};

export type TeamAvailableEmployeeResource = {
  id: string;
  employee_code: string;
  full_name: string;
  department_name?: string;
  job_title?: string;
  job_level_name?: string;
  employment_status?: string;
  location_name?: string;
  work_mode?: string;
};

export type DepartmentPayload = {
  name: string;
  level: number;
  location_id: string;
  manager_employee_id?: string | null;
  parent_department_id?: string | null;
};

export type JobPayload = {
  title: string;
  primary_department_id: string;
  job_level_id: string;
  min_salary?: number | null;
  max_salary?: number | null;
  job_description?: string;
};

export type AttendancePayload = {
  employee_id?: string;
  attendance_date: string;
  check_in_at?: string | null;
  check_out_at?: string | null;
  status: string;
  notes?: string;
};

export type LeavePayload = {
  employee_id?: string;
  approver_employee_id?: string | null;
  start_date: string;
  end_date: string;
  leave_type_id: string;
  reason?: string;
  status: string;
};

export type AssignmentPayload = {
  employee_id: string;
  job_id: string;
  department_id: string;
  estimated_hours_per_week?: number | null;
  start_date: string;
  end_date?: string | null;
  notes?: string;
};

export type HolidayPayload = {
  name: string;
  holiday_date: string;
  location_id?: string | null;
  notes?: string;
};

const lookupPaths: Record<LookupKey, string> = {
  employees: "employees",
  departments: "departments",
  jobs: "jobs",
  jobLevels: "job-levels",
  locations: "locations",
  employeeTypes: "employee-types",
  leaveTypes: "leave-types",
  employmentStatuses: "employment-statuses",
  holidays: "holidays",
};

function withQuery(path: string, query?: Record<string, string | number | undefined | null>) {
  if (!query) return path;
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== null && `${value}` !== "") {
      params.set(key, `${value}`);
    }
  });
  const suffix = params.toString();
  return suffix ? `${path}?${suffix}` : path;
}

function buildListPath(path: string, query?: ResourceListQuery, defaults?: { sort?: string; order?: "asc" | "desc" }) {
  return withQuery(path, {
    page: query?.page ?? 1,
    limit: query?.limit ?? 10,
    search: query?.search,
    sort: query?.sort ?? defaults?.sort,
    order: query?.order ?? defaults?.order,
    ...(query?.filters ?? {}),
  });
}

export async function listEmployees(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<EmployeeResource>>(buildListPath("/employees", query, { sort: "created_at", order: "desc" }), { method: "GET" }, accessToken);
}

export async function getEmployee(accessToken: string, id: string) {
  return apiFetch<EmployeeResource>(`/employees/${id}`, { method: "GET" }, accessToken);
}

export async function createEmployee(accessToken: string, payload: EmployeePayload) {
  return apiFetch<EmployeeResource>("/employees", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateEmployee(accessToken: string, id: string, payload: EmployeePayload) {
  return apiFetch<EmployeeResource>(`/employees/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteEmployee(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/employees/${id}`, { method: "DELETE" }, accessToken);
}

export async function listDepartments(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<DepartmentResource>>(buildListPath("/departments", query, { sort: "name", order: "asc" }), { method: "GET" }, accessToken);
}

export async function getDepartment(accessToken: string, id: string) {
  return apiFetch<DepartmentResource>(`/departments/${id}`, { method: "GET" }, accessToken);
}

export async function createDepartment(accessToken: string, payload: DepartmentPayload) {
  return apiFetch<DepartmentResource>("/departments", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateDepartment(accessToken: string, id: string, payload: DepartmentPayload) {
  return apiFetch<DepartmentResource>(`/departments/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteDepartment(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/departments/${id}`, { method: "DELETE" }, accessToken);
}

export async function listJobs(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<JobResource>>(buildListPath("/jobs", query, { sort: "title", order: "asc" }), { method: "GET" }, accessToken);
}

export async function getJob(accessToken: string, id: string) {
  return apiFetch<JobResource>(`/jobs/${id}`, { method: "GET" }, accessToken);
}

export async function createJob(accessToken: string, payload: JobPayload) {
  return apiFetch<JobResource>("/jobs", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateJob(accessToken: string, id: string, payload: JobPayload) {
  return apiFetch<JobResource>(`/jobs/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteJob(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/jobs/${id}`, { method: "DELETE" }, accessToken);
}

export async function listAttendances(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<AttendanceResource>>(buildListPath("/attendances", query, { sort: "attendance_date", order: "desc" }), { method: "GET" }, accessToken);
}

export async function getAttendance(accessToken: string, id: string) {
  return apiFetch<AttendanceResource>(`/attendances/${id}`, { method: "GET" }, accessToken);
}

export async function createAttendance(accessToken: string, payload: AttendancePayload) {
  return apiFetch<AttendanceResource>("/attendances", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateAttendance(accessToken: string, id: string, payload: AttendancePayload) {
  return apiFetch<AttendanceResource>(`/attendances/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteAttendance(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/attendances/${id}`, { method: "DELETE" }, accessToken);
}

export async function listLeaveRequests(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<LeaveResource>>(buildListPath("/leave-requests", query, { sort: "created_at", order: "desc" }), { method: "GET" }, accessToken);
}

export async function getLeaveRequest(accessToken: string, id: string) {
  return apiFetch<LeaveResource>(`/leave-requests/${id}`, { method: "GET" }, accessToken);
}

export async function createLeaveRequest(accessToken: string, payload: LeavePayload) {
  return apiFetch<LeaveResource>("/leave-requests", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateLeaveRequest(accessToken: string, id: string, payload: LeavePayload) {
  return apiFetch<LeaveResource>(`/leave-requests/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteLeaveRequest(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/leave-requests/${id}`, { method: "DELETE" }, accessToken);
}

export async function listAssignments(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<AssignmentResource>>(buildListPath("/employee-role-assignments", query, { sort: "created_at", order: "desc" }), { method: "GET" }, accessToken);
}

export async function getAssignment(accessToken: string, id: string) {
  return apiFetch<AssignmentResource>(`/employee-role-assignments/${id}`, { method: "GET" }, accessToken);
}

export async function createAssignment(accessToken: string, payload: AssignmentPayload) {
  return apiFetch<AssignmentResource>("/employee-role-assignments", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateAssignment(accessToken: string, id: string, payload: AssignmentPayload) {
  return apiFetch<AssignmentResource>(`/employee-role-assignments/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteAssignment(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/employee-role-assignments/${id}`, { method: "DELETE" }, accessToken);
}

export async function listHolidays(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<HolidayResource>>(buildListPath("/holidays", query, { sort: "holiday_date", order: "asc" }), { method: "GET" }, accessToken);
}

export async function listTeams(accessToken: string) {
  return apiFetch<{ items: TeamResource[] }>("/teams", { method: "GET" }, accessToken);
}

export async function listTeamMembers(accessToken: string, teamId: string) {
  return apiFetch<{ team: TeamResource; items: TeamMemberResource[] }>(`/teams/${teamId}/members`, { method: "GET" }, accessToken);
}

export async function listAvailableTeamEmployees(accessToken: string, teamId: string, search?: string) {
  return apiFetch<{ team: TeamResource; items: TeamAvailableEmployeeResource[] }>(withQuery(`/teams/${teamId}/available-employees`, { search }), { method: "GET" }, accessToken);
}

export async function addTeamMember(accessToken: string, teamId: string, payload: { employee_id: string; role_name?: string; notes?: string }) {
  return apiFetch<TeamMemberResource>(`/teams/${teamId}/members`, { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function removeTeamMember(accessToken: string, teamId: string, membershipId: string) {
  return apiFetch<{ message: string }>(`/teams/${teamId}/members/${membershipId}`, { method: "DELETE" }, accessToken);
}

export async function getHoliday(accessToken: string, id: string) {
  return apiFetch<HolidayResource>(`/holidays/${id}`, { method: "GET" }, accessToken);
}

export async function createHoliday(accessToken: string, payload: HolidayPayload) {
  return apiFetch<HolidayResource>("/holidays", { method: "POST", body: JSON.stringify(payload) }, accessToken);
}

export async function updateHoliday(accessToken: string, id: string, payload: HolidayPayload) {
  return apiFetch<HolidayResource>(`/holidays/${id}`, { method: "PUT", body: JSON.stringify(payload) }, accessToken);
}

export async function deleteHoliday(accessToken: string, id: string) {
  return apiFetch<{ message: string }>(`/holidays/${id}`, { method: "DELETE" }, accessToken);
}

export async function listAuditLogs(accessToken: string, query?: ResourceListQuery) {
  return apiFetch<PaginatedResponse<AuditLogResource>>(buildListPath("/audit-logs", query, { sort: "created_at", order: "desc" }), { method: "GET" }, accessToken);
}

export async function listLookup(accessToken: string, key: LookupKey) {
  return apiFetch<{ items: LookupItem[] }>(`/lookups/${lookupPaths[key]}`, { method: "GET" }, accessToken);
}
