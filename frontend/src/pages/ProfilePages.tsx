import { useEffect, useMemo, useState } from "react";
import { listLookup, type LookupItem } from "../api/resources";
import { PageCard } from "../components/PageCard";
import { RoleLayout } from "../layouts/RoleLayout";
import { useAuth } from "../auth/AuthContext";
import { getProfile, type ProfileResponse } from "../api/client";
import type { PageConfig, Role } from "../types";

function toOptions(items: LookupItem[]) {
  return items.map((item) => ({ value: item.id, label: item.label }));
}

export function RoleProfilePage({ role, config }: { role: Role; config: PageConfig }) {
  const { user, updateUser, accessToken, logout } = useAuth();
  const [open, setOpen] = useState(false);
  const [profile, setProfile] = useState<ProfileResponse | null>(null);
  const [departmentOptions, setDepartmentOptions] = useState<Array<{ value: string; label: string }>>([]);
  const [jobOptions, setJobOptions] = useState<Array<{ value: string; label: string }>>([]);
  const [values, setValues] = useState({
    name: user?.name ?? "",
    email: user?.email ?? "",
    department_id: "",
    job_id: "",
  });
  const [saving, setSaving] = useState(false);
  const [exporting, setExporting] = useState(false);
  const [loadingProfile, setLoadingProfile] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!accessToken) {
      setLoadingProfile(false);
      return;
    }

    let active = true;
    setLoadingProfile(true);
    setError("");

    void Promise.all([
      getProfile(accessToken),
      listLookup(accessToken, "departments"),
      listLookup(accessToken, "jobs"),
    ])
      .then(([profileResponse, departmentsResponse, jobsResponse]) => {
        if (!active) return;
        setProfile(profileResponse);
        setDepartmentOptions(toOptions(departmentsResponse.items));
        setJobOptions(toOptions(jobsResponse.items));
        setValues({
          name: profileResponse.employee?.full_name ?? user?.name ?? "",
          email: profileResponse.email ?? user?.email ?? "",
          department_id: profileResponse.employee?.department_id ?? "",
          job_id: profileResponse.employee?.job_id ?? "",
        });
      })
      .catch((nextError) => {
        if (!active) return;
        setError(nextError instanceof Error ? nextError.message : "Could not load profile data.");
      })
      .finally(() => {
        if (active) setLoadingProfile(false);
      });

    return () => {
      active = false;
    };
  }, [accessToken, user?.email, user?.name]);

  const employee = profile?.employee;
  const displayName = employee?.full_name ?? user?.name ?? "Ariana Reyes";
  const displayTitle = employee?.job_title ?? user?.title ?? "HR Director";
  const displayDepartment = employee?.department_name ?? user?.department ?? "People & Culture";
  const displayWorkMode = employee?.work_mode ?? user?.workMode ?? "onsite";
  const displayManagementScope = employee?.management_scope ?? user?.managementScope ?? "individual_contributor";

  const statusLabel = useMemo(() => {
    if (profile?.is_active === false) return "Inactive";
    return employee?.employment_status_name ?? "Active";
  }, [employee?.employment_status_name, profile?.is_active]);

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaving(true);
    setError("");
    try {
      await updateUser({ name: values.name, email: values.email });
      if (profile) {
        setProfile({
          ...profile,
          email: values.email,
          employee: profile.employee ? { ...profile.employee, full_name: values.name } : profile.employee,
        });
      }
      setOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not update profile.");
    } finally {
      setSaving(false);
    }
  };

  const exportProfile = async () => {
    if (!accessToken) return;
    setExporting(true);
    setError("");
    try {
      const profileData = await getProfile(accessToken);
      const blob = new Blob([JSON.stringify(profileData, null, 2)], { type: "application/json;charset=utf-8;" });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${role}-profile-export.json`;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not export profile data.");
    } finally {
      setExporting(false);
    }
  };

  return (
    <RoleLayout role={role} config={config}>
      <div className="content-grid content-grid--profile">
        <PageCard title="Profile" subtitle="Identity, employment, and access context.">
          <div className="profile-hero">
            <div className="profile-hero__avatar">{displayName.split(" ").slice(0, 2).map((part) => part.charAt(0)).join("")}</div>
            <div>
              <div className="profile-hero__name">{displayName}</div>
              <div className="muted-line">{displayTitle}</div>
              <div className="muted-line">{profile?.email ?? user?.email ?? "hr@codeid.co.id"}</div>
            </div>
          </div>
          <div className="mini-grid mini-grid--two">
            <div className="mini-stat"><span>Role</span><strong>{role}</strong></div>
            <div className="mini-stat"><span>Department</span><strong>{displayDepartment}</strong></div>
            <div className="mini-stat"><span>Workspace</span><strong>PT. CODEID</strong></div>
            <div className="mini-stat"><span>Status</span><strong>{statusLabel}</strong></div>
            <div className="mini-stat"><span>Work mode</span><strong>{displayWorkMode}</strong></div>
            <div className="mini-stat"><span>Management scope</span><strong>{displayManagementScope}</strong></div>
          </div>
        </PageCard>

        <PageCard title="Profile actions" subtitle="Common account operations.">
          <div className="quick-actions quick-actions--stacked">
            <button className="primary-button" onClick={() => {
              setValues({
                name: employee?.full_name ?? user?.name ?? "",
                email: profile?.email ?? user?.email ?? "",
                department_id: employee?.department_id ?? "",
                job_id: employee?.job_id ?? "",
              });
              setOpen(true);
            }} type="button">Edit profile</button>
            <button className="ghost-button" type="button" onClick={exportProfile} disabled={exporting}>{exporting ? "Exporting..." : "Export my data"}</button>
            <button className="ghost-button ghost-button--danger" type="button" onClick={() => void logout()}>Logout</button>
          </div>
          {error ? <div className="form-error">{error}</div> : null}
          {loadingProfile ? <div className="notice-bar notice-bar--soft"><span>Loading profile details from the database.</span><strong>Please wait</strong></div> : null}
        </PageCard>
      </div>

      {open ? (
        <div className="modal-backdrop" onClick={() => setOpen(false)} role="presentation">
          <div className="modal-card" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
            <div className="modal-card__head">
              <div>
                <div className="card-title">Edit profile</div>
                <div className="modal-card__subtitle">Update your account details for this workspace. Department and job title are shown from the live database.</div>
              </div>
              <button className="ghost-button" onClick={() => setOpen(false)} type="button">Close</button>
            </div>
            <form className="crud-form" onSubmit={submit}>
              <label className="field-block">
                <span>Full name</span>
                <input type="text" value={values.name} onChange={(event) => setValues((current) => ({ ...current, name: event.target.value }))} required />
              </label>
              <label className="field-block">
                <span>Email</span>
                <input type="email" value={values.email} onChange={(event) => setValues((current) => ({ ...current, email: event.target.value }))} required />
              </label>
              <label className="field-block">
                <span>Department</span>
                <select value={values.department_id} disabled>
                  <option value="">No department assigned</option>
                  {departmentOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                </select>
              </label>
              <label className="field-block">
                <span>Job title</span>
                <select value={values.job_id} disabled>
                  <option value="">No job assigned</option>
                  {jobOptions.map((option) => <option key={option.value} value={option.value}>{option.label}</option>)}
                </select>
              </label>
              {error ? <div className="form-error form-error--inline">{error}</div> : null}
              <div className="crud-form__actions">
                <button className="ghost-button" onClick={() => setOpen(false)} type="button">Cancel</button>
                <button className="primary-button primary-button--inline" disabled={saving} type="submit">{saving ? "Saving..." : "Save profile"}</button>
              </div>
            </form>
          </div>
        </div>
      ) : null}
    </RoleLayout>
  );
}
