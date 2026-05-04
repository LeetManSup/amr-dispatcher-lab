import { clearNode, el } from "../utils/dom.js";

export function renderDataTable(container, options) {
  const {
    columns,
    rows,
    emptyText = "No data",
    rowKey = "id",
    selectedKey = null,
    onRowClick = null,
  } = options;

  clearNode(container);
  if (!rows?.length) {
    container.appendChild(el("div", { className: "empty-state" }, emptyText));
    return;
  }

  const table = el("table");
  const thead = el("thead");
  const tbody = el("tbody");

  const headerRow = el("tr");
  for (const column of columns) {
    headerRow.appendChild(el("th", {}, column.label));
  }
  thead.appendChild(headerRow);

  for (const row of rows) {
    const key = row[rowKey];
    const tr = el("tr", {
      className: `${onRowClick ? "clickable-row" : ""}${selectedKey != null && key === selectedKey ? " is-selected" : ""}`,
      onclick: onRowClick ? () => onRowClick(row) : null,
    });
    for (const column of columns) {
      const value = column.render ? column.render(row[column.key], row) : row[column.key];
      const td = el("td");
      if (typeof value === "string") {
        td.textContent = value;
      } else if (value instanceof Node) {
        td.appendChild(value);
      } else if (value != null) {
        td.textContent = String(value);
      }
      tr.appendChild(td);
    }
    tbody.appendChild(tr);
  }

  table.append(thead, tbody);
  container.appendChild(el("div", { className: "table-wrap" }, table));
}
