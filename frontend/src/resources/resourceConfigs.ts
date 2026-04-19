import {
  createAssignment,
  createAttendance,
  createDepartment,
  createEmployee,
  createHoliday,
  createJob,
  createLeaveRequest,
  deleteAssignment,
  deleteAttendance,
  deleteDepartment,
  deleteEmployee,
  deleteHoliday,
  deleteJob,
  deleteLeaveRequest,
  getAssignment,
  getAttendance,
  getDepartment,
  getEmployee,
  getHoliday,
  getJob,
  getLeaveRequest,
  listAssignments,
  listAttendances,
  listDepartments,
  listEmployees,
  listHolidays,
  listJobs,
  listLeaveRequests,
  updateAssignment,
  updateAttendance,
  updateDepartment,
  updateEmployee,
  updateHoliday,
  updateJob,
  updateLeaveRequest,
  type PaginatedResponse,
  type ResourceListQuery,
  type AssignmentPayload,
  type AssignmentResource,
  type AttendancePayload,
  type AttendanceResource,
  type DepartmentPayload,
  type DepartmentResource,
  type EmployeePayload,
  type EmployeeResource,
  type HolidayPayload,
  type HolidayResource,
  type JobPayload,
  type JobResource,
  type LeavePayload,
  type LeaveResource,
} from "../api/resources";
import type { CrudField, CrudRow, LookupKey } from "../types";

export type DetailField = {
  label: string;
  key: string;
};

export type ApiCrudConfig<TResource, TPayload> = {
  recordLabel: string;
  fields: CrudField[];
  createDefaults: Record<string, string>;
  lookupKeys?: LookupKey[];
  list: (accessToken: string, query?: ResourceListQuery) => Promise<PaginatedResponse<TResource>>;
  get: (accessToken: string, id: string) => Promise<TResource>;
  create: (accessToken: string, payload: TPayload) => Promise<TResource>;
  update: (accessToken: string, id: string, payload: TPayload) => Promise<TResource>;
  remove: (accessToken: string, id: string) => Promise<void>;
  toRow: (resource: TResource) => CrudRow;
  toPayload: (values: Record<string, string>) => TPayload;
  detailFields?: DetailField[];
};

function initialsFrom(name: string) {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join("");
}

