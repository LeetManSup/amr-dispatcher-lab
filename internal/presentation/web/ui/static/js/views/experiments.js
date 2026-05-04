import { renderDataTable } from "../components/table.js";
import { renderStatusPill } from "../components/status.js";
import { clearNode, el } from "../utils/dom.js";
import { formatDateTime, truncate } from "../utils/format.js";

export function createExperimentsView(ctx) {
  return {
    id: "experiments",
    title: "Experiments",
    description: "Create new runs, browse fixtures and manage experiment history.",
    fragment: "experiments",
    async refresh(root) {
      const { catalog, runs, currentRunId } = ctx.store.getState();
      fillCatalogSelect(root.querySelector("#mapPathInput"), catalog.mapItems);
      fillCatalogSelect(root.querySelector("#scenarioPathInput"), catalog.scenarioItems);
      renderCatalogList(root.querySelector("#mapCatalogList"), catalog.mapItems);
      renderCatalogList(root.querySelector("#scenarioCatalogList"), catalog.scenarioItems);
      renderDataTable(root.querySelector("#runsTable"), {
        rows: runs,
        rowKey: "id",
        selectedKey: currentRunId,
        onRowClick: (run) => ctx.changeRun(run.id),
        columns: [
          { key: "id", label: "Run", render: (value) => truncate(value, 14, 8) },
          { key: "status", label: "Status", render: (value) => renderStatusPill(value) },
          { key: "algorithm", label: "Algorithm" },
          { key: "currentTick", label: "Tick" },
          { key: "mapName", label: "Map" },
          { key: "scenarioName", label: "Scenario" },
          { key: "createdAt", label: "Created", render: (value) => formatDateTime(value) },
        ],
      });

      const form = root.querySelector("#createRunForm");
      if (!form.dataset.bound) {
        form.dataset.bound = "true";
        form.addEventListener("submit", async (event) => {
          event.preventDefault();
          await ctx.createRun({
            algorithm: root.querySelector("#algorithmInput").value,
            mapPath: root.querySelector("#mapPathInput").value,
            scenarioPath: root.querySelector("#scenarioPathInput").value,
            seed: Number(root.querySelector("#seedInput").value || 0),
          });
        });
      }
    },
  };
}

function fillCatalogSelect(select, items = []) {
  clearNode(select);
  for (const item of items) {
    const option = document.createElement("option");
    option.value = item.path;
    option.textContent = item.label;
    option.title = item.path;
    select.appendChild(option);
  }
}

function renderCatalogList(container, items = []) {
  clearNode(container);
  if (!items.length) {
    container.appendChild(el("div", { className: "empty-state" }, "No fixtures found."));
    return;
  }
  for (const item of items) {
    container.appendChild(el("article", { className: "catalog-item" }, [
      el("div", { className: "catalog-item__title" }, item.label),
      el("div", { className: "catalog-item__path" }, item.path),
      item.description ? el("p", {}, item.description) : null,
    ]));
  }
}
