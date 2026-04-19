import { useMemo } from "react";
import { CrudSection } from "../../components/CrudSection";
import { MiniStatGrid } from "../../components/MiniStatGrid";
import { PageCard } from "../../components/PageCard";
import { StatGrid } from "../../components/StatGrid";
import { useAuth } from "../../auth/AuthContext";
import { employeeConfig } from "../../data/appData";
import { useApiCrudResource } from "../../hooks/useApiCrudResource";
import { useEmployeeDashboardInsights } from "../../hooks/useLiveInsights";
import { RoleLayout } from "../../layouts/RoleLayout";
import { employeeAttendanceConfig, employeeLeaveConfig } from "../../resources/resourceConfigs";
import type { MiniStatItem } from "../../types";

function fallbackMiniStats(labels: string[]): MiniStatItem[] {
  return labels.map((label) => ({ label, value: "--" }));
}

export function EmployeeDashboardPage() {
  const insights = useEmployeeDashboardInsights();
  return (
    <RoleLayout role="employee" config={employeeConfig}>
      <StatGrid stats={insights.stats} />
      <div className="content-grid">
        <PageCard title="My overview" subtitle="Track attendance, leave requests, and profile context from one place.">
          <MiniStatGrid items={insights.overview.length ? insights.overview : fallbackMiniStats(["Latest attendance", "Manager", "Office", "Primary role"])} />
        </PageCard>
        <PageCard title="Balance summary" subtitle="Live personal work status">
          <MiniStatGrid items={insights.balance.length ? insights.balance : fallbackMiniStats(["Approved leaves", "Pending approvals", "Healthy entries", "Secondary roles"])} />
        </PageCard>
      </div>
    </RoleLayout>
  );
}

export function EmployeeAttendancePage() {
  const { user } = useAuth();
  const isRemoteWorker = user?.workMode === "remote" || user?.workMode === "client-based";
  const attendanceConfig = useMemo(() => ({
    ...employeeAttendanceConfig,
    fields: isRemoteWorker
      ? employeeAttendanceConfig.fields.filter((field) => field.name !== "status")
      : employeeAttendanceConfig.fields,
    createDefaults: {
      ...employeeAttendanceConfig.createDefaults,
      status: isRemoteWorker ? "remote" : "on time",
    },
  }), [isRemoteWorker]);
  const { rows, config, saveRecord, removeRecord, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(attendanceConfig);

  return (
    <RoleLayout role="employee" config={employeeConfig}>
      <CrudSection
        title="My attendance"
        subtitle={isRemoteWorker ? "Create, adjust, and review personal attendance entries. Remote status is applied automatically from your employee profile." : "Create, adjust, and review personal attendance entries."}
        recordLabel={config.recordLabel}
        actions={<div className="filter-pills"><span className="pill pill--teal">This month</span><span className="pill pill--soft">Recent first</span></div>}
        filters={["Status: All", "Shift: Primary", "Location: Mixed"]}
        primaryLabel="Add attendance"
        columns={["Entry", "Check in", "Check out", "Status", "Location"]}
        variant="employees"
        rows={rows}
        fields={config.fields}
        createDefaults={config.createDefaults}
        onSave={saveRecord}
        onDelete={(row) => removeRecord(row.id)}
        footerText={(count) => `Showing ${count} personal attendance entries`}
        loading={loading}
        saving={saving}
        exporting={exporting}
        error={error}
        emptyMessage="No personal attendance entries are available yet."
        searchValue={search}
        onSearchChange={setSearch}
        searchPlaceholder="Search date, status, or location"
        page={page}
        pageCount={pageCount}
        totalCount={total}
        onPageChange={setPage}
        onExport={exportRows}
        exportFileName="my-attendance.csv"
      />
    </RoleLayout>
  );
}

export function EmployeeLeavePage() {
  const { rows, config, saveRecord, removeRecord, lookupOptions, loading, saving, exporting, error, search, setSearch, page, pageCount, total, setPage, exportRows } = useApiCrudResource(employeeLeaveConfig);

  return (
    <RoleLayout role="employee" config={employeeConfig}>
      <CrudSection
        title="My leave"
        subtitle="Create, adjust, and track personal leave requests."
        recordLabel={config.recordLabel}
        actions={<div className="filter-pills"><span className="pill pill--teal">All requests</span><span className="pill pill--soft">Pending</span></div>}
        filters={["Type: All", "Status: Mixed", "Balance: Current"]}
        primaryLabel="Request leave"
        columns={["Request", "Dates", "Leave type", "Status", "Approver", "Updated"]}
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
        searchPlaceholder="Search leave type or status"
        page={page}
        pageCount={pageCount}
        totalCount={total}
        onPageChange={setPage}
        onExport={exportRows}
        exportFileName="my-leave.csv"
      />
    </RoleLayout>
  );
}
