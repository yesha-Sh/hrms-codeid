import type { ActivityItem, ActionRow } from "../types";

export function ActivityList({ items }: { items: ActivityItem[] }) {
  return (
    <div className="activity-list">
      {items.map((item, index) => (
        <div className="activity-item" key={item.id ?? `${item.title}-${item.meta}-${index}`}>
          <span className={`dot dot--${item.tone}`} />
          <div>
            <div className="activity-item__title">
              <strong>{item.title}</strong> {item.text}
            </div>
            <div className="activity-item__meta">{item.meta}</div>
          </div>
        </div>
      ))}
    </div>
  );
}

export function ActionList({ items, notice }: { items: ActionRow[]; notice?: { text: string; action: string } }) {
  return (
    <div className="approval-list">
      {items.map((item, index) => (
        <div className="approval-row" key={item.id ?? `${item.name}-${item.status}-${index}`}>
          <div className={`initial-badge initial-badge--${item.tone}`}>{item.initials}</div>
          <div className="approval-row__copy">
            <strong>{item.name}</strong>
            <span>{item.detail}</span>
          </div>
          <span className={`pill pill--${item.tone === "default" ? "teal" : item.tone}`}>{item.status}</span>
        </div>
      ))}
      {notice ? (
        <div className="notice-bar notice-bar--soft">
          <span>{notice.text}</span>
          <strong>{notice.action}</strong>
        </div>
      ) : null}
    </div>
  );
}
