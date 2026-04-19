import type { MiniStatItem } from "../types";

export function MiniStatGrid({
  items,
  columns = 2,
}: {
  items: MiniStatItem[];
  columns?: 2 | 3;
}) {
  return (
    <div className={`mini-grid ${columns === 3 ? "mini-grid--three" : "mini-grid--two"}`}>
      {items.map((item, index) => (
        <div className="mini-stat" key={item.id ?? `${item.label}-${index}`}>
          <span>{item.label}</span>
          <strong>{item.value}</strong>
        </div>
      ))}
    </div>
  );
}
