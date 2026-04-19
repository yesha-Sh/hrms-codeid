import { useEffect, useMemo, useState } from "react";
import { getDashboard } from "../api/client";
import { useAuth } from "../auth/AuthContext";
import type { Role, StatCard } from "../types";

const fallbackLabels: Record<Role, StatCard[]> = {
  admin: [
    { label: "Total employees", value: "--", detail: "Live org count from PT. CODEID." },
    { label: "Today attendance", value: "--", detail: "Current attendance entries for today.", tone: "accent" },
    { label: "Pending leave requests", value: "--", detail: "Leave requests waiting for review.", tone: "warning" },
    { label: "Total departments", value: "--", detail: "Current department structure count.", tone: "cool" },
  ],
  manager: [
    { label: "Team members", value: "--", detail: "People currently in your scope." },
    { label: "Today present", value: "--", detail: "Attendance entries recorded today.", tone: "accent" },
    { label: "Pending approvals", value: "--", detail: "Items needing your review.", tone: "warning" },
    { label: "Secondary assignments", value: "--", detail: "Active flexible staffing roles.", tone: "cool" },
  ],
  employee: [
    { label: "Attendance entries", value: "--", detail: "Your recorded attendance history." },
    { label: "Pending leaves", value: "--", detail: "Leave requests still in progress.", tone: "accent" },
    { label: "Secondary assignments", value: "--", detail: "Additional active role assignments.", tone: "warning" },
    { label: "Job level", value: "--", detail: "Your current primary job level.", tone: "cool" },
  ],
};

function stringify(value: unknown) {
  if (typeof value === "number") return `${value}`;
  if (typeof value === "string") return value;
  return "--";
}

export function useDashboardStats(role: Role) {
  const { accessToken, loading: authLoading } = useAuth();
  const [data, setData] = useState<Record<string, unknown> | null>(null);

  useEffect(() => {
    if (authLoading || !accessToken) return;

    let active = true;
    void getDashboard(accessToken, role)
      .then((response) => {
        if (active) setData(response);
      })
      .catch(() => {
        if (active) setData(null);
      });

    return () => {
      active = false;
    };
  }, [accessToken, authLoading, role]);

  return useMemo<StatCard[]>(() => {
    if (!data) return fallbackLabels[role];

    if (role === "admin") {
      return [
        { label: "Total employees", value: stringify(data.total_employees), detail: "Live org count from PT. CODEID." },
        { label: "Today attendance", value: stringify(data.today_attendance), detail: "Current attendance entries for today.", tone: "accent" },
        { label: "Pending leave requests", value: stringify(data.pending_leave_requests), detail: "Leave requests waiting for review.", tone: "warning" },
        { label: "Total departments", value: stringify(data.total_departments), detail: "Current department structure count.", tone: "cool" },
      ];
    }

    if (role === "manager") {
      return [
        { label: "Team members", value: stringify(data.team_members), detail: "People currently in your scope." },
        { label: "Today present", value: stringify(data.today_present), detail: "Attendance entries recorded today.", tone: "accent" },
        { label: "Pending approvals", value: stringify(data.pending_approvals), detail: "Items needing your review.", tone: "warning" },
        { label: "Secondary assignments", value: stringify(data.active_secondary_assignments), detail: "Active flexible staffing roles.", tone: "cool" },
      ];
    }

    return [
      { label: "Attendance entries", value: stringify(data.attendance_entries), detail: "Your recorded attendance history." },
      { label: "Pending leaves", value: stringify(data.pending_leaves), detail: "Leave requests still in progress.", tone: "accent" },
      { label: "Secondary assignments", value: stringify(data.active_secondary_assignments), detail: "Additional active role assignments.", tone: "warning" },
      { label: "Job level", value: stringify(data.job_level_name), detail: "Your current primary job level.", tone: "cool" },
    ];
  }, [data, role]);
}
