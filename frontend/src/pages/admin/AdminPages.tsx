import { useEffect, useMemo, useState } from "react";
import { CrudSection } from "../../components/CrudSection";
import { DataTable } from "../../components/DataTable";
import { ActivityList, ActionList } from "../../components/Lists";
import { MiniStatGrid } from "../../components/MiniStatGrid";
import { PageCard } from "../../components/PageCard";
import { StatGrid } from "../../components/StatGrid";
import { listAuditLogs, type AuditLogResource } from "../../api/resources";
import { useAuth } from "../../auth/AuthContext";
import { adminConfigs } from "../../data/appData";
import {
  useAdminAttendanceInsights,
  useAdminAuditInsights,
  useAdminDashboardInsights,
  useAdminDepartmentsInsights,
  useAdminEmployeesInsights,
  useAdminHolidayInsights,
  useAdminJobsInsights,
  useAdminLeaveInsights,
} from "../../hooks/useLiveInsights";
import { useApiCrudResource } from "../../hooks/useApiCrudResource";
import { RoleLayout } from "../../layouts/RoleLayout";
import {
  adminAttendanceConfig,
  adminLeaveConfig,
  departmentConfig,
  employeeConfig as employeeResourceConfig,
  holidayConfig,
  jobConfig,
} from "../../resources/resourceConfigs";
import type { MiniStatItem, TableRow } from "../../types";

function fallbackMiniStats(labels: string[]): MiniStatItem[] {
  return labels.map((label) => ({ label, value: "--" }));
}

function AttendancePatternCard({ metrics, bars }: { metrics: MiniStatItem[]; bars: Array<{ day: string; outerHeight: number; innerHeight: number; tone: "teal" | "deep" | "soft" | "gold" }> }) {
  return (
    <PageCard title="Attendance pattern" subtitle="Live attendance shape across recent PT. CODEID activity." actions={<span className="pill pill--soft">Live</span>}>
      <div className="attendance-overview">
        <div className="bar-cluster">
          {bars.map((bar, index) => (
            <div className="bar-stack" key={`${bar.day}-${index}`}>
              <div className={`bar-stack__outer bar-stack__outer--${bar.tone}`} style={{ height: `${bar.outerHeight}px` }}>
                <div className="bar-stack__inner" style={{ height: `${bar.innerHeight}px` }} />
              </div>
              <span>{bar.day}</span>
            </div>
          ))}
        </div>
        <div className="metric-strip">
          {metrics.map((metric, index) => (
            <div className="inline-metric" key={metric.id ?? `${metric.label}-${index}`}>
              <span>{metric.label}</span><strong>{metric.value}</strong>
            </div>
          ))}
        </div>
      </div>
    </PageCard>
  );
}

function downloadAuditCsv(rows: TableRow[]) {
  const header = ["Actor", "Entity", "Action", "Record", "Time", "Severity", "Meta"];
  const csv = [header, ...rows.map((row) => [row.name, ...row.cols, row.meta])]
    .map((line) => line.map((cell) => `\"${String(cell ?? "").replaceAll("\"", "\"\"")}\"`).join(","))
    .join("\n");

  const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "audit-logs.csv";
  link.click();
  window.URL.revokeObjectURL(url);
}

