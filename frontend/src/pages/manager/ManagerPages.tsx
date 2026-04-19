import { useEffect, useMemo, useState, type FormEvent } from "react";
import { Download, Plus } from "lucide-react";
import {
  addTeamMember,
  createAttendance,
  listAvailableTeamEmployees,
  listTeamMembers,
  listTeams,
  removeTeamMember,
  type TeamAvailableEmployeeResource,
  type TeamMemberResource,
  type TeamResource,
} from "../../api/resources";
import { useAuth } from "../../auth/AuthContext";
import { CrudSection } from "../../components/CrudSection";
import { DataTable } from "../../components/DataTable";
import { ActivityList } from "../../components/Lists";
import { MiniStatGrid } from "../../components/MiniStatGrid";
import { PageCard } from "../../components/PageCard";
import { StatGrid } from "../../components/StatGrid";
import { managerConfig } from "../../data/appData";
import { useApiCrudResource } from "../../hooks/useApiCrudResource";
import { useAdminAttendanceInsights, useAdminLeaveInsights, useManagerDashboardInsights } from "../../hooks/useLiveInsights";
import { RoleLayout } from "../../layouts/RoleLayout";
import {
  adminAttendanceConfig,
  adminLeaveConfig,
  managerLeaveConfig,
} from "../../resources/resourceConfigs";
import type { CrudRow, MiniStatItem } from "../../types";

function fallbackMiniStats(labels: string[]): MiniStatItem[] {
  return labels.map((label) => ({ label, value: "--" }));
}

function managementScopeLabel(scope?: string) {
  switch (scope) {
    case "division_manager":
      return "Division manager";
    case "subdepartment_manager":
      return "Subdepartment manager";
    case "team_manager":
      return "Team manager";
    default:
      return "Individual contributor";
  }
}

function downloadCsv(fileName: string, columns: string[], rows: CrudRow[]) {
  const header = [...columns, "Meta"];
  const lines = rows.map((row) => [row.name, ...row.cols, row.meta]);
  const csv = [header, ...lines]
    .map((line) => line.map((cell) => `"${String(cell ?? "").replaceAll("\"", '""')}"`).join(","))
    .join("\n");

  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  link.click();
  window.URL.revokeObjectURL(url);
}

