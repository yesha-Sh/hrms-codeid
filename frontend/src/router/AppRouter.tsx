import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { LoginPage } from "../pages/LoginPage";
import { RoleProfilePage } from "../pages/ProfilePages";
import {
  AdminDepartmentDetailPage,
  AdminEmployeeDetailPage,
  AdminJobDetailPage,
} from "../pages/admin/AdminDetailPages";
import {
  AdminAttendancePage,
  AdminAuditLogsPage,
  AdminDashboardPage,
  AdminDepartmentsPage,
  AdminEmployeesPage,
  AdminHolidaysPage,
  AdminJobsPage,
  AdminLeavePage,
} from "../pages/admin/AdminPages";
import {
  EmployeeAttendancePage,
  EmployeeDashboardPage,
  EmployeeLeavePage,
} from "../pages/employee/EmployeePages";
import {
  ManagerApprovalsPage,
  ManagerAttendancePage,
  ManagerDashboardPage,
  ManagerLeavePage,
  ManagerTeamPage,
} from "../pages/manager/ManagerPages";
import { adminConfigs, employeeConfig, managerConfig } from "../data/appData";
import { ProtectedRoute, getDefaultPath } from "./ProtectedRoute";

export function AppRouter() {
  const { role } = useAuth();

  return (
    <Routes>
      <Route path="/" element={<Navigate to={role ? getDefaultPath(role) : "/login"} replace />} />
      <Route path="/login" element={<LoginPage />} />

      <Route element={<ProtectedRoute allowedRoles={["admin"]} />}>
        <Route path="/admin/dashboard" element={<AdminDashboardPage />} />
        <Route path="/admin/employees" element={<AdminEmployeesPage />} />
        <Route path="/admin/employees/:recordId" element={<AdminEmployeeDetailPage />} />
        <Route path="/admin/departments" element={<AdminDepartmentsPage />} />
        <Route path="/admin/departments/:recordId" element={<AdminDepartmentDetailPage />} />
        <Route path="/admin/attendance" element={<AdminAttendancePage />} />
        <Route path="/admin/leave" element={<AdminLeavePage />} />
        <Route path="/admin/jobs" element={<AdminJobsPage />} />
        <Route path="/admin/jobs/:recordId" element={<AdminJobDetailPage />} />
        <Route path="/admin/holidays" element={<AdminHolidaysPage />} />
        <Route path="/admin/audit-logs" element={<AdminAuditLogsPage />} />
        <Route path="/admin/profile" element={<RoleProfilePage role="admin" config={adminConfigs.dashboard} />} />
      </Route>

      <Route element={<ProtectedRoute allowedRoles={["manager"]} />}>
        <Route path="/manager/dashboard" element={<ManagerDashboardPage />} />
        <Route path="/manager/team" element={<ManagerTeamPage />} />
        <Route path="/manager/attendance" element={<ManagerAttendancePage />} />
        <Route path="/manager/leave" element={<ManagerLeavePage />} />
        <Route path="/manager/approvals" element={<ManagerApprovalsPage />} />
        <Route path="/manager/profile" element={<RoleProfilePage role="manager" config={managerConfig} />} />
      </Route>

      <Route element={<ProtectedRoute allowedRoles={["employee"]} />}>
        <Route path="/employee/dashboard" element={<EmployeeDashboardPage />} />
        <Route path="/employee/attendance" element={<EmployeeAttendancePage />} />
        <Route path="/employee/leave" element={<EmployeeLeavePage />} />
        <Route path="/employee/profile" element={<RoleProfilePage role="employee" config={employeeConfig} />} />
      </Route>

      <Route path="*" element={<Navigate to={role ? getDefaultPath(role) : "/login"} replace />} />
    </Routes>
  );
}
