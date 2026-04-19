import { useMemo, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { Download, PencilLine, Search, Trash2 } from "lucide-react";
import { DataTable } from "./DataTable";
import { PageCard } from "./PageCard";
import type { CrudField, CrudRow, LookupOption } from "../types";

type CrudSectionProps = {
  title: string;
  subtitle: string;
  recordLabel: string;
  actions?: React.ReactNode;
  filters: string[];
  primaryLabel: string;
  columns: string[];
  variant: "employees" | "departments" | "attendance" | "compact";
  rows: CrudRow[];
  fields: CrudField[];
  createDefaults: Record<string, string>;
  onSave: (values: Record<string, string>, id?: string) => Promise<void> | void;
  onDelete?: (row: CrudRow) => Promise<void> | void;
  getViewPath?: (row: CrudRow) => string;
  footerText: (count: number) => string;
  loading?: boolean;
  saving?: boolean;
  exporting?: boolean;
  error?: string | null;
  lookupOptions?: Record<string, LookupOption[]>;
  emptyMessage?: string;
  searchValue?: string;
  onSearchChange?: (value: string) => void;
  searchPlaceholder?: string;
  page?: number;
  pageCount?: number;
  totalCount?: number;
  onPageChange?: (page: number) => void;
  onExport?: () => Promise<CrudRow[]>;
  exportFileName?: string;
};

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

export function CrudSection({
  title,
  subtitle,
  recordLabel,
  actions,
  filters,
  primaryLabel,
  columns,
  variant,
  rows,
  fields,
  createDefaults,
  onSave,
  onDelete,
  getViewPath,
  footerText,
  loading = false,
  saving = false,
  exporting = false,
  error,
  lookupOptions = {},
  emptyMessage,
  searchValue = "",
  onSearchChange,
  searchPlaceholder,
  page = 1,
  pageCount = 1,
  totalCount,
  onPageChange,
  onExport,
  exportFileName,
}: CrudSectionProps) {
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<CrudRow | null>(null);
  const [values, setValues] = useState<Record<string, string>>(createDefaults);
  const [deleteTarget, setDeleteTarget] = useState<CrudRow | null>(null);

  const modalTitle = useMemo(() => (editing ? `Edit ${recordLabel}` : `Create ${recordLabel}`), [editing, recordLabel]);

  const optionsForField = (field: CrudField, sourceValues: Record<string, string> = values) => {
    const baseOptions = lookupOptions[field.name] ?? field.options ?? [];
    if (!field.filterByField) return baseOptions;

    const selectedValue = sourceValues[field.filterByField];
    if (!selectedValue) return baseOptions;

    const contextKey = field.filterContextKey ?? field.filterByField;
    const filtered = baseOptions.filter((option) => `${option.context?.[contextKey] ?? ""}` === selectedValue);
    return filtered.length > 0 ? filtered : baseOptions;
  };

  const syncDependentValues = (nextValues: Record<string, string>, changedField: string) => {
    const synced = { ...nextValues };
    fields
      .filter((field) => field.filterByField === changedField)
      .forEach((field) => {
        const currentValue = synced[field.name];
        if (!currentValue) return;
        const validOptions = optionsForField(field, synced);
        const stillValid = validOptions.some((option) => option.value === currentValue);
        if (!stillValid) {
          synced[field.name] = "";
        }
      });
    return synced;
  };

  const setFieldValue = (fieldName: string, value: string) => {
    setValues((current) => syncDependentValues({ ...current, [fieldName]: value }, fieldName));
  };

  const openCreate = () => {
    setEditing(null);
    setValues(createDefaults);
    setOpen(true);
  };

  const openEdit = (row: CrudRow) => {
    setEditing(row);
    setValues(row.formValues);
    setOpen(true);
  };

  const close = () => {
    setOpen(false);
    setEditing(null);
    setValues(createDefaults);
  };

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    await onSave(values, editing?.id);
    close();
  };

  const confirmDelete = async () => {
    if (!deleteTarget || !onDelete) return;
    await onDelete(deleteTarget);
    setDeleteTarget(null);
  };

  const exportRecords = async () => {
    if (!onExport) return;
    const exportRows = await onExport();
    const fileName = exportFileName ?? `${recordLabel.replaceAll(" ", "-")}-${new Date().toISOString().slice(0, 10)}.csv`;
    downloadCsv(fileName, columns, exportRows);
  };

  return (
    <>
      <PageCard title={title} subtitle={subtitle} actions={actions}>
        <div className="section-toolbar">
          <div className="filter-row filter-row--compact">
            {filters.map((filter) => (
              <span key={filter} className="chip">{filter}</span>
            ))}
          </div>
          <div className="section-toolbar__actions">
            {onSearchChange ? (
              <label className="table-search">
                <Search size={16} />
                <input
                  type="search"
                  value={searchValue}
                  onChange={(event) => onSearchChange(event.target.value)}
                  placeholder={searchPlaceholder ?? `Search ${recordLabel}s`}
                />
              </label>
            ) : null}
            {onExport ? (
              <button className="ghost-button" onClick={exportRecords} type="button" disabled={exporting || loading}>
                <Download size={16} />
                {exporting ? "Exporting..." : "Export CSV"}
              </button>
            ) : null}
            <button className="primary-button primary-button--inline" onClick={openCreate} type="button">
              {primaryLabel}
            </button>
          </div>
        </div>
        {error ? <div className="notice-bar notice-bar--soft"><span>{error}</span><strong>Check request</strong></div> : null}
        <DataTable
          columns={columns}
          rows={rows}
          variant={variant}
          emptyMessage={loading ? "Loading records..." : (emptyMessage ?? "No records available yet.")}
          actions={(row) => {
            const crudRow = row as CrudRow;
            const canEdit = crudRow.canEdit ?? true;
            const canDelete = crudRow.canDelete ?? Boolean(onDelete);

            return (
              <>
                {getViewPath ? (
                  <Link className="table-action table-action--link" to={getViewPath(crudRow)}>
                    Open
                  </Link>
                ) : null}
                {canEdit ? (
                  <button className="table-action table-action--icon" onClick={() => openEdit(crudRow)} type="button" aria-label={`Edit ${crudRow.name}`} title={`Edit ${crudRow.name}`}>
                    <PencilLine size={16} />
                  </button>
                ) : null}
                {onDelete && canDelete ? (
                  <button className="table-action table-action--danger table-action--icon" onClick={() => setDeleteTarget(crudRow)} type="button" aria-label={`Delete ${crudRow.name}`} title={`Delete ${crudRow.name}`}>
                    <Trash2 size={16} />
                  </button>
                ) : null}
                {!canEdit && !canDelete && crudRow.lockedReason ? (
                  <span className="table-muted">{crudRow.lockedReason}</span>
                ) : null}
              </>
            );
          }}
          footer={
            <>
              <span>{footerText(totalCount ?? rows.length)}</span>
              {onPageChange ? (
                <div className="pagination">
                  <button className="pagination__item" onClick={() => onPageChange(1)} type="button" disabled={page <= 1}>«</button>
                  <button className="pagination__item" onClick={() => onPageChange(page - 1)} type="button" disabled={page <= 1}>‹</button>
                  <span className="pagination__item pagination__item--active">{page}</span>
                  <span className="pagination__text">of {pageCount}</span>
                  <button className="pagination__item" onClick={() => onPageChange(page + 1)} type="button" disabled={page >= pageCount}>›</button>
                  <button className="pagination__item" onClick={() => onPageChange(pageCount)} type="button" disabled={page >= pageCount}>»</button>
                </div>
              ) : null}
            </>
          }
        />
      </PageCard>

      {open ? (
        <div className="modal-backdrop" onClick={close} role="presentation">
          <div className="modal-card" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
            <div className="modal-card__head">
              <div>
                <div className="card-title">{modalTitle}</div>
                <div className="modal-card__subtitle">Update the record details and save them to the HRMS backend.</div>
              </div>
              <button className="ghost-button" onClick={close} type="button">Close</button>
            </div>
            <form className="crud-form" onSubmit={submit}>
              {fields.map((field) => (
                <label className="field-block" key={field.name}>
                  <span>{field.label}</span>
                  {field.type === "textarea" ? (
                    <textarea
                      value={values[field.name] ?? ""}
                      onChange={(event) => setFieldValue(field.name, event.target.value)}
                      placeholder={field.placeholder}
                      rows={4}
                    />
                  ) : field.type === "select" ? (
                    <select
                      value={values[field.name] ?? ""}
                      onChange={(event) => setFieldValue(field.name, event.target.value)}
                      required={field.required}
                    >
                      <option value="">{field.emptyOptionLabel ?? `Choose ${field.label.toLowerCase()}`}</option>
                      {optionsForField(field).map((option) => (
                        <option key={option.value} value={option.value}>{option.label}</option>
                      ))}
                    </select>
                  ) : (
                    <input
                      type={field.type ?? "text"}
                      value={values[field.name] ?? ""}
                      onChange={(event) => setFieldValue(field.name, event.target.value)}
                      placeholder={field.placeholder}
                      required={field.required}
                    />
                  )}
                </label>
              ))}
              <div className="crud-form__actions">
                <button className="ghost-button" onClick={close} type="button">Cancel</button>
                <button className="primary-button primary-button--inline" type="submit" disabled={saving}>
                  {saving ? "Saving..." : editing ? "Save changes" : "Create record"}
                </button>
              </div>
            </form>
          </div>
        </div>
      ) : null}

      {deleteTarget ? (
        <div className="modal-backdrop" onClick={() => setDeleteTarget(null)} role="presentation">
          <div className="modal-card modal-card--confirm" onClick={(event) => event.stopPropagation()} role="dialog" aria-modal="true">
            <div className="modal-card__head">
              <div>
                <div className="card-title">Delete {recordLabel}</div>
                <div className="modal-card__subtitle">
                  This will remove <strong>{deleteTarget.name}</strong> from the current workspace view. You can keep going if that is intentional.
                </div>
              </div>
            </div>
            <div className="confirm-actions">
              <button className="ghost-button" onClick={() => setDeleteTarget(null)} type="button">Cancel</button>
              <button className="danger-button" onClick={confirmDelete} type="button" disabled={saving}>Delete record</button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
