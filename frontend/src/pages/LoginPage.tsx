import { useLocation, useNavigate } from "react-router-dom";
import { useState } from "react";
import { useAuth } from "../auth/AuthContext";
import type { Role } from "../types";

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const nextPath = (location.state as { from?: { pathname?: string } } | undefined)?.from?.pathname;
  const [form, setForm] = useState({ email: "", password: "" });
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!form.email.trim() || !form.password.trim()) {
      setError("Email and password are required.");
      return;
    }

    setError("");
    setSubmitting(true);
    try {
      const role = await login(form);
      navigate(nextPath && nextPath !== "/login" ? nextPath : getLandingPath(role), { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Sign in failed.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="login-shell">
      <div className="login-card login-card--form">
        <div className="eyebrow login-eyebrow">PT. CODEID HRMS</div>
        <h1 className="login-title">Sign in to the HRMS workspace</h1>
        <p className="login-copy">
          Sign in with your workspace email and password. Your role and access level are loaded from the backend after authentication.
        </p>

        <form className="login-form" onSubmit={handleSubmit}>
          <label className="field-block">
            <span>Email</span>
            <input
              type="email"
              value={form.email}
              onChange={(event) => setForm((current) => ({ ...current, email: event.target.value }))}
              placeholder="name@codeid.co.id"
            />
          </label>

          <label className="field-block">
            <span>Password</span>
            <input
              type="password"
              value={form.password}
              onChange={(event) => setForm((current) => ({ ...current, password: event.target.value }))}
              placeholder="Enter your password"
            />
          </label>

          {error ? <div className="form-error">{error}</div> : null}

          <button className="login-submit" disabled={submitting} type="submit">
            {submitting ? "Signing in..." : "Continue"}
          </button>
        </form>
      </div>
    </div>
  );
}

function getLandingPath(role: Role) {
  if (role === "admin") return "/admin/dashboard";
  if (role === "manager") return "/manager/dashboard";
  return "/employee/dashboard";
}
