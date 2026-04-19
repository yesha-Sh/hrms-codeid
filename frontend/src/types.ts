import type { LucideIcon } from "lucide-react";

export type Role = "admin" | "manager" | "employee";

export type AuthUser = {
  id?: string;
  role: Role;
  email: string;
  employeeId?: string | null;
  name: string;
  title: string;
  department: string;
  workMode?: string;
  managementScope?: string;
};

export type NavItem = {
  label: string;
  path: string;
  icon: LucideIcon;
};

export type Insight = {
  label: string;
  value: string;
  detail: string;
  statusLabel?: string;
  progressA: { label: string; value: string; width: string };
  progressB: { label: string; value: string; width: string };
};

export type PageConfig = {
  sectionLabel: string;
  badge: string;
  rangeLabel: string;
  searchLabel: string;
  insight: Insight;
};

export type StatTone = "default" | "accent" | "warning" | "cool";

export type StatCard = {
  id?: string;
  label: string;
  value: string;
  detail: string;
  tone?: StatTone;
};

export type MiniStatItem = {
  id?: string;
  label: string;
  value: string;
};

export type DistributionItem = {
  label: string;
  value: string;
  width: string;
  tone: "teal" | "blue" | "gold" | "soft";
};

export type ActivityItem = {
  id?: string;
  title: string;
  text: string;
  meta: string;
  tone: "teal" | "gold" | "blue";
};

export type ActionRow = {
  id?: string;
  initials: string;
  name: string;
  detail: string;
  status: string;
  tone: "default" | "warning" | "cool" | "soft";
};

export type TableRow = {
  id?: string;
  initials: string;
  name: string;
  meta: string;
  tone: "default" | "warning" | "cool" | "soft";
  cols: string[];
  pillIndex?: number;
  pillTone?: "teal" | "warning" | "cool" | "soft";
};

export type LookupKey =
  | "employees"
  | "departments"
  | "jobs"
  | "jobLevels"
  | "locations"
  | "employeeTypes"
  | "leaveTypes"
  | "employmentStatuses"
  | "holidays";

export type CrudField = {
  name: string;
  label: string;
  type?: "text" | "email" | "password" | "date" | "time" | "textarea" | "number" | "select";
  placeholder?: string;
  required?: boolean;
  options?: Array<{ label: string; value: string }>;
  lookupKey?: LookupKey;
  emptyOptionLabel?: string;
  filterByField?: string;
  filterContextKey?: string;
};

export type CrudRow = TableRow & {
  id: string;
  formValues: Record<string, string>;
  canEdit?: boolean;
  canDelete?: boolean;
  lockedReason?: string;
};

export type LookupOption = {
  value: string;
  label: string;
  meta?: string;
  context?: Record<string, string | number | boolean | null>;
};
