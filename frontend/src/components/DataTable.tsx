import type { TableRow } from "../types";

export function DataTable({
  columns,
  rows,
  variant,
  footer,
  actions,
  emptyMessage = "No records yet.",
}: {
  columns: string[];
  rows: TableRow[];
  variant: "employees" | "departments" | "attendance" | "compact";
  footer?: React.ReactNode;
  actions?: (row: TableRow) => React.ReactNode;
  emptyMessage?: string;
}) {
  const hasActions = Boolean(actions);
  const detailColumns = columns.slice(1);
  const gridTemplateByVariant = {
    employees: {
      lead: "minmax(190px, 1.8fr)",
      detail: "minmax(74px, 0.74fr)",
      actions: "minmax(118px, 0.82fr)",
    },
    departments: {
      lead: "minmax(190px, 1.75fr)",
      detail: "minmax(78px, 0.76fr)",
      actions: "minmax(116px, 0.8fr)",
    },
    attendance: {
      lead: "minmax(184px, 1.65fr)",
      detail: "minmax(76px, 0.72fr)",
      actions: "minmax(112px, 0.76fr)",
    },
    compact: {
      lead: "minmax(184px, 1.65fr)",
      detail: "minmax(76px, 0.72fr)",
      actions: "minmax(108px, 0.72fr)",
    },
  }[variant];
  const gridTemplate = [
    gridTemplateByVariant.lead,
    ...detailColumns.map(() => gridTemplateByVariant.detail),
    hasActions ? gridTemplateByVariant.actions : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div className="table-shell">
      <div className="table-shell__scroll">
        <div
          className={[`table-head table-head--${variant}`, hasActions ? "table-grid--with-actions" : ""].filter(Boolean).join(" ")}
          style={{ gridTemplateColumns: gridTemplate }}
        >
          {columns.map((column) => (
            <span key={column}>{column}</span>
          ))}
          {hasActions ? <span>Actions</span> : null}
        </div>
        <div className="table-body">
          {rows.length === 0 ? (
            <div className="table-empty">{emptyMessage}</div>
          ) : null}
          {rows.map((row) => (
            <div
              key={row.id ?? row.name}
              className={[`table-row table-row--${variant}`, hasActions ? "table-grid--with-actions" : ""].filter(Boolean).join(" ")}
              style={{ gridTemplateColumns: gridTemplate }}
            >
              <div className="person-cell" data-label={columns[0]}>
                <div className={`initial-badge initial-badge--${row.tone}`}>{row.initials}</div>
                <div className="person-cell__copy">
                  <strong>{row.name}</strong>
                  <span>{row.meta}</span>
                </div>
              </div>
              {row.cols.map((col, index) => {
                const isPill = row.pillIndex === index;
                return (
                  <div key={`${row.name}-${index}`} className="table-cell" data-label={detailColumns[index] ?? `Column ${index + 2}`}>
                    {isPill ? (
                      <span className={`pill pill--${row.pillTone ?? "teal"}`}>
                        {col}
                      </span>
                    ) : (
                      <span>{col}</span>
                    )}
                  </div>
                );
              })}
              {hasActions ? <div className="table-actions" data-label="Actions">{actions?.(row)}</div> : null}
            </div>
          ))}
        </div>
      </div>
      {footer ? <div className="table-footer">{footer}</div> : null}
    </div>
  );
}