export function AdminDashboardPage() {
  const insights = useAdminDashboardInsights();
  const metrics = insights.attendanceMetrics.length ? insights.attendanceMetrics : fallbackMiniStats(["On-time entries", "Late departments", "Remote check-ins", "Pending approvals"]);
  const bars = insights.attendanceBars.length
    ? insights.attendanceBars
    : [
        { day: "Mon", outerHeight: 48, innerHeight: 18, tone: "soft" as const },
        { day: "Tue", outerHeight: 48, innerHeight: 18, tone: "soft" as const },
        { day: "Wed", outerHeight: 48, innerHeight: 18, tone: "soft" as const },
        { day: "Thu", outerHeight: 48, innerHeight: 18, tone: "soft" as const },
        { day: "Fri", outerHeight: 48, innerHeight: 18, tone: "soft" as const },
      ];

  return (
    <RoleLayout role="admin" config={adminConfigs.dashboard}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <AttendancePatternCard metrics={metrics} bars={bars} />
        <div className="side-stack">
          <PageCard title="Recent activity" subtitle="Live audit-aware stream">
            {insights.error ? <div className="notice-bar notice-bar--soft"><span>{insights.error}</span><strong>Retry later</strong></div> : null}
            <ActivityList items={insights.recentActivity.length ? insights.recentActivity : [{ title: "Loading activity", text: "from the backend.", meta: "Please wait", tone: "teal" }]} />
          </PageCard>
          <PageCard title="Priority approvals" subtitle="Current review queue">
            <ActionList items={insights.priorityApprovals.length ? insights.priorityApprovals : [{ initials: "--", name: "Loading approvals", detail: "Pulling the current queue.", status: "Loading", tone: "soft" }]} notice={{ text: "Approvals stay in sync with the live leave queue.", action: "Backend linked" }} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminEmployeesPage() {
  const insights = useAdminEmployeesInsights();
  const [managementScope, setManagementScope] = useState("all");
  const scopedEmployeeConfig = useMemo(() => ({
    ...employeeResourceConfig,
    list: (accessToken: string, query?: Parameters<typeof employeeResourceConfig.list>[1]) =>
      employeeResourceConfig.list(accessToken, {
        ...query,
        filters: {
          ...(query?.filters ?? {}),
          management_scope: managementScope === "all" ? undefined : managementScope,
        },
      }),
  }), [managementScope]);
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(scopedEmployeeConfig);

  return (
    <RoleLayout role="admin" config={adminConfigs.employees}>
      <StatGrid stats={insights.stats} />
      <div className="main-stack">
        <CrudSection
          title="Employee directory"
          subtitle="Manage PT. CODEID employee records, primary placement, and staffing status from one control surface."
          recordLabel={config.recordLabel}
          actions={
            <div className="section-toolbar__actions">
              <div className="filter-pills">
                <span className="pill pill--teal">All employees</span>
                <span className="pill pill--soft">Needs review</span>
              </div>
              <label className="toolbar-select">
                <span>Management scope</span>
                <select value={managementScope} onChange={(event) => {
                  setManagementScope(event.target.value);
                  setPage(1);
                }}>
                  <option value="all">All scopes</option>
                  <option value="individual_contributor">Individual contributor</option>
                  <option value="division_manager">Division manager</option>
                  <option value="subdepartment_manager">Subdepartment manager</option>
                  <option value="team_manager">Team manager</option>
                </select>
              </label>
            </div>
          }
          filters={[`Management scope: ${managementScope === "all" ? "All" : managementScope.replaceAll("_", " ")}`, "Status: Active + probation", "Location: All"]}
          primaryLabel="Add employee"
          columns={["Employee", "Department", "Primary job", "Level", "Management", "Status", "Location", "Updated"]}
          variant="employees"
          rows={rows}
          fields={config.fields}
          createDefaults={config.createDefaults}
          onSave={saveRecord}
          onDelete={(row) => removeRecord(row.id)}
          getViewPath={(row) => `/admin/employees/${row.id}`}
          footerText={(count) => `Showing ${count} employee records`}
          lookupOptions={lookupOptions}
          loading={loading}
          saving={saving}
          exporting={exporting}
          error={error}
          emptyMessage="No employee records are available yet."
          searchValue={search}
          onSearchChange={setSearch}
          searchPlaceholder="Search employee, email, or employee code"
          page={page}
          pageCount={pageCount}
          totalCount={total}
          onPageChange={setPage}
          onExport={exportRows}
          exportFileName="employees.csv"
        />
        <div className="content-grid content-grid--profile">
          <PageCard title="Staffing spotlight" subtitle="Live employee coverage and reporting readiness.">
            {insights.error ? <div className="notice-bar notice-bar--soft"><span>{insights.error}</span><strong>Using live retry</strong></div> : null}
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Primary roles", "Secondary roles", "Probation", "Manager links"])} />
          </PageCard>
          <PageCard title="Directory guidance" subtitle="Keep core records clean while workload planning stays separate.">
            <div className="notice-bar notice-bar--soft">
              <span>Primary employee records should stay focused on the main placement. Use Secondary Assignments from the employee detail page for temporary cross-functional work.</span>
              <strong>Table-first layout</strong>
            </div>
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminDepartmentsPage() {
  const insights = useAdminDepartmentsInsights();
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(departmentConfig);

  return (
    <RoleLayout role="admin" config={adminConfigs.departments}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <CrudSection
          title="Department directory"
          subtitle="Shape the PT. CODEID hierarchy, assign managers, and maintain reporting clarity."
          recordLabel={config.recordLabel}
          actions={<div className="filter-pills"><span className="pill pill--teal">All departments</span><span className="pill pill--soft">Hierarchy view</span></div>}
          filters={["Location: All", "Manager: Assigned", "Level: 0-2"]}
          primaryLabel="Add department"
          columns={["Department", "Manager", "Location", "Level", "Parent", "Updated"]}
          variant="departments"
          rows={rows}
          fields={config.fields}
          createDefaults={config.createDefaults}
          onSave={saveRecord}
          onDelete={(row) => removeRecord(row.id)}
          getViewPath={(row) => `/admin/departments/${row.id}`}
          footerText={(count) => `Showing ${count} departments`}
          lookupOptions={lookupOptions}
          loading={loading}
          saving={saving}
          exporting={exporting}
          error={error}
          emptyMessage="No departments are available yet."
          searchValue={search}
          onSearchChange={setSearch}
          searchPlaceholder="Search department, manager, or location"
          page={page}
          pageCount={pageCount}
          totalCount={total}
          onPageChange={setPage}
          onExport={exportRows}
          exportFileName="departments.csv"
        />
        <div className="side-stack">
          <PageCard title="Hierarchy focus" subtitle="Live organization structure">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Root node", "Level 1", "Level 2", "Managers mapped"])} />
          </PageCard>
          <PageCard title="Department guidance" subtitle="Team membership follows the department tree.">
            <div className="notice-bar notice-bar--soft">
              <span>Managers should add team members from their accessible department structure so reporting stays consistent.</span>
              <strong>Scope aligned</strong>
            </div>
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminAttendancePage() {
  const insights = useAdminAttendanceInsights();
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(adminAttendanceConfig);

  return (
    <RoleLayout role="admin" config={adminConfigs.attendance}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <div className="main-stack">
          <CrudSection
            title="Live attendance log"
            subtitle="Current attendance activity across PT. CODEID divisions and locations."
            recordLabel={config.recordLabel}
            actions={<div className="filter-pills"><span className="pill pill--teal">All records</span><span className="pill pill--soft">Exceptions first</span></div>}
            filters={["Status: All", "Location: All", "Window: Today"]}
            primaryLabel="Add record"
            columns={["Employee", "Date", "Check in", "Check out", "Status", "Location"]}
            variant="attendance"
            rows={rows}
            fields={config.fields}
            createDefaults={config.createDefaults}
            onSave={saveRecord}
            onDelete={(row) => removeRecord(row.id)}
            footerText={(count) => `Showing ${count} attendance records`}
            lookupOptions={lookupOptions}
            loading={loading}
            saving={saving}
            exporting={exporting}
            error={error}
            emptyMessage="No attendance records are available yet."
            searchValue={search}
            onSearchChange={setSearch}
            searchPlaceholder="Search employee, date, or status"
            page={page}
            pageCount={pageCount}
            totalCount={total}
            onPageChange={setPage}
            onExport={exportRows}
            exportFileName="attendance.csv"
          />
        </div>
        <div className="side-stack">
          <PageCard title="Exception summary" subtitle="Live attendance conditions">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Late", "Remote", "Missing checkout", "Locations active"])} />
          </PageCard>
          <PageCard title="Activity stream" subtitle="Latest attendance activity">
            <ActivityList items={insights.activities?.length ? insights.activities : [{ title: "Loading attendance", text: "records from the backend.", meta: "Please wait", tone: "teal" }]} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminLeavePage() {
  const insights = useAdminLeaveInsights();
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(adminLeaveConfig);

  return (
    <RoleLayout role="admin" config={adminConfigs.leave}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <CrudSection
          title="Leave request queue"
          subtitle="Review typed leave requests, approvals, and coverage impact."
          recordLabel={config.recordLabel}
          actions={<div className="filter-pills"><span className="pill pill--teal">All requests</span><span className="pill pill--soft">Pending first</span></div>}
          filters={["Leave type: All", "Status: Mixed", "Coverage risk: Highlighted"]}
          primaryLabel="Create request"
          columns={["Employee", "Dates", "Leave type", "Status", "Approver", "Updated"]}
          variant="employees"
          rows={rows}
          fields={config.fields}
          createDefaults={config.createDefaults}
          onSave={saveRecord}
          onDelete={(row) => removeRecord(row.id)}
          footerText={(count) => `Showing ${count} leave requests`}
          lookupOptions={lookupOptions}
          loading={loading}
          saving={saving}
          exporting={exporting}
          error={error}
          emptyMessage="No leave requests are available yet."
          searchValue={search}
          onSearchChange={setSearch}
          searchPlaceholder="Search employee, leave type, or approver"
          page={page}
          pageCount={pageCount}
          totalCount={total}
          onPageChange={setPage}
          onExport={exportRows}
          exportFileName="leave-requests.csv"
        />
        <div className="side-stack">
          <PageCard title="Leave structure" subtitle="Live type and status mix">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Pending", "Review", "Approved", "Rejected"])} />
          </PageCard>
          <PageCard title="Pending approvals" subtitle="Current request queue">
            <ActionList items={insights.actions?.length ? insights.actions : [{ initials: "--", name: "Loading queue", detail: "Pulling live leave requests.", status: "Loading", tone: "soft" }]} notice={{ text: "Leave approvals stay synced with backend status changes.", action: "Live queue" }} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminJobsPage() {
  const insights = useAdminJobsInsights();
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(jobConfig);

  return (
    <RoleLayout role="admin" config={adminConfigs.jobs}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <CrudSection
          title="Job catalog"
          subtitle="Maintain PT. CODEID job roles, department ownership, descriptions, and levels."
          recordLabel={config.recordLabel}
          actions={<div className="filter-pills"><span className="pill pill--teal">All roles</span><span className="pill pill--soft">Level-aware</span></div>}
          filters={["Department: All", "Level: All", "Compensation: Active"]}
          primaryLabel="Add job"
          columns={["Role", "Primary department", "Level", "Salary range", "Updated"]}
          variant="employees"
          rows={rows}
          fields={config.fields}
          createDefaults={config.createDefaults}
          onSave={saveRecord}
          onDelete={(row) => removeRecord(row.id)}
          getViewPath={(row) => `/admin/jobs/${row.id}`}
          footerText={(count) => `Showing ${count} job records`}
          lookupOptions={lookupOptions}
          loading={loading}
          saving={saving}
          exporting={exporting}
          error={error}
          emptyMessage="No jobs are available yet."
          searchValue={search}
          onSearchChange={setSearch}
          searchPlaceholder="Search job title or level"
          page={page}
          pageCount={pageCount}
          totalCount={total}
          onPageChange={setPage}
          onExport={exportRows}
          exportFileName="jobs.csv"
        />
        <div className="side-stack">
          <PageCard title="Catalog summary" subtitle="Live role coverage">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Roles with descriptions", "Departments", "Levels", "Salary bands"])} />
          </PageCard>
          <PageCard title="Catalog guidance" subtitle="Jobs stay tied to one primary department.">
            <div className="notice-bar notice-bar--soft">
              <span>Use the primary department field to keep job ownership clear, even when the role is reused in staffing plans.</span>
              <strong>Catalog aligned</strong>
            </div>
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminHolidaysPage() {
  const insights = useAdminHolidayInsights();
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(holidayConfig);

  return (
    <RoleLayout role="admin" config={adminConfigs.holidays}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <CrudSection
          title="Holiday calendar"
          subtitle="Manage company-wide and location-specific holidays by year."
          recordLabel={config.recordLabel}
          actions={<div className="filter-pills"><span className="pill pill--teal">Current year</span><span className="pill pill--soft">All locations</span></div>}
          filters={["Year: Current", "Scope: Company + branch", "Status: Planned"]}
          primaryLabel="Add holiday"
          columns={["Holiday", "Date", "Year", "Location", "Updated"]}
          variant="departments"
          rows={rows}
          fields={config.fields}
          createDefaults={config.createDefaults}
          onSave={saveRecord}
          onDelete={(row) => removeRecord(row.id)}
          footerText={(count) => `Showing ${count} holiday records`}
          lookupOptions={lookupOptions}
          loading={loading}
          saving={saving}
          exporting={exporting}
          error={error}
          emptyMessage="No holiday records are available yet."
          searchValue={search}
          onSearchChange={setSearch}
          searchPlaceholder="Search holiday name or location"
          page={page}
          pageCount={pageCount}
          totalCount={total}
          onPageChange={setPage}
          onExport={exportRows}
          exportFileName="holidays.csv"
        />
        <div className="side-stack">
          <PageCard title="Calendar summary" subtitle="Year-based live planning">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Current year", "Company-wide", "Location scoped", "Next holiday"])} />
          </PageCard>
          <PageCard title="Planning guidance" subtitle="Next upcoming calendar milestone">
            <div className="notice-bar notice-bar--soft">
              <span>{insights.notice?.text ?? "Loading holiday planning data from the backend."}</span>
              <strong>{insights.notice?.action ?? "Please wait"}</strong>
            </div>
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminAuditLogsPage() {
  const { accessToken } = useAuth();
  const insights = useAdminAuditInsights();
  const [rows, setRows] = useState<TableRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [meta, setMeta] = useState({ page: 1, limit: 10, total: 0 });
  const [exporting, setExporting] = useState(false);

  useEffect(() => {
    if (!accessToken) {
      setLoading(false);
      return;
    }

    let active = true;
    setLoading(true);
    setError(null);

    void listAuditLogs(accessToken, { page, limit: 10, search: search.trim() || undefined, order: "desc", sort: "created_at" })
      .then((response) => {
        if (!active) return;
        setRows(response.items.map(toAuditRow));
        setMeta(response.meta);
      })
      .catch((loadError) => {
        if (!active) return;
        setError(loadError instanceof Error ? loadError.message : "Could not load audit logs");
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [accessToken, page, search]);

  const exportLogs = async () => {
    if (!accessToken) return;
    setExporting(true);
    setError(null);
    try {
      const firstPage = await listAuditLogs(accessToken, { page: 1, limit: 100, search: search.trim() || undefined, order: "desc", sort: "created_at" });
      const allRows = [...firstPage.items.map(toAuditRow)];
      const pageCount = Math.max(1, Math.ceil(firstPage.meta.total / firstPage.meta.limit));
      for (let currentPage = 2; currentPage <= pageCount; currentPage += 1) {
        const response = await listAuditLogs(accessToken, { page: currentPage, limit: 100, search: search.trim() || undefined, order: "desc", sort: "created_at" });
        allRows.push(...response.items.map(toAuditRow));
      }
      downloadAuditCsv(allRows);
    } catch (exportError) {
      setError(exportError instanceof Error ? exportError.message : "Could not export audit logs");
    } finally {
      setExporting(false);
    }
  };

  return (
    <RoleLayout role="admin" config={adminConfigs.audit}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <PageCard title="Audit stream" subtitle="Track sensitive changes and system events." actions={<div className="filter-pills"><span className="pill pill--teal">All entities</span><span className="pill pill--soft">High priority</span></div>}>
          <div className="section-toolbar">
            <div className="filter-row filter-row--compact"><span className="chip">Entity: All</span><span className="chip">Actor: All</span><span className="chip">Severity: Mixed</span></div>
            <div className="section-toolbar__actions">
              <label className="table-search">
                <span>Search</span>
                <input type="search" value={search} onChange={(event) => { setSearch(event.target.value); setPage(1); }} placeholder="Search actor, entity, or action" />
              </label>
              <button className="ghost-button" type="button" onClick={exportLogs} disabled={loading || exporting}>
                {exporting ? "Exporting..." : "Export CSV"}
              </button>
            </div>
          </div>
          {error ? <div className="notice-bar notice-bar--soft"><span>{error}</span><strong>Check request</strong></div> : null}
          <DataTable
            columns={["Actor", "Entity", "Action", "Record", "Time", "Severity"]}
            rows={rows}
            variant="departments"
            emptyMessage={loading ? "Loading audit logs..." : "No audit logs available yet."}
            footer={
              <>
                <span>Showing {meta.total} audit log entries</span>
                <div className="pagination">
                  <button className="pagination__item" onClick={() => setPage(1)} type="button" disabled={page <= 1}>«</button>
                  <button className="pagination__item" onClick={() => setPage((current) => Math.max(1, current - 1))} type="button" disabled={page <= 1}>‹</button>
                  <span className="pagination__item pagination__item--active">{page}</span>
                  <span className="pagination__text">of {Math.max(1, Math.ceil(meta.total / meta.limit || 1))}</span>
                  <button className="pagination__item" onClick={() => setPage((current) => Math.min(Math.max(1, Math.ceil(meta.total / meta.limit || 1)), current + 1))} type="button" disabled={page >= Math.max(1, Math.ceil(meta.total / meta.limit || 1))}>›</button>
                  <button className="pagination__item" onClick={() => setPage(Math.max(1, Math.ceil(meta.total / meta.limit || 1)))} type="button" disabled={page >= Math.max(1, Math.ceil(meta.total / meta.limit || 1))}>»</button>
                </div>
              </>
            }
          />
        </PageCard>
        <div className="side-stack">
          <PageCard title="Entity coverage" subtitle="Live audit coverage snapshot">
            <MiniStatGrid items={insights.miniStats.length ? insights.miniStats : fallbackMiniStats(["Latest event", "High priority", "Actors", "Entities"])} />
          </PageCard>
          <PageCard title="Security notes" subtitle="Latest audit activity">
            <ActivityList items={insights.activities?.length ? insights.activities : [{ title: "Loading activity", text: "from the backend.", meta: "Please wait", tone: "teal" }]} />
          </PageCard>
        </div>
      </div>
    </RoleLayout>
  );
}

function toAuditRow(item: AuditLogResource): TableRow {
  const actor = item.actor_email ?? "System";
  const severity = ["delete"].includes(item.action.toLowerCase()) ? "High" : "Normal";

  return {
    id: item.id,
    initials: actor.slice(0, 2).toUpperCase(),
    name: actor,
    meta: `${item.entity} ${item.action}`,
    cols: [
      item.entity,
      item.action,
      item.entity_id ?? "--",
      new Intl.DateTimeFormat("en-GB", { day: "2-digit", month: "short", hour: "2-digit", minute: "2-digit", hour12: false }).format(new Date(item.created_at)),
      severity,
    ],
    pillIndex: 4,
    pillTone: severity === "High" ? "warning" : "teal",
    tone: severity === "High" ? "warning" : "default",
  };
}
