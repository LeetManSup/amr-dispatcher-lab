import { clearNode, el } from "../utils/dom.js";

export function renderMetricCards(container, items) {
  clearNode(container);
  if (!items?.length) {
    container.appendChild(el("div", { className: "empty-state" }, "No metrics yet."));
    return;
  }
  for (const item of items) {
    container.appendChild(el("article", { className: "metric-card" }, [
      el("span", { className: "metric-card__label" }, item.label),
      el("strong", { className: "metric-card__value" }, item.value),
      item.meta ? el("span", { className: "metric-card__meta" }, item.meta) : null,
    ]));
  }
}

export function renderSummaryList(container, items) {
  clearNode(container);
  if (!items?.length) {
    container.appendChild(el("div", { className: "empty-state" }, "Nothing to show."));
    return;
  }
  for (const item of items) {
    const valueNode = el("div", { className: "summary-item__value" });
    if (item.value instanceof Node) {
      valueNode.appendChild(item.value);
    } else {
      valueNode.textContent = item.value;
    }

    container.appendChild(el("article", { className: "summary-item" }, [
      el("span", { className: "summary-item__label" }, item.label),
      valueNode,
      item.meta ? el("div", { className: "summary-item__meta" }, item.meta) : null,
    ]));
  }
}
