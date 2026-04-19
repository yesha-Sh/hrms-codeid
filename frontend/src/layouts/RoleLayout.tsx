import { LogOut, UserRound } from "lucide-react";
import { NavLink, Link } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { navByRole, roleLabels } from "../data/appData";
import type { PageConfig, Role } from "../types";

export function RoleLayout({
  role,
  config,
  children,
}: {
  role: Role;
  config: PageConfig;
  children: React.ReactNode;
}) {
  const { logout, user } = useAuth();
  const navItems = navByRole[role].filter((item) => {
    if (role !== "manager") return true;
    const scope = user?.managementScope ?? "individual_contributor";
    if (scope === "division_manager" || scope === "subdepartment_manager") {
      return item.label !== "Team";
    }
    if (scope === "team_manager") {
      return item.label !== "Approvals";
    }
    return item.label === "Dashboard" || item.label === "Attendance" || item.label === "Leave";
  });
  const profilePath = `/${role}/profile`;
  const initials = (user?.name ?? "AR")
    .split(" ")
    .slice(0, 2)
    .map((part) => part.charAt(0))
    .join("");

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-head">
            <div className="brand-mark">CI</div>
            <div>
              <div className="eyebrow">PT. CODEID</div>
              <div className="brand-name">CODEID HRMS</div>
            </div>
          </div>
        </div>

        <nav className="sidebar-nav">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <NavLink
                key={item.label}
                to={item.path}
                className={({ isActive }) =>
                  ["nav-item", isActive ? "nav-item--active" : ""]
                    .filter(Boolean)
                    .join(" ")
                }
              >
                <span className="nav-icon">
                  <Icon size={16} />
                </span>
                <span>{item.label}</span>
              </NavLink>
            );
          })}
        </nav>

        <div className="sidebar-profile-nav">
          <div className="sidebar-profile-nav__label">Account</div>
          <NavLink
            to={profilePath}
            className={({ isActive }) =>
              ["nav-item", isActive ? "nav-item--active" : ""]
                .filter(Boolean)
                .join(" ")
            }
          >
            <span className="nav-icon">
              <UserRound size={16} />
            </span>
            <span>Profile</span>
          </NavLink>
          <button className="nav-item nav-item--button" onClick={logout} type="button">
            <span className="nav-icon">
              <LogOut size={16} />
            </span>
            <span>Logout</span>
          </button>
        </div>
      </aside>

      <main className="main-panel">
        <header className="topbar topbar--compact">
          <div className="topbar-copy topbar-copy--compact">
            <div className="topbar-meta">
              <span className="section-label">{config.sectionLabel}</span>
              <span className="live-pill">{config.badge}</span>
            </div>
            <div className="role-caption">Signed in as {roleLabels[role]}</div>
          </div>

          <div className="topbar-actions">
            <Link className="profile-chip profile-chip--link" to={profilePath}>
              <div className="profile-avatar">{initials}</div>
              <div>
            <div className="profile-name">{user?.name ?? "Ariana Reyes"}</div>
                <div className="profile-role">{user?.title ?? "HR Director"} · {user?.email ?? "hr@codeid.co.id"}</div>
              </div>
            </Link>
          </div>
        </header>

        {children}
      </main>
    </div>
  );
}
