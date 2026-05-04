import { el } from "../utils/dom.js";
import { labelize, statusTone } from "../utils/format.js";

export function renderBadge(text, tone = "neutral") {
  return el("span", { className: `badge badge--${tone}` }, text);
}

export function renderStatusPill(status) {
  const normalized = String(status || "").toLowerCase().replaceAll("_", "-");
  return el("span", { className: `pill pill--${normalized}` }, labelize(status));
}

export function toneForStatus(status) {
  return statusTone(status);
}
