import { clearNode, el } from "../utils/dom.js";
import { formatNumber } from "../utils/format.js";

export function renderSparkline(container, { title, points, color = "#25624a", valueFormatter = formatNumber, className = "" }) {
  if (!points?.length) {
    if (!container.childNodes.length) {
      container.appendChild(el("div", { className: "empty-state" }, "Not enough data for a chart."));
    }
    return;
  }

  const width = 360;
  const height = 160;
  const yTop = 16;
  const yBottom = height - 18;
  const values = points.map((point) => Number(point.value || 0));
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const step = points.length === 1 ? width : width / (points.length - 1);

  const path = points.map((point, index) => {
    const x = index * step;
    const y = yBottom - ((Number(point.value || 0) - min) / range) * (yBottom - yTop);
    return `${index === 0 ? "M" : "L"} ${x.toFixed(2)} ${y.toFixed(2)}`;
  }).join(" ");

  const latest = points[points.length - 1];
  container.appendChild(el("article", { className: ["chart-card", className].filter(Boolean).join(" ") }, [
    el("div", { className: "panel__header" }, [
      el("div", {}, [
        el("span", { className: "panel__eyebrow" }, "Chart"),
        el("h3", { className: "panel__title" }, title),
      ]),
      el("span", { className: "badge badge--accent" }, valueFormatter(latest.value)),
    ]),
    createSparklineSVG(path, width, height, color, min, max, points, valueFormatter),
  ]));
}

function createSparklineSVG(path, width, height, color, min, max, points, valueFormatter) {
  const range = max - min || 1;
  const yTop = 16;
  const yBottom = height - 18;
  const step = points.length === 1 ? width : width / (points.length - 1);

  const tickLevels = [max, min + range * 0.5, min];
  const grid = tickLevels.map((value) => {
    const y = yBottom - ((value - min) / range) * (yBottom - yTop);
    return `
      <line x1="0" y1="${y.toFixed(2)}" x2="${width}" y2="${y.toFixed(2)}" stroke="rgba(30,36,34,0.14)" stroke-dasharray="4 6"></line>
      <text x="0" y="${Math.max(12, y - 6).toFixed(2)}" font-size="12" fill="#5c6762">${valueFormatter(value)}</text>
    `;
  }).join("");

  const markers = points.map((point, index) => {
    const x = index * step;
    const y = yBottom - ((Number(point.value || 0) - min) / range) * (yBottom - yTop);
    const anchor = index === 0 ? "start" : index === points.length - 1 ? "end" : "middle";
    return `
      <circle cx="${x.toFixed(2)}" cy="${y.toFixed(2)}" r="4" fill="${color}">
        <title>${point.label ?? index}: ${valueFormatter(point.value)}</title>
      </circle>
      <text x="${x.toFixed(2)}" y="${height - 2}" font-size="11" text-anchor="${anchor}" fill="#5c6762">${point.label ?? index}</text>
    `;
  }).join("");

  const wrapper = document.createElement("div");
  wrapper.innerHTML = `
    <svg class="chart-card__svg" viewBox="0 0 ${width} ${height}" role="img" aria-label="chart">
      ${grid}
      <path d="${path}" fill="none" stroke="${color}" stroke-width="4" stroke-linecap="round" stroke-linejoin="round"></path>
      ${markers}
    </svg>
  `;
  return wrapper.firstElementChild;
}

export function renderBarChart(container, { title, items, valueFormatter = formatNumber }) {
  clearNode(container);
  if (!items?.length) {
    container.appendChild(el("div", { className: "empty-state" }, "No comparison data."));
    return;
  }

  const max = Math.max(...items.map((item) => Number(item.value || 0)), 1);
  container.appendChild(el("div", { className: "bar-chart" }, [
    el("div", { className: "panel__header" }, [
      el("div", {}, [
        el("span", { className: "panel__eyebrow" }, "Comparison"),
        el("h3", { className: "panel__title" }, title),
      ]),
    ]),
    ...items.map((item) => el("div", { className: "bar-chart__row" }, [
      el("span", {}, item.label),
      el("div", { className: "bar-chart__track" }, el("div", {
        className: "bar-chart__fill",
        style: `width:${Math.max(6, (Number(item.value || 0) / max) * 100)}%;${item.color ? `background:${item.color};` : ""}`,
      })),
      el("strong", {}, valueFormatter(item.value)),
    ])),
  ]));
}
