import { renderDataTable } from "../components/table.js";
import { bindMapControls, renderMap } from "../components/map.js";
import { clearNode, el } from "../utils/dom.js";
import { formatPercent, truncate } from "../utils/format.js";
import { renderBadge } from "../components/status.js";
import { taskPathSummary, taskPhase, taskPhaseTone } from "../utils/task.js";

export function createMapTrafficView(ctx) {
  return {
    id: "map-traffic",
    title: "Map & Traffic",
    description: "Topology view with route overlays, robots and segment load.",
    fragment: "map-traffic",
    async refresh(root) {
      const runId = ctx.store.getState().currentRunId;
      if (!runId) return;
      bindMapControls(root);
      const [mapData, robots, tasks, segmentLoads] = await Promise.all([
        ctx.api.getMap(runId),
        ctx.api.getRobots(runId),
        ctx.api.getTasks(runId),
        ctx.api.getSegmentLoads(runId),
      ]);

      const latestTick = Math.max(...segmentLoads.map((item) => item.tick), -1);
      const latestLoads = segmentLoads.filter((item) => item.tick === latestTick);
      const selection = ctx.store.getState().selection;
      if (selection?.kind === "segment") {
        const freshSegment = mapData.segments.find((item) => item.segmentId === selection.data.segmentId);
        const freshLoad = latestLoads.find((item) => item.segmentId === selection.data.segmentId);
        if (freshSegment) {
          ctx.setSelection({ kind: "segment", data: freshSegment, meta: { load: freshLoad?.load ?? 0 } });
        } else {
          ctx.setSelection(null);
        }
      } else if (selection?.kind === "point") {
        const freshPoint = mapData.points.find((item) => item.pointId === selection.data.pointId);
        if (freshPoint) {
          ctx.setSelection({ kind: "point", data: freshPoint });
        } else {
          ctx.setSelection(null);
        }
      }

      renderLegend(root.querySelector("#mapLegend"));
      renderMap(root.querySelector("#mapCanvas"), mapData, {
        robots,
        tasks,
        segmentLoads: latestLoads,
        selection: ctx.store.getState().selection,
        onSelect: (payload) => ctx.setSelection(payload),
      });

      const activePaths = tasks.filter((task) => ["assigned", "in_progress"].includes(task.taskStatus));
      renderDataTable(root.querySelector("#activePathsTable"), {
        rows: activePaths,
        rowKey: "taskId",
        selectedKey: ctx.store.getState().selection?.kind === "task" ? ctx.store.getState().selection.data.taskId : null,
        onRowClick: (task) => ctx.setSelection({ kind: "task", data: task }),
        emptyText: "No active routes.",
        columns: [
          { key: "taskId", label: "Task", render: (value) => truncate(value) },
          { key: "robotId", label: "Robot" },
          { key: "taskStatus", label: "Phase", render: (_, row) => renderBadge(taskPhase(row), taskPhaseTone(row)) },
          { key: "currentTargetPoint", label: "Current leg" },
          { key: "plannedPath", label: "Path", render: (_, row) => taskPathSummary(row) },
        ],
      });

      renderDataTable(root.querySelector("#segmentLoadTable"), {
        rows: latestLoads,
        rowKey: "segmentId",
        selectedKey: ctx.store.getState().selection?.kind === "segment" ? ctx.store.getState().selection.data.segmentId : null,
        onRowClick: (segmentLoad) => {
          const segment = mapData.segments.find((item) => item.segmentId === segmentLoad.segmentId);
          if (segment) ctx.setSelection({ kind: "segment", data: segment, meta: { load: segmentLoad.load } });
        },
        emptyText: "No segment load snapshots yet.",
        columns: [
          { key: "segmentId", label: "Segment" },
          { key: "tick", label: "Tick" },
          { key: "load", label: "Load", render: (value) => formatPercent(value) },
        ],
      });
    },
  };
}

function renderLegend(container) {
  clearNode(container);
  container.appendChild(createLegendGroup("Segment load", [
    { label: "Low load", kind: "line", color: "#7ea691" },
    { label: "Medium load", kind: "line", color: "#c28b2f" },
    { label: "High load", kind: "line", color: "#bf4c39" },
  ]));
  container.appendChild(createLegendGroup("Overlay objects", [
    { label: "Active path", kind: "path", color: "#25624a" },
    { label: "Robot", kind: "robot", color: "#25624a" },
  ]));
}

function createLegendGroup(title, items) {
  return el("section", { className: "legend__group" }, [
    el("span", { className: "legend__title" }, title),
    el("div", { className: "legend__items" }, items.map((item) => (
      el("span", { className: "legend__item" }, [
        el("span", { className: `legend__swatch legend__swatch--${item.kind}`, style: `--legend-color:${item.color}` }),
        item.label,
      ])
    ))),
  ]);
}