export function ManagerDashboardPage() {
  const insights = useManagerDashboardInsights();

  return (
    <RoleLayout role="manager" config={managerConfig}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <PageCard title="Team overview" subtitle="Current team signal across attendance, approvals, and staffing.">
          <MiniStatGrid items={insights.overview.length ? insights.overview : fallbackMiniStats(["Present today", "Late arrivals", "Remote active", "Active departments"])} />
        </PageCard>
        <div className="side-stack">
          <PageCard title="Approvals summary" subtitle="Manager queue">
            <MiniStatGrid items={insights.approvals.length ? insights.approvals : fallbackMiniStats(["Pending", "Review", "Approved", "Coverage watch"])} />
          </PageCard>
          <PageCard title="Recent activity" subtitle="Team operations">
            <ActivityList items={insights.recentActivity.length ? insights.recentActivity : [{ title: "Loading activity", text: "from team attendance and leave records.", meta: "Please wait", tone: "teal" }]} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function ManagerTeamPage() {
  const { user, accessToken } = useAuth();
  const [teams, setTeams] = useState<TeamResource[]>([]);
  const [selectedTeamId, setSelectedTeamId] = useState("");
  const [members, setMembers] = useState<TeamMemberResource[]>([]);
  const [availableEmployees, setAvailableEmployees] = useState<TeamAvailableEmployeeResource[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");

  const selectedTeam = useMemo(
    () => teams.find((team) => team.id === selectedTeamId) ?? teams[0] ?? null,
    [selectedTeamId, teams],
  );

  const loadTeams = async (activeSearch = search) => {
    if (!accessToken) return;
    setLoading(true);
    setError(null);
    try {
      const teamResponse = await listTeams(accessToken);
      const nextTeams = teamResponse.items;
      setTeams(nextTeams);
      const nextSelectedTeamId = selectedTeamId && nextTeams.some((team) => team.id === selectedTeamId)
        ? selectedTeamId
        : nextTeams[0]?.id ?? "";
      setSelectedTeamId(nextSelectedTeamId);

      if (nextSelectedTeamId) {
        const [membersResponse, availableResponse] = await Promise.all([
          listTeamMembers(accessToken, nextSelectedTeamId),
          listAvailableTeamEmployees(accessToken, nextSelectedTeamId, activeSearch.trim() || undefined),
        ]);
        setMembers(membersResponse.items);
        setAvailableEmployees(availableResponse.items);
      } else {
        setMembers([]);
        setAvailableEmployees([]);
      }
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Could not load managed teams");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!accessToken) return;
    void loadTeams();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [accessToken, selectedTeamId]);

  useEffect(() => {
    if (!accessToken || !selectedTeamId) return;
    const timer = window.setTimeout(() => {
      void loadTeams(search);
    }, 220);
    return () => window.clearTimeout(timer);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search]);

  const memberRows: CrudRow[] = members.map((member) => ({
    id: member.id,
    initials: member.employee_name.split(" ").slice(0, 2).map((part) => part.charAt(0)).join(""),
    name: member.employee_name,
    meta: `${member.employee_code} · ${member.role_name ?? "Member"}`,
    tone: member.work_mode === "remote" ? "cool" : "default",
    cols: [
      member.department_name ?? "--",
      member.job_title ?? "--",
      member.job_level_name ?? "--",
      member.work_mode ?? "--",
      member.location_name ?? "--",
      member.start_date ?? "--",
    ],
    formValues: {},
  }));

  const availableRows: CrudRow[] = availableEmployees.map((employee) => ({
    id: employee.id,
    initials: employee.full_name.split(" ").slice(0, 2).map((part) => part.charAt(0)).join(""),
    name: employee.full_name,
    meta: `${employee.employee_code} · ${employee.employment_status ?? "Active"}`,
    tone: employee.work_mode === "remote" ? "cool" : "default",
    cols: [
      employee.department_name ?? "--",
      employee.job_title ?? "--",
      employee.job_level_name ?? "--",
      employee.work_mode ?? "--",
      employee.location_name ?? "--",
      "Ready",
    ],
    formValues: {},
  }));

  const exportTeamRows = async () => {
    downloadCsv(
      `${(selectedTeam?.name ?? "managed-team").replaceAll(" ", "-").toLowerCase()}.csv`,
      ["Employee", "Department", "Primary job", "Level", "Work mode", "Location", "Assigned"],
      memberRows,
    );
  };

  const handleAssign = async (row: CrudRow) => {
    if (!accessToken || !selectedTeam) return;
    setSaving(true);
    setError(null);
    try {
      await addTeamMember(accessToken, selectedTeam.id, { employee_id: row.id });
      await loadTeams(search);
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Could not add team member");
    } finally {
      setSaving(false);
    }
  };

  const handleRemove = async (row: CrudRow) => {
    if (!accessToken || !selectedTeam) return;
    setSaving(true);
    setError(null);
    try {
      await removeTeamMember(accessToken, selectedTeam.id, row.id);
      await loadTeams(search);
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Could not remove team member");
    } finally {
      setSaving(false);
    }
  };

  return (
    <RoleLayout role="manager" config={managerConfig}>
      {user?.managementScope !== "team_manager" ? (
        <PageCard title="Operational team management" subtitle="This manager account governs a department or subdepartment, not a daily delivery team.">
          <div className="notice-bar notice-bar--soft">
            <span>Division and subdepartment managers in v1 focus on department oversight and approvals. Dedicated team managers handle cross-department team operations.</span>
            <strong>{managementScopeLabel(user?.managementScope)}</strong>
          </div>
        </PageCard>
      ) : null}
      <div className="main-stack">
        <PageCard title="Managed teams" subtitle="Cross-department team operations stay separate from employee master records.">
          <div className="section-toolbar">
            <div className="filter-row filter-row--compact">
              <span className="chip">Manager type: {managementScopeLabel(user?.managementScope)}</span>
              <span className="chip">Operational teams only</span>
              {selectedTeam?.is_cross_department ? <span className="chip">Cross-department</span> : null}
            </div>
            <div className="section-toolbar__actions">
              <label className="table-search">
                <span>Search</span>
                <input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search available employee, department, or role" />
              </label>
              <button className="ghost-button" type="button" onClick={exportTeamRows} disabled={loading || memberRows.length === 0}> 
                <Download size={16} />
                Export CSV
              </button>
            </div>
          </div>
          {teams.length > 0 ? (
            <div className="filter-row filter-row--compact">
              {teams.map((team) => (
                <button
                  key={team.id}
                  className={["chip", selectedTeam?.id === team.id ? "chip--active" : ""].join(" ").trim()}
                  onClick={() => setSelectedTeamId(team.id)}
                  type="button"
                >
                  {team.name}
                </button>
              ))}
            </div>
          ) : null}
          {selectedTeam?.focus_area ? (
            <div className="notice-bar notice-bar--soft">
              <span>{selectedTeam.focus_area}</span>
              <strong>{selectedTeam.department_name ?? "Cross-team"}</strong>
            </div>
          ) : null}
          {error ? <div className="notice-bar notice-bar--soft"><span>{error}</span><strong>Team sync</strong></div> : null}
          <DataTable
            columns={["Employee", "Department", "Primary job", "Level", "Work mode", "Location", "Assigned"]}
            rows={memberRows}
            variant="employees"
            emptyMessage={loading ? "Loading team members..." : "No active members are assigned to the selected team yet."}
            actions={(row) => (
              <button className="table-action table-action--danger" onClick={() => void handleRemove(row as CrudRow)} type="button" disabled={saving}>
                Remove
              </button>
            )}
            footer={
              <>
                <span>Showing {memberRows.length} active members in {selectedTeam?.name ?? "your managed team"}</span>
              </>
            }
          />
        </PageCard>

        <PageCard title="Available employees without a team" subtitle="Only active employees without another operational team can be added here.">
          <DataTable
            columns={["Employee", "Department", "Primary job", "Level", "Work mode", "Location", "Availability"]}
            rows={availableRows}
            variant="employees"
            emptyMessage={loading ? "Loading available employees..." : "No unassigned employees are available on this page."}
            actions={(row) => (
              <button className="table-action" onClick={() => void handleAssign(row as CrudRow)} type="button" disabled={saving || !selectedTeam}>
                <Plus size={14} />
                Add to team
              </button>
            )}
          />
        </PageCard>
      </div>
    </RoleLayout>
  );
}

export function ManagerAttendancePage() {
  const { accessToken, user } = useAuth();
  const insights = useAdminAttendanceInsights();
  const { rows, loading, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows, refresh } = useApiCrudResource(adminAttendanceConfig);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [values, setValues] = useState({
    attendance_date: "",
    check_in_time: "",
    check_out_time: "",
    status: "on time",
    notes: "",
  });
  const attendanceStatusLocked = user?.workMode === "remote" || user?.workMode === "client-based";

  useEffect(() => {
    if (attendanceStatusLocked) {
      setValues((current) => ({ ...current, status: "remote" }));
    }
  }, [attendanceStatusLocked]);

  const exportAttendance = async () => {
    const data = await exportRows();
    downloadCsv("manager-attendance.csv", ["Employee", "Date", "Check in", "Check out", "Status", "Location"], data);
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!accessToken || !user?.employeeId) return;
    setSaving(true);
    setSubmitError(null);
    try {
      const checkIn = values.attendance_date && values.check_in_time ? new Date(`${values.attendance_date}T${values.check_in_time}:00`).toISOString() : null;
      const checkOut = values.attendance_date && values.check_out_time ? new Date(`${values.attendance_date}T${values.check_out_time}:00`).toISOString() : null;
      await createAttendance(accessToken, {
        employee_id: user.employeeId,
        attendance_date: values.attendance_date,
        check_in_at: checkIn,
        check_out_at: checkOut,
        status: values.status,
        notes: values.notes,
      });
      setValues({ attendance_date: "", check_in_time: "", check_out_time: "", status: "on time", notes: "" });
      await refresh();
    } catch (nextError) {
      setSubmitError(nextError instanceof Error ? nextError.message : "Could not add attendance");
    } finally {
      setSaving(false);
    }
  };

  return (
    <RoleLayout role="manager" config={managerConfig}>
      {user?.managementScope !== "team_manager" ? (
        <PageCard title="Attendance operations" subtitle="In v1, day-to-day attendance monitoring belongs to designated team managers.">
          <div className="notice-bar notice-bar--soft">
            <span>This account is configured as {managementScopeLabel(user?.managementScope).toLowerCase()}. Division and subdepartment managers stay focused on department oversight instead of operational attendance handling.</span>
            <strong>Read-only policy</strong>
          </div>
        </PageCard>
      ) : null}
      <div className="content-grid">
        <PageCard title="Team attendance" subtitle="Review attendance history for your team. Managers can export records and submit their own attendance only.">
          <div className="section-toolbar">
            <div className="filter-row filter-row--compact">
              <span className="chip">Management scope: {managementScopeLabel(user?.managementScope)}</span>
              <span className="chip">Scope: {user?.managementScope === "team_manager" ? "Team + self" : "Self only"}</span>
              <span className="chip">Edit rights: Self only</span>
            </div>
            <div className="section-toolbar__actions">
              <label className="table-search">
                <span>Search</span>
                <input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search employee, date, or status" />
              </label>
              <button className="ghost-button" type="button" onClick={exportAttendance} disabled={exporting || loading}>
                <Download size={16} />
                {exporting ? "Exporting..." : "Export CSV"}
              </button>
            </div>
          </div>
          {error ? <div className="notice-bar notice-bar--soft"><span>{error}</span><strong>Attendance sync</strong></div> : null}
          <DataTable
            columns={["Employee", "Date", "Check in", "Check out", "Status", "Location"]}
            rows={rows}
            variant="attendance"
            emptyMessage={loading ? "Loading attendance records..." : "No attendance records are available yet."}
            footer={
              <>
                <span>Showing {total} attendance records in your current scope</span>
                <div className="pagination">
                  <button className="pagination__item" onClick={() => setPage(1)} type="button" disabled={page <= 1}>«</button>
                  <button className="pagination__item" onClick={() => setPage(page - 1)} type="button" disabled={page <= 1}>‹</button>
                  <span className="pagination__item pagination__item--active">{page}</span>
                  <span className="pagination__text">of {pageCount}</span>
                  <button className="pagination__item" onClick={() => setPage(page + 1)} type="button" disabled={page >= pageCount}>›</button>
                  <button className="pagination__item" onClick={() => setPage(pageCount)} type="button" disabled={page >= pageCount}>»</button>
                </div>
              </>
            }
          />
        </PageCard>
        <div className="side-stack">
          <PageCard title="My attendance entry" subtitle="Submit your own attendance record.">
            <form className="crud-form crud-form--single" onSubmit={submit}>
              <label className="field-block">
                <span>Attendance date</span>
                <input type="date" value={values.attendance_date} onChange={(event) => setValues((current) => ({ ...current, attendance_date: event.target.value }))} required />
              </label>
              <label className="field-block">
                <span>Check in</span>
                <input type="time" value={values.check_in_time} onChange={(event) => setValues((current) => ({ ...current, check_in_time: event.target.value }))} />
              </label>
              <label className="field-block">
                <span>Check out</span>
                <input type="time" value={values.check_out_time} onChange={(event) => setValues((current) => ({ ...current, check_out_time: event.target.value }))} />
              </label>
              <label className="field-block">
                <span>Status</span>
                <select value={values.status} onChange={(event) => setValues((current) => ({ ...current, status: event.target.value }))} disabled={attendanceStatusLocked}>
                  <option value="on time">On time</option>
                  <option value="late">Late</option>
                  <option value="remote">Remote</option>
                  <option value="present">Present</option>
                </select>
              </label>
              {attendanceStatusLocked ? <div className="notice-bar notice-bar--soft"><span>Your employee record is marked as remote/client-based, so attendance is submitted as remote automatically.</span><strong>Auto-applied</strong></div> : null}
              <label className="field-block field-block--full">
                <span>Notes</span>
                <textarea rows={4} value={values.notes} onChange={(event) => setValues((current) => ({ ...current, notes: event.target.value }))} placeholder="Optional attendance context" />
              </label>
              {submitError ? <div className="form-error form-error--inline">{submitError}</div> : null}
              <div className="crud-form__actions">
                <button className="primary-button primary-button--inline" type="submit" disabled={saving}>{saving ? "Saving..." : "Add my attendance"}</button>
              </div>
            </form>
          </PageCard>
          <PageCard title="Attendance summary" subtitle="Current team conditions">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Late", "Remote", "Missing checkout", "Locations active"])} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function ManagerLeavePage() {
  const { user } = useAuth();
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(managerLeaveConfig);

  return (
    <RoleLayout role="manager" config={managerConfig}>
      <CrudSection
        title="My leave"
        subtitle="Create, adjust, and track your own leave requests. Approver routing follows the PT. CODEID management chain automatically."
        recordLabel={config.recordLabel}
        actions={<div className="filter-pills"><span className="pill pill--teal">Self service</span><span className="pill pill--soft">{managementScopeLabel(user?.managementScope)}</span></div>}
        filters={["Scope: My requests", "Status: Pending + history", "Approver: Auto-assigned"]}
        primaryLabel="Request leave"
        columns={["Request", "Dates", "Leave type", "Status", "Approver", "Updated"]}
        variant="employees"
        rows={rows}
        fields={config.fields}
        createDefaults={config.createDefaults}
        onSave={saveRecord}
        onDelete={(row: CrudRow) => removeRecord(row.id)}
        footerText={(count: number) => `Showing ${count} leave requests`}
        lookupOptions={lookupOptions}
        loading={loading}
        saving={saving}
        exporting={exporting}
        error={error}
        emptyMessage="No leave requests are available yet."
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search leave type or status"
        page={page}
        pageCount={pageCount}
        totalCount={total}
        onPageChange={setPage}
        onExport={exportRows}
        exportFileName="manager-my-leave.csv"
      />
    </RoleLayout>
  );
}

export function ManagerApprovalsPage() {
  const { user } = useAuth();
  const managerEmployeeId = user?.employeeId ?? "";
  const insights = useAdminLeaveInsights();
  const { rows, loading, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows, saveRecord, saving } = useApiCrudResource(adminLeaveConfig);
  const [actionError, setActionError] = useState<string | null>(null);

  const exportApprovals = async () => {
    const data = await exportRows();
    downloadCsv("manager-approvals.csv", ["Employee", "Dates", "Leave type", "Status", "Approver", "Updated"], data);
  };

  const updateStatus = async (row: CrudRow, status: "approved" | "rejected") => {
    setActionError(null);
    try {
      await saveRecord({ ...row.formValues, approver_employee_id: "", status }, row.id);
    } catch (nextError) {
      setActionError(nextError instanceof Error ? nextError.message : `Could not mark request as ${status}`);
    }
  };

  const reviewableRows = useMemo(() => rows, [rows]);

  return (
    <RoleLayout role="manager" config={managerConfig}>
      {user?.managementScope === "team_manager" ? (
        <PageCard title="Leave approvals" subtitle="In v1, leave approvals belong to division or subdepartment managers and admins.">
          <div className="notice-bar notice-bar--soft">
            <span>This account is configured as a team manager, so leave approvals are handled by the relevant department leadership instead of the operational team lead.</span>
            <strong>Scope limited</strong>
          </div>
        </PageCard>
      ) : null}
      <div className="content-grid">
        <PageCard title="Approval center" subtitle="Approve or reject leave requests in your department scope. Approver assignment is automatic.">
          <div className="section-toolbar">
            <div className="filter-row filter-row--compact">
              <span className="chip">Management scope: {managementScopeLabel(user?.managementScope)}</span>
              <span className="chip">Queue: Team leave</span>
              <span className="chip">Actions: Approve or reject</span>
            </div>
            <div className="section-toolbar__actions">
              <label className="table-search">
                <span>Search</span>
                <input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search employee, leave type, or status" />
              </label>
              <button className="ghost-button" type="button" onClick={exportApprovals} disabled={exporting || loading}>
                <Download size={16} />
                {exporting ? "Exporting..." : "Export CSV"}
              </button>
            </div>
          </div>
          {error || actionError ? <div className="notice-bar notice-bar--soft"><span>{actionError ?? error}</span><strong>Approval sync</strong></div> : null}
          <DataTable
            columns={["Employee", "Dates", "Leave type", "Status", "Approver", "Updated"]}
            rows={reviewableRows}
            variant="employees"
            emptyMessage={loading ? "Loading approval requests..." : "No approval requests are available yet."}
            actions={(row) => {
              const crudRow = row as CrudRow;
              const currentStatus = (crudRow.formValues.status ?? "").toLowerCase();
              const requestOwnerId = crudRow.formValues.employee_id ?? "";

              if (requestOwnerId === managerEmployeeId) {
                return <span className="table-muted">Self request</span>;
              }
              if (!["pending", "review"].includes(currentStatus)) {
                return <span className="table-muted">{currentStatus || "Closed"}</span>;
              }

              return (
                <>
                  <button className="table-action" onClick={() => void updateStatus(crudRow, "approved")} type="button" disabled={saving}>Approve</button>
                  <button className="table-action table-action--danger" onClick={() => void updateStatus(crudRow, "rejected")} type="button" disabled={saving}>Reject</button>
                </>
              );
            }}
            footer={
              <>
                <span>Showing {total} approval requests in your current scope</span>
                <div className="pagination">
                  <button className="pagination__item" onClick={() => setPage(1)} type="button" disabled={page <= 1}>«</button>
                  <button className="pagination__item" onClick={() => setPage(page - 1)} type="button" disabled={page <= 1}>‹</button>
                  <span className="pagination__item pagination__item--active">{page}</span>
                  <span className="pagination__text">of {pageCount}</span>
                  <button className="pagination__item" onClick={() => setPage(page + 1)} type="button" disabled={page >= pageCount}>›</button>
                  <button className="pagination__item" onClick={() => setPage(pageCount)} type="button" disabled={page >= pageCount}>»</button>
                </div>
              </>
            }
          />
        </PageCard>
        <div className="side-stack">
          <PageCard title="Approval summary" subtitle="Current team leave mix">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Pending", "Review", "Approved", "Rejected"])} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}
