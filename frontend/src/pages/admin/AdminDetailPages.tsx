import { useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { CrudSection } from "../../components/CrudSection";
import { PageCard } from "../../components/PageCard";
import { useApiRecord, useApiCrudResource } from "../../hooks/useApiCrudResource";
import { RoleLayout } from "../../layouts/RoleLayout";
import {
  createAssignmentConfig,
  departmentConfig as departmentResourceConfig,
  employeeConfig as employeeResourceConfig,
  jobConfig as jobResourceConfig,
  type ApiCrudConfig,
} from "../../resources/resourceConfigs";
import { adminConfigs } from "../../data/appData";
import type { PageConfig } from "../../types";

type DetailResourcePageProps<TResource, TPayload> = {
  config: PageConfig;
  resourceConfig: ApiCrudConfig<TResource, TPayload>;
  recordId: string | undefined;
  backPath: string;
  summaryTitle: string;
  summaryCopy: string;
  metricLabel: string;
  metricValue: (values: Record<string, string>) => string;
};

function FieldList({ fields, values }: { fields: NonNullable<ApiCrudConfig<unknown, unknown>["detailFields"]>; values: Record<string, string> }) {
  return (
    <div className="detail-grid">
      {fields.map((field) => (
        <div className="detail-item" key={field.key}>
          <span>{field.label}</span>
          <strong>{values[field.key] || "Not set"}</strong>
        </div>
      ))}
    </div>
  );
}

function DetailState({
  config,
  title,
  subtitle,
  backPath,
}: {
  config: PageConfig;
  title: string;
  subtitle: string;
  backPath: string;
}) {
  return (
    <RoleLayout role="admin" config={config}>
      <PageCard title={title} subtitle={subtitle}>
        <div className="empty-state">
          <Link className="primary-button primary-button--inline" to={backPath}>Return to list</Link>
        </div>
      </PageCard>
    </RoleLayout>
  );
}

function DetailResourcePage<TResource, TPayload>({
  config,
  resourceConfig,
  recordId,
  backPath,
  summaryTitle,
  summaryCopy,
  metricLabel,
  metricValue,
}: DetailResourcePageProps<TResource, TPayload>) {
  const { row, loading, error } = useApiRecord(resourceConfig, recordId);

  if (loading) {
    return <DetailState config={config} title={`Loading ${summaryTitle.toLowerCase()}`} subtitle="Pulling the latest record from the backend." backPath={backPath} />;
  }

  if (!row) {
    return <DetailState config={config} title={`${summaryTitle} not found`} subtitle={error ?? "The selected record is no longer available."} backPath={backPath} />;
  }

  return (
    <RoleLayout role="admin" config={config}>
      <div className="detail-page">
        <div className="detail-toolbar">
          <Link className="ghost-button" to={backPath}>Back to list</Link>
        </div>
        <div className="content-grid content-grid--profile">
          <PageCard title={summaryTitle} subtitle={summaryCopy}>
            <div className="detail-hero">
              <div className={`initial-badge initial-badge--${row.tone}`}>{row.initials}</div>
              <div>
                <div className="profile-hero__name">{row.name}</div>
                <div className="muted-line">{row.meta}</div>
              </div>
            </div>
            <FieldList fields={resourceConfig.detailFields ?? []} values={row.formValues} />
          </PageCard>

          <div className="side-stack">
            <PageCard title="Record snapshot" subtitle="Current state in the HRMS API.">
              <div className="mini-grid mini-grid--two">
                <div className="mini-stat"><span>{metricLabel}</span><strong>{metricValue(row.formValues)}</strong></div>
                <div className="mini-stat"><span>Record ID</span><strong>{row.id}</strong></div>
                <div className="mini-stat"><span>Workspace</span><strong>PT. CODEID</strong></div>
                <div className="mini-stat"><span>Profile</span><strong>{row.name}</strong></div>
              </div>
            </PageCard>
            <PageCard title="Next actions" subtitle="Suggested follow-up from this detail view.">
              <div className="quick-actions quick-actions--stacked">
                <Link className="primary-button" to={backPath}>Return to directory</Link>
              </div>
            </PageCard>
          </div>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminEmployeeDetailPage() {
  const { recordId } = useParams();
  const assignmentConfig = useMemo(() => (recordId ? createAssignmentConfig(recordId) : null), [recordId]);
  const { row, loading, error } = useApiRecord(employeeResourceConfig, recordId);
  const assignments = useApiCrudResource(assignmentConfig ?? createAssignmentConfig(""));

  if (loading) {
    return <DetailState config={adminConfigs.employees} title="Loading employee profile" subtitle="Pulling the latest employee record from the backend." backPath="/admin/employees" />;
  }

  if (!row || !recordId || !assignmentConfig) {
    return <DetailState config={adminConfigs.employees} title="Employee profile not found" subtitle={error ?? "The selected employee is no longer available."} backPath="/admin/employees" />;
  }

  const assignmentRows = assignments.rows;
  const activeAssignmentRows = assignmentRows.filter((assignment) => (assignment.formValues.assignment_status ?? "").toLowerCase() === "active");
  const activeWorkload = activeAssignmentRows.reduce((sum, assignment) => sum + Number(assignment.formValues.estimated_hours_per_week ?? 0), 0);

  return (
    <RoleLayout role="admin" config={adminConfigs.employees}>
      <div className="detail-page">
        <div className="detail-toolbar">
          <Link className="ghost-button" to="/admin/employees">Back to directory</Link>
        </div>
        <div className="content-grid content-grid--profile">
          <PageCard title="Employee profile" subtitle="Primary profile data, PT. CODEID placement, and flexible staffing context.">
            <div className="detail-hero">
              <div className={`initial-badge initial-badge--${row.tone}`}>{row.initials}</div>
              <div>
                <div className="profile-hero__name">{row.name}</div>
                <div className="muted-line">{row.meta}</div>
              </div>
            </div>
            <FieldList fields={employeeResourceConfig.detailFields ?? []} values={row.formValues} />
          </PageCard>

          <div className="side-stack">
            <PageCard title="Primary role snapshot" subtitle="Current PT. CODEID assignment.">
              <div className="mini-grid mini-grid--two">
                <div className="mini-stat"><span>Department</span><strong>{row.formValues.department_name || "--"}</strong></div>
                <div className="mini-stat"><span>Primary job</span><strong>{row.formValues.job_title || "--"}</strong></div>
                <div className="mini-stat"><span>Job level</span><strong>{row.formValues.job_level_name || "--"}</strong></div>
                <div className="mini-stat"><span>Status</span><strong>{row.formValues.employment_status_name || "--"}</strong></div>
              </div>
            </PageCard>
            <PageCard title="Assignment guidance" subtitle="Flexible staffing workload is advisory in v1.">
              <div className="notice-bar notice-bar--soft">
                <span>Use secondary assignments to capture cross-functional work without replacing the employee's primary role.</span>
                <strong>Planning only</strong>
              </div>
            </PageCard>
          </div>
        </div>

        <div className="content-grid">
          <CrudSection
            title="Secondary assignments"
            subtitle="Cross-functional roles, managed workload hours, and active date windows for this employee."
            recordLabel={assignments.config.recordLabel}
            actions={<div className="filter-pills"><span className="pill pill--teal">Active assignments</span><span className="pill pill--soft">Workload planning</span></div>}
            filters={["Employee: Current profile", "Scope: PT. CODEID", "Status: All"]}
            primaryLabel="Add assignment"
            columns={["Role", "Department", "Level", "Hours/week", "Active window", "Status", "Updated"]}
            variant="employees"
            rows={assignments.rows}
            fields={assignments.config.fields}
            createDefaults={assignments.config.createDefaults}
            onSave={assignments.saveRecord}
            onDelete={(target) => assignments.removeRecord(target.id)}
            footerText={(count) => `Showing ${count} secondary assignments`}
            lookupOptions={assignments.lookupOptions}
            loading={assignments.loading}
            saving={assignments.saving}
            exporting={assignments.exporting}
            error={assignments.error}
            emptyMessage="No secondary assignments are recorded for this employee yet."
            searchValue={assignments.search}
            onSearchChange={assignments.setSearch}
            searchPlaceholder="Search assignment role or department"
            page={assignments.page}
            pageCount={assignments.pageCount}
            totalCount={assignments.total}
            onPageChange={assignments.setPage}
            onExport={assignments.exportRows}
            exportFileName="secondary-assignments.csv"
          />
          <div className="side-stack">
            <PageCard title="Current staffing note" subtitle="Primary role remains authoritative while secondary assignments stay temporary and managed.">
              <div className="mini-grid mini-grid--two">
                <div className="mini-stat"><span>Primary role</span><strong>{row.formValues.job_title || "--"}</strong></div>
                <div className="mini-stat"><span>Assignments</span><strong>{assignmentRows.length}</strong></div>
                <div className="mini-stat"><span>Location</span><strong>{row.formValues.location_name || "--"}</strong></div>
                <div className="mini-stat"><span>Manager</span><strong>{row.formValues.manager_name || "--"}</strong></div>
              </div>
            </PageCard>
            <PageCard title="Workload overview" subtitle="Active secondary work should stay realistic and time-bound.">
              <div className="mini-grid mini-grid--two">
                <div className="mini-stat"><span>Active roles</span><strong>{activeAssignmentRows.length}</strong></div>
                <div className="mini-stat"><span>Active workload</span><strong>{activeWorkload}h/week</strong></div>
                <div className="mini-stat"><span>Ended roles</span><strong>{assignmentRows.filter((assignment) => (assignment.formValues.assignment_status ?? "").toLowerCase() === "ended").length}</strong></div>
                <div className="mini-stat"><span>Scheduled roles</span><strong>{assignmentRows.filter((assignment) => (assignment.formValues.assignment_status ?? "").toLowerCase() === "scheduled").length}</strong></div>
              </div>
            </PageCard>
          </div>
        </div>
      </div>
    </RoleLayout>
  );
}

export function AdminDepartmentDetailPage() {
  const { recordId } = useParams();

  return (
    <DetailResourcePage
      config={adminConfigs.departments}
      resourceConfig={departmentResourceConfig}
      recordId={recordId}
      backPath="/admin/departments"
      summaryTitle="Department profile"
      summaryCopy="Manager ownership, location, and hierarchy context for the selected department."
      metricLabel="Level"
      metricValue={(values) => values.level || "--"}
    />
  );
}

export function AdminJobDetailPage() {
  const { recordId } = useParams();

  return (
    <DetailResourcePage
      config={adminConfigs.jobs}
      resourceConfig={jobResourceConfig}
      recordId={recordId}
      backPath="/admin/jobs"
      summaryTitle="Job profile"
      summaryCopy="Primary department, job level, and compensation range for the selected role."
      metricLabel="Level"
      metricValue={(values) => values.job_level_name || "--"}
    />
  );
}
