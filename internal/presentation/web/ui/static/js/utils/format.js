export function truncate(value, left = 12, right = 10) {
  if (value == null) return "";
  const text = String(value);
  if (text.length <= left + right + 3) return text;
  return `${text.slice(0, left)}...${text.slice(-right)}`;
}

export function formatNumber(value, digits = 2) {
  if (value == null || Number.isNaN(Number(value))) return "--";
  return Number(value).toFixed(digits);
}

export function formatPercent(value, digits = 0) {
  if (value == null || Number.isNaN(Number(value))) return "--";
  return `${(Number(value) * 100).toFixed(digits)}%`;
}

export function formatTick(value) {
  return value == null ? "--" : `Tick ${value}`;
}

export function formatDateTime(value) {
  if (!value) return "--";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return String(value);
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "medium",
  }).format(date);
}

export function titleFromPath(path) {
  if (!path) return "--";
  const name = path.split("/").pop().replace(/\.[^.]+$/, "");
  return name
    .split(/[-_\s]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join(" ");
}

export function statusTone(status) {
  const normalized = String(status || "").toLowerCase().replaceAll("_", "-");
  if (["completed", "idle", "ok", "created"].includes(normalized)) return "success";
  if (["running", "busy", "assigned", "in-progress"].includes(normalized)) return "accent";
  if (["queued", "paused", "charging", "stopped"].includes(normalized)) return "warning";
  if (["failed", "fault", "cancelled"].includes(normalized)) return "danger";
  return "neutral";
}

export function labelize(status) {
  return String(status || "")
    .replaceAll("_", " ")
    .replaceAll("-", " ")
    .replace(/\b\w/g, (match) => match.toUpperCase());
}

export function sumValues(values) {
  return values.reduce((total, value) => total + Number(value || 0), 0);
}
