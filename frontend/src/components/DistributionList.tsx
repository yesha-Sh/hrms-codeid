import type { DistributionItem } from "../types";

export function DistributionList({ items }: { items: DistributionItem[] }) {
  return (
    <div className="mini-chart-list">
      {items.map((item) => (
        <div className="mini-chart-item" key={item.label}>
          <div className="progress-meta">
            <span>{item.label}</span>
            <span>{item.value}</span>
          </div>
          <div className="progress-track">
            <span className={`progress-fill progress-fill--${item.tone}`} style={{ width: item.width }} />
          </div>
        </div>
      ))}
    </div>
  );
}