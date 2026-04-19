import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import type { Role } from "../types";

export function ProtectedRoute({ allowedRoles }: { allowedRoles: Role[] }) {
  const { loading, role } = useAuth();
  const location = useLocation();

  if (loading) {
    return <div className="app-loading">Loading workspace...</div>;
  }

  if (!role) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  if (!allowedRoles.includes(role)) {
    return <Navigate to={getDefaultPath(role)} replace />;
  }

  return <Outlet />;
}

export function getDefaultPath(role: Role) {
  if (role === "admin") return "/admin/dashboard";
  if (role === "manager") return "/manager/dashboard";
  return "/employee/dashboard";
}