function titleCase(value: string) {
  return value
    .split(/[\s_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatDate(value?: string | null) {
  if (!value) return "--";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short", year: "numeric" }).format(parsed);
}

function formatShortDate(value?: string | null) {
  if (!value) return "--";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short" }).format(parsed);
}

function formatTime(value?: string | null) {
  if (!value) return "--";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat("en-GB", { hour: "2-digit", minute: "2-digit", hour12: false }).format(parsed);
}

function formatCurrency(value?: number | null) {
  if (value == null) return "--";
  return new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(value);
}

function formatRange(start: string, end: string) {
  if (!start && !end) return "--";
  if (start === end) return formatDate(start);
  return `${formatDate(start)} - ${formatDate(end)}`;
}

function optionalValue(value?: string | null) {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? null : normalized;
}

function optionalNumber(value?: string | null) {
  const normalized = value?.trim() ?? "";
  if (normalized === "") return null;
  const parsed = Number(normalized);
  return Number.isNaN(parsed) ? null : parsed;
}

function requiredNumber(value?: string | null) {
  const parsed = Number(value ?? "");
  return Number.isNaN(parsed) ? 0 : parsed;
}

function requiredText(value?: string | null) {
  return value?.trim() ?? "";
}

function optionalText(value?: string | null) {
  return value?.trim() ?? "";
}

function optionalIsoDateTime(date?: string | null, time?: string | null) {
  if (!date || !time) return null;
  const parsed = new Date(`${date}T${time}:00`);
  if (Number.isNaN(parsed.getTime())) return null;
  return parsed.toISOString();
}

function optionalNotes(value?: string | null) {
  const normalized = optionalText(value);
  return normalized === "" ? "" : normalized;
}

function createListLoader<TResource>(loader: (accessToken: string, query?: ResourceListQuery) => Promise<PaginatedResponse<TResource>>) {
  return loader;
}

function optionalSelectValue(value?: string | null) {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? undefined : normalized;
}

function toneFromStatus(status?: string) {
  const value = (status ?? "").toLowerCase();
  if (["late", "pending", "review", "rejected", "inactive", "probation", "on leave"].includes(value)) return "warning" as const;
  if (["remote", "approved"].includes(value)) return "cool" as const;
  return "default" as const;
}

function pillToneFromStatus(status?: string) {
  const value = (status ?? "").toLowerCase();
  if (["late", "pending", "review", "rejected", "inactive", "probation", "on leave"].includes(value)) return "warning" as const;
  if (["remote", "approved"].includes(value)) return "cool" as const;
  return "teal" as const;
}

function isFinalLeaveStatus(status?: string) {
  const value = (status ?? "").toLowerCase();
  return value === "approved" || value === "rejected";
}

function assignmentWindowLabel(start: string, end?: string | null) {
  return formatRange(start, end ?? start);
}

function assignmentStatusTone(status?: string) {
  const value = (status ?? "").toLowerCase();
  if (value === "active") return "cool" as const;
  if (value === "scheduled") return "teal" as const;
  return "soft" as const;
}

function managementScopeLabel(value?: string | null) {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? "--" : titleCase(normalized);
}

function employeeFormValues(resource: EmployeeResource) {
  return {
    first_name: resource.first_name,
    last_name: resource.last_name,
    employee_code: resource.employee_code,
    email: resource.email,
    phone_number: resource.phone_number ?? "",
    hire_date: resource.hire_date,
    salary: resource.salary?.toString() ?? "",
    employee_type_id: resource.employee_type_id,
    employee_type_name: resource.employee_type_name ?? "",
    employment_status_id: resource.employment_status_id,
    employment_status_name: resource.employment_status_name ?? "",
    department_id: resource.department_id,
    department_name: resource.department_name ?? "",
    job_id: resource.job_id,
    job_title: resource.job_title ?? "",
    job_level_name: resource.job_level_name ?? "",
    location_id: resource.location_id,
    location_name: resource.location_name ?? "",
    work_mode: resource.work_mode ?? "onsite",
    management_scope: resource.management_scope ?? "individual_contributor",
    manager_employee_id: resource.manager_employee_id ?? "",
    manager_name: resource.manager_name ?? "",
  };
}

export const employeeFields: CrudField[] = [
  { name: "first_name", label: "First name", required: true },
  { name: "last_name", label: "Last name", required: true },
  { name: "employee_code", label: "Employee code", required: true },
  { name: "email", label: "Email", type: "email", required: true },
  { name: "phone_number", label: "Phone number" },
  { name: "hire_date", label: "Hire date", type: "date", required: true },
  { name: "salary", label: "Salary", type: "number", placeholder: "15000000" },
  { name: "employee_type_id", label: "Employee type", type: "select", required: true, lookupKey: "employeeTypes", emptyOptionLabel: "Choose employee type" },
  { name: "employment_status_id", label: "Employment status", type: "select", required: true, lookupKey: "employmentStatuses", emptyOptionLabel: "Choose status" },
  { name: "department_id", label: "Department", type: "select", required: true, lookupKey: "departments", emptyOptionLabel: "Choose a department" },
  { name: "job_id", label: "Primary job", type: "select", required: true, lookupKey: "jobs", emptyOptionLabel: "Choose a primary job", filterByField: "department_id", filterContextKey: "primary_department_id" },
  { name: "location_id", label: "Location", type: "select", required: true, lookupKey: "locations", emptyOptionLabel: "Choose a location" },
  {
    name: "work_mode",
    label: "Work mode",
    type: "select",
    required: true,
    options: [
      { label: "Onsite", value: "onsite" },
      { label: "Hybrid", value: "hybrid" },
      { label: "Remote", value: "remote" },
      { label: "Client-based", value: "client-based" },
    ],
  },
  {
    name: "management_scope",
    label: "Management scope",
    type: "select",
    required: true,
    options: [
      { label: "Individual contributor", value: "individual_contributor" },
      { label: "Division manager", value: "division_manager" },
      { label: "Subdepartment manager", value: "subdepartment_manager" },
      { label: "Team manager", value: "team_manager" },
    ],
  },
  { name: "manager_employee_id", label: "Manager", type: "select", lookupKey: "employees", emptyOptionLabel: "No manager assigned" },
];

export const employeeConfig: ApiCrudConfig<EmployeeResource, EmployeePayload> = {
  recordLabel: "employee",
  fields: employeeFields,
  createDefaults: {
    first_name: "",
    last_name: "",
    employee_code: "",
    email: "",
    phone_number: "",
    hire_date: "",
    salary: "",
    employee_type_id: "",
    employment_status_id: "",
    department_id: "",
    job_id: "",
    location_id: "",
    work_mode: "onsite",
    management_scope: "individual_contributor",
    manager_employee_id: "",
  },
  lookupKeys: ["employees", "departments", "jobs", "locations", "employeeTypes", "employmentStatuses"],
  list: createListLoader(listEmployees),
  get: getEmployee,
  create: createEmployee,
  update: updateEmployee,
  remove: async (accessToken, id) => { await deleteEmployee(accessToken, id); },
  toRow: (resource) => ({
    id: resource.id,
    initials: initialsFrom(resource.full_name),
    name: resource.full_name,
    meta: `${resource.employee_code} · ${resource.email}`,
    cols: [
      resource.department_name ?? "--",
      resource.job_title ?? "--",
      resource.job_level_name ?? "--",
      managementScopeLabel(resource.management_scope),
      resource.employment_status_name ?? "--",
      `${resource.location_name ?? "--"} · ${titleCase(resource.work_mode ?? "onsite")}`,
      formatShortDate(resource.updated_at),
    ],
    pillIndex: 4,
    pillTone: pillToneFromStatus(resource.employment_status_name),
    tone: toneFromStatus(resource.employment_status_name),
    formValues: employeeFormValues(resource),
  }),
  toPayload: (values) => ({
    first_name: requiredText(values.first_name),
    last_name: requiredText(values.last_name),
    employee_code: requiredText(values.employee_code),
    email: requiredText(values.email),
    phone_number: optionalText(values.phone_number),
    hire_date: values.hire_date,
    salary: optionalNumber(values.salary),
    employee_type_id: values.employee_type_id,
    employment_status_id: values.employment_status_id,
    department_id: values.department_id,
    job_id: values.job_id,
    location_id: values.location_id,
    work_mode: values.work_mode,
    management_scope: values.management_scope,
    manager_employee_id: optionalValue(values.manager_employee_id),
  }),
  detailFields: [
    { label: "Employee code", key: "employee_code" },
    { label: "Email", key: "email" },
    { label: "Phone number", key: "phone_number" },
    { label: "Hire date", key: "hire_date" },
    { label: "Employee type", key: "employee_type_name" },
    { label: "Employment status", key: "employment_status_name" },
    { label: "Department", key: "department_name" },
    { label: "Primary job", key: "job_title" },
    { label: "Job level", key: "job_level_name" },
    { label: "Location", key: "location_name" },
    { label: "Work mode", key: "work_mode" },
    { label: "Management scope", key: "management_scope" },
    { label: "Manager", key: "manager_name" },
  ],
};

export const managerEmployeeConfig: ApiCrudConfig<EmployeeResource, EmployeePayload> = {
  ...employeeConfig,
  fields: employeeFields.filter((field) => field.name !== "manager_employee_id"),
};

export const departmentConfig: ApiCrudConfig<DepartmentResource, DepartmentPayload> = {
  recordLabel: "department",
  fields: [
    { name: "name", label: "Department name", required: true },
    {
      name: "level",
      label: "Level",
      type: "select",
      required: true,
      options: [
        { label: "Level 0", value: "0" },
        { label: "Level 1", value: "1" },
        { label: "Level 2", value: "2" },
      ],
    },
    { name: "location_id", label: "Location", type: "select", required: true, lookupKey: "locations", emptyOptionLabel: "Choose a location" },
    { name: "manager_employee_id", label: "Manager", type: "select", lookupKey: "employees", emptyOptionLabel: "No manager assigned" },
    { name: "parent_department_id", label: "Parent department", type: "select", lookupKey: "departments", emptyOptionLabel: "Top-level department" },
  ],
  createDefaults: { name: "", level: "1", location_id: "", manager_employee_id: "", parent_department_id: "" },
  lookupKeys: ["employees", "departments", "locations"],
  list: createListLoader(listDepartments),
  get: getDepartment,
  create: createDepartment,
  update: updateDepartment,
  remove: async (accessToken, id) => { await deleteDepartment(accessToken, id); },
  toRow: (resource) => ({
    id: resource.id,
    initials: initialsFrom(resource.name),
    name: resource.name,
    meta: resource.parent_department_name ? `Child of ${resource.parent_department_name}` : "Root or top-level department",
    cols: [
      resource.manager_name ?? "Unassigned",
      resource.location_name ?? "--",
      `Level ${resource.level}`,
      resource.parent_department_name ?? "--",
      formatShortDate(resource.updated_at),
    ],
    tone: resource.manager_employee_id ? "default" : "warning",
    formValues: {
      name: resource.name,
      level: String(resource.level),
      location_id: resource.location_id,
      location_name: resource.location_name ?? "",
      manager_employee_id: resource.manager_employee_id ?? "",
      manager_name: resource.manager_name ?? "",
      parent_department_id: resource.parent_department_id ?? "",
      parent_department_name: resource.parent_department_name ?? "",
    },
  }),
  toPayload: (values) => ({
    name: requiredText(values.name),
    level: requiredNumber(values.level),
    location_id: values.location_id,
    manager_employee_id: optionalValue(values.manager_employee_id),
    parent_department_id: optionalValue(values.parent_department_id),
  }),
  detailFields: [
    { label: "Department", key: "name" },
    { label: "Level", key: "level" },
    { label: "Location", key: "location_name" },
    { label: "Manager", key: "manager_name" },
    { label: "Parent department", key: "parent_department_name" },
  ],
};

export const jobConfig: ApiCrudConfig<JobResource, JobPayload> = {
  recordLabel: "job",
  fields: [
    { name: "title", label: "Job title", required: true },
    { name: "primary_department_id", label: "Primary department", type: "select", required: true, lookupKey: "departments", emptyOptionLabel: "Choose a department" },
    { name: "job_level_id", label: "Job level", type: "select", required: true, lookupKey: "jobLevels", emptyOptionLabel: "Choose a level" },
    { name: "min_salary", label: "Minimum salary", type: "number" },
    { name: "max_salary", label: "Maximum salary", type: "number" },
    { name: "job_description", label: "Job description", type: "textarea", placeholder: "Describe the role, scope, and responsibilities." },
  ],
  createDefaults: { title: "", primary_department_id: "", job_level_id: "", min_salary: "", max_salary: "", job_description: "" },
  lookupKeys: ["departments", "jobLevels"],
  list: createListLoader(listJobs),
  get: getJob,
  create: createJob,
  update: updateJob,
  remove: async (accessToken, id) => { await deleteJob(accessToken, id); },
  toRow: (resource) => ({
    id: resource.id,
    initials: initialsFrom(resource.title),
    name: resource.title,
    meta: resource.job_description?.trim() || "Reusable role catalog entry",
    cols: [
      resource.primary_department_name ?? "--",
      resource.job_level_name ?? "--",
      `${formatCurrency(resource.min_salary)} - ${formatCurrency(resource.max_salary)}`,
      formatShortDate(resource.updated_at),
    ],
    tone: "default",
    formValues: {
      title: resource.title,
      primary_department_id: resource.primary_department_id,
      primary_department_name: resource.primary_department_name ?? "",
      job_level_id: resource.job_level_id,
      job_level_name: resource.job_level_name ?? "",
      min_salary: resource.min_salary?.toString() ?? "",
      max_salary: resource.max_salary?.toString() ?? "",
      job_description: resource.job_description ?? "",
    },
  }),
  toPayload: (values) => ({
    title: requiredText(values.title),
    primary_department_id: values.primary_department_id,
    job_level_id: values.job_level_id,
    min_salary: optionalNumber(values.min_salary),
    max_salary: optionalNumber(values.max_salary),
    job_description: optionalText(values.job_description),
  }),
  detailFields: [
    { label: "Job title", key: "title" },
    { label: "Primary department", key: "primary_department_name" },
    { label: "Job level", key: "job_level_name" },
    { label: "Minimum salary", key: "min_salary" },
    { label: "Maximum salary", key: "max_salary" },
    { label: "Job description", key: "job_description" },
  ],
};

function attendanceFormValues(resource: AttendanceResource) {
  return {
    employee_id: resource.employee_id,
    attendance_date: resource.attendance_date,
    check_in_time: resource.check_in_at ? formatTime(resource.check_in_at) : "",
    check_out_time: resource.check_out_at ? formatTime(resource.check_out_at) : "",
    status: resource.status,
    notes: resource.notes ?? "",
  };
}

const attendanceStatusOptions = [
  { label: "On time", value: "on time" },
  { label: "Late", value: "late" },
  { label: "Remote", value: "remote" },
  { label: "Present", value: "present" },
];

export const attendanceFieldsWithEmployee: CrudField[] = [
  { name: "employee_id", label: "Employee", type: "select", required: true, lookupKey: "employees", emptyOptionLabel: "Choose an employee" },
  { name: "attendance_date", label: "Attendance date", type: "date", required: true },
  { name: "check_in_time", label: "Check in", type: "time" },
  { name: "check_out_time", label: "Check out", type: "time" },
  { name: "status", label: "Status", type: "select", required: true, options: attendanceStatusOptions },
  { name: "notes", label: "Notes", type: "textarea", placeholder: "Optional attendance context" },
];

export const attendanceFieldsSelf: CrudField[] = attendanceFieldsWithEmployee.filter((field) => field.name !== "employee_id");

export const adminAttendanceConfig: ApiCrudConfig<AttendanceResource, AttendancePayload> = {
  recordLabel: "attendance record",
  fields: attendanceFieldsWithEmployee,
  createDefaults: { employee_id: "", attendance_date: "", check_in_time: "", check_out_time: "", status: "on time", notes: "" },
  lookupKeys: ["employees"],
  list: createListLoader(listAttendances),
  get: getAttendance,
  create: createAttendance,
  update: updateAttendance,
  remove: async (accessToken, id) => { await deleteAttendance(accessToken, id); },
  toRow: (resource) => ({
    id: resource.id,
    initials: initialsFrom(resource.employee_name),
    name: resource.employee_name,
    meta: `${resource.employee_code ?? "--"} · ${resource.department_name ?? "--"}`,
    cols: [
      formatShortDate(resource.attendance_date),
      formatTime(resource.check_in_at),
      formatTime(resource.check_out_at),
      titleCase(resource.status),
      resource.location_name ?? "--",
    ],
    pillIndex: 3,
    pillTone: pillToneFromStatus(resource.status),
    tone: toneFromStatus(resource.status),
    formValues: attendanceFormValues(resource),
  }),
  toPayload: (values) => ({
    employee_id: optionalSelectValue(values.employee_id),
    attendance_date: values.attendance_date,
    check_in_at: optionalIsoDateTime(values.attendance_date, values.check_in_time),
    check_out_at: optionalIsoDateTime(values.attendance_date, values.check_out_time),
    status: values.status,
    notes: optionalNotes(values.notes),
  }),
};

export const employeeAttendanceConfig: ApiCrudConfig<AttendanceResource, AttendancePayload> = {
  ...adminAttendanceConfig,
  fields: attendanceFieldsSelf,
  createDefaults: { attendance_date: "", check_in_time: "", check_out_time: "", status: "on time", notes: "" },
  lookupKeys: [],
  toRow: (resource) => ({
    ...adminAttendanceConfig.toRow(resource),
    cols: [
      formatShortDate(resource.attendance_date),
      formatTime(resource.check_in_at),
      formatTime(resource.check_out_at),
      titleCase(resource.status),
      resource.location_name ?? "--",
    ],
  }),
  toPayload: (values) => ({
    attendance_date: values.attendance_date,
    check_in_at: optionalIsoDateTime(values.attendance_date, values.check_in_time),
    check_out_at: optionalIsoDateTime(values.attendance_date, values.check_out_time),
    status: values.status,
    notes: optionalNotes(values.notes),
  }),
};

function leaveFormValues(resource: LeaveResource) {
  return {
    employee_id: resource.employee_id,
    approver_employee_id: resource.approver_employee_id ?? "",
    start_date: resource.start_date,
    end_date: resource.end_date,
    leave_type_id: resource.leave_type_id,
    leave_type_name: resource.leave_type_name ?? "",
    reason: resource.reason ?? "",
    status: resource.status,
  };
}

const leaveStatusOptions = [
  { label: "Pending", value: "pending" },
  { label: "Approved", value: "approved" },
  { label: "Rejected", value: "rejected" },
  { label: "Cancelled", value: "cancelled" },
];

const selfServiceLeaveStatusOptions = [
  { label: "Pending", value: "pending" },
  { label: "Cancelled", value: "cancelled" },
];

export const adminLeaveConfig: ApiCrudConfig<LeaveResource, LeavePayload> = {
  recordLabel: "leave request",
  fields: [
    { name: "employee_id", label: "Employee", type: "select", required: true, lookupKey: "employees", emptyOptionLabel: "Choose an employee" },
    { name: "approver_employee_id", label: "Approver", type: "select", lookupKey: "employees", emptyOptionLabel: "No approver assigned" },
    { name: "leave_type_id", label: "Leave type", type: "select", required: true, lookupKey: "leaveTypes", emptyOptionLabel: "Choose leave type" },
    { name: "start_date", label: "Start date", type: "date", required: true },
    { name: "end_date", label: "End date", type: "date", required: true },
    { name: "status", label: "Status", type: "select", required: true, options: leaveStatusOptions },
    { name: "reason", label: "Reason", type: "textarea", placeholder: "Optional approval context" },
  ],
  createDefaults: { employee_id: "", approver_employee_id: "", leave_type_id: "", start_date: "", end_date: "", status: "pending", reason: "" },
  lookupKeys: ["employees", "leaveTypes"],
  list: createListLoader(listLeaveRequests),
  get: getLeaveRequest,
  create: createLeaveRequest,
  update: updateLeaveRequest,
  remove: async (accessToken, id) => { await deleteLeaveRequest(accessToken, id); },
  toRow: (resource) => ({
    id: resource.id,
    initials: initialsFrom(resource.employee_name),
    name: resource.employee_name,
    meta: `${resource.employee_code ?? "--"} · ${resource.leave_type_name ?? "Leave"}`,
    cols: [
      formatRange(resource.start_date, resource.end_date),
      resource.leave_type_name ?? "--",
      titleCase(resource.status),
      resource.approver_name ?? "Unassigned",
      formatShortDate(resource.updated_at),
    ],
    pillIndex: 2,
    pillTone: pillToneFromStatus(resource.status),
    tone: toneFromStatus(resource.status),
    formValues: leaveFormValues(resource),
    canEdit: !isFinalLeaveStatus(resource.status),
    canDelete: !isFinalLeaveStatus(resource.status),
    lockedReason: isFinalLeaveStatus(resource.status) ? "Locked" : undefined,
  }),
  toPayload: (values) => ({
    employee_id: optionalSelectValue(values.employee_id),
    approver_employee_id: optionalValue(values.approver_employee_id),
    leave_type_id: values.leave_type_id,
    start_date: values.start_date,
    end_date: values.end_date,
    reason: optionalNotes(values.reason),
    status: values.status,
  }),
};

export const employeeLeaveConfig: ApiCrudConfig<LeaveResource, LeavePayload> = {
  ...adminLeaveConfig,
  fields: adminLeaveConfig.fields
    .filter((field) => !["employee_id", "approver_employee_id"].includes(field.name))
    .map((field) => (field.name === "status" ? { ...field, options: selfServiceLeaveStatusOptions } : field)),
  createDefaults: { leave_type_id: "", start_date: "", end_date: "", status: "pending", reason: "" },
  lookupKeys: ["leaveTypes"],
  toPayload: (values) => ({
    leave_type_id: values.leave_type_id,
    start_date: values.start_date,
    end_date: values.end_date,
    reason: optionalNotes(values.reason),
    status: values.status,
  }),
};

export const managerLeaveConfig: ApiCrudConfig<LeaveResource, LeavePayload> = {
  ...employeeLeaveConfig,
  list: (accessToken, query) => listLeaveRequests(accessToken, {
    ...query,
    filters: {
      ...(query?.filters ?? {}),
      mine: "true",
    },
  }),
};

export const holidayConfig: ApiCrudConfig<HolidayResource, HolidayPayload> = {
  recordLabel: "holiday",
  fields: [
    { name: "name", label: "Holiday name", required: true },
    { name: "holiday_date", label: "Holiday date", type: "date", required: true },
    { name: "location_id", label: "Location scope", type: "select", lookupKey: "locations", emptyOptionLabel: "Company-wide holiday" },
    { name: "notes", label: "Notes", type: "textarea", placeholder: "Optional holiday notes" },
  ],
  createDefaults: { name: "", holiday_date: "", location_id: "", notes: "" },
  lookupKeys: ["locations"],
  list: createListLoader(listHolidays),
  get: getHoliday,
  create: createHoliday,
  update: updateHoliday,
  remove: async (accessToken, id) => { await deleteHoliday(accessToken, id); },
  toRow: (resource) => ({
    id: resource.id,
    initials: initialsFrom(resource.name),
    name: resource.name,
    meta: resource.location_name ?? "Company-wide holiday",
    cols: [formatDate(resource.holiday_date), String(resource.year), resource.location_name ?? "All locations", formatShortDate(resource.updated_at)],
    tone: "cool",
    formValues: {
      name: resource.name,
      holiday_date: resource.holiday_date,
      location_id: resource.location_id ?? "",
      location_name: resource.location_name ?? "",
      notes: resource.notes ?? "",
      year: String(resource.year),
    },
  }),
  toPayload: (values) => ({
    name: requiredText(values.name),
    holiday_date: values.holiday_date,
    location_id: optionalValue(values.location_id),
    notes: optionalNotes(values.notes),
  }),
  detailFields: [
    { label: "Holiday name", key: "name" },
    { label: "Date", key: "holiday_date" },
    { label: "Year", key: "year" },
    { label: "Location", key: "location_name" },
    { label: "Notes", key: "notes" },
  ],
};

function assignmentFormValues(resource: AssignmentResource) {
  return {
    job_id: resource.job_id,
    job_title: resource.job_title,
    department_id: resource.department_id,
    department_name: resource.department_name,
    estimated_hours_per_week: resource.estimated_hours_per_week.toString(),
    start_date: resource.start_date,
    end_date: resource.end_date ?? "",
    assignment_status: resource.assignment_status ?? "active",
    notes: resource.notes ?? "",
  };
}

export function createAssignmentConfig(employeeId: string): ApiCrudConfig<AssignmentResource, AssignmentPayload> {
  return {
    recordLabel: "secondary assignment",
    fields: [
      { name: "department_id", label: "Department", type: "select", required: true, lookupKey: "departments", emptyOptionLabel: "Choose a department" },
      { name: "job_id", label: "Assignment role", type: "select", required: true, lookupKey: "jobs", emptyOptionLabel: "Choose a job", filterByField: "department_id", filterContextKey: "primary_department_id" },
      { name: "estimated_hours_per_week", label: "Hours per week", type: "number", required: true, placeholder: "8" },
      { name: "start_date", label: "Start date", type: "date", required: true },
      { name: "end_date", label: "End date", type: "date" },
      { name: "notes", label: "Notes", type: "textarea", placeholder: "Scope, reason, or workload context" },
    ],
    createDefaults: { department_id: "", job_id: "", estimated_hours_per_week: "", start_date: "", end_date: "", notes: "" },
    lookupKeys: ["departments", "jobs"],
    list: (accessToken, query) => listAssignments(accessToken, { ...query, filters: { ...(query?.filters ?? {}), employee_id: employeeId } }),
    get: getAssignment,
    create: createAssignment,
    update: updateAssignment,
    remove: async (accessToken, id) => { await deleteAssignment(accessToken, id); },
    toRow: (resource) => ({
      id: resource.id,
      initials: initialsFrom(resource.job_title),
      name: resource.job_title,
      meta: `${resource.department_name} · ${titleCase(resource.assignment_status ?? "active")}`,
      cols: [
        resource.department_name,
        resource.job_level_name ?? "--",
        `${resource.estimated_hours_per_week}h/week`,
        assignmentWindowLabel(resource.start_date, resource.end_date),
        titleCase(resource.assignment_status ?? "active"),
        formatShortDate(resource.updated_at),
      ],
      pillIndex: 4,
      pillTone: assignmentStatusTone(resource.assignment_status),
      tone: resource.assignment_status?.toLowerCase() === "ended" ? "soft" : "cool",
      formValues: assignmentFormValues(resource),
    }),
    toPayload: (values) => ({
      employee_id: employeeId,
      job_id: values.job_id,
      department_id: values.department_id,
      estimated_hours_per_week: optionalNumber(values.estimated_hours_per_week),
      start_date: values.start_date,
      end_date: optionalValue(values.end_date),
      notes: optionalNotes(values.notes),
    }),
  };
}
