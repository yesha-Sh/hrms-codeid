import type { StatCard } from "../types";

const iconLabelMap = {
  default: "EM",
  accent: "AT",
  warning: "LV",
  cool: "DP",
};

export function StatGrid({ stats }: { stats: StatCard[] }) {
  return (
    <div className="stat-grid">
      {stats.map((stat, index) => {
        const tone = stat.tone ?? "default";
        return (
          <article key={stat.id ?? `${stat.label}-${tone}-${index}`} className={["stat-card", tone !== "default" ? `stat-card--${tone}` : ""].filter(Boolean).join(" ")}>
            <div className="stat-card__head">
              <span>{stat.label}</span>
              <span className="stat-card__icon">{iconLabelMap[tone]}</span>
            </div>
            <div className="stat-card__value">{stat.value}</div>
            <p className="stat-card__detail">{stat.detail}</p>
          </article>
        );
      })}
    </div>
  );
}
