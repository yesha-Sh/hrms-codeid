import type { AuthUser, Role } from "../types";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080/api/v1";

type AuthMeResponse = {
  id: string;
  email: string;
  role: Role;
  employee?: {
    id: string;
    first_name: string;
    last_name: string;
    full_name: string;
    department_name?: string;
    job_title?: string;
    work_mode?: string;
    management_scope?: string;
  };
};

type LoginResponse = {
  access_token: string;
  expires_at: string;
  user: AuthMeResponse;
};

export type ProfileResponse = AuthMeResponse & {
  organization_id?: string;
  is_active?: boolean;
  employee?: AuthMeResponse["employee"] & {
    employee_code?: string;
    phone_number?: string;
    department_id?: string;
    location_id?: string;
    location_name?: string;
    job_id?: string;
    employee_type_id?: string;
    employee_type_name?: string;
    employment_status_id?: string;
    employment_status_name?: string;
    work_mode?: string;
    management_scope?: string;
    manager_employee_id?: string | null;
    manager_name?: string;
    secondary_assignments?: Array<{
      id: string;
      job_id: string;
      job_title: string;
      job_level_name?: string;
      department_id: string;
      department_name: string;
      estimated_hours_per_week: number;
      start_date: string;
      end_date?: string;
      notes?: string;
    }>;
  };
};

export type DashboardResponse = Record<string, unknown>;

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function apiFetch<T>(path: string, init: RequestInit = {}, accessToken?: string | null): Promise<T> {
  const headers = new Headers(init.headers);
  headers.set("Content-Type", "application/json");
  if (accessToken) {
    headers.set("Authorization", `Bearer ${accessToken}`);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });

  if (!response.ok) {
    let message = "Request failed";
    try {
      const payload = await response.json() as { message?: string };
      if (payload.message) message = payload.message;
    } catch {
      // ignore json parse errors
    }
    throw new ApiError(response.status, message);
  }

  return response.json() as Promise<T>;
}

export async function login(email: string, password: string) {
  return apiFetch<LoginResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export async function refresh() {
  return apiFetch<LoginResponse>("/auth/refresh", {
    method: "POST",
  });
}

export async function me(accessToken: string) {
  return apiFetch<AuthMeResponse>("/auth/me", { method: "GET" }, accessToken);
}

export async function logout(accessToken?: string | null) {
  return apiFetch<{ message: string }>("/auth/logout", { method: "POST" }, accessToken);
}

export async function updateProfile(
  accessToken: string,
  payload: { email: string; first_name?: string; last_name?: string },
) {
  return apiFetch<ProfileResponse>("/profile", {
    method: "PUT",
    body: JSON.stringify(payload),
  }, accessToken);
}

export async function getProfile(accessToken: string) {
  return apiFetch<ProfileResponse>("/profile", {
    method: "GET",
  }, accessToken);
}

export async function getDashboard(accessToken: string, role: Role) {
  return apiFetch<DashboardResponse>(`/dashboard/${role}`, {
    method: "GET",
  }, accessToken);
}

export function toAuthUser(input: AuthMeResponse): AuthUser {
  return {
    id: input.id,
    role: input.role,
    email: input.email,
    employeeId: input.employee?.id ?? null,
    name: input.employee?.full_name ?? input.email.split("@")[0] ?? "User",
    title: input.employee?.job_title ?? roleTitle(input.role),
    department: input.employee?.department_name ?? "Unassigned",
    workMode: input.employee?.work_mode ?? "onsite",
    managementScope: input.employee?.management_scope ?? "individual_contributor",
  };
}

function roleTitle(role: Role) {
  if (role === "admin") return "Admin";
  if (role === "manager") return "Manager";
  return "Employee";
}
