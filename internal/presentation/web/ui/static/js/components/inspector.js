import { clearNode, el } from "../utils/dom.js";
import { formatDateTime, formatNumber, formatPercent } from "../utils/format.js";
import { taskPathSummary, taskPhase, taskPhaseTone } from "../utils/task.js";
import { renderBadge, renderStatusPill } from "./status.js";

export function renderInspector(container, selection, actions = {}) {
  clearNode(container);
  if (!selection?.kind || !selection?.data) {
    container.appendChild(el("div", { className: "empty-state" }, [
      el("h3", {}, "No selection"),
      el("p", {}, "Select a run, task, robot or segment to inspect details here."),
    ]));
    return;
  }

  const card = el("section", { className: "inspector-card inspector-grid" });
  const title = inspectorTitle(selection);
  card.appendChild(el("div", {}, [
    el("span", { className: "panel__eyebrow" }, "Selection"),
    el("h3", { className: "panel__title" }, title),
  ]));

  if (selection.kind === "task") {
    appendDefinitionList(card, [
      ["Task ID", selection.data.taskId],
      ["Request", selection.data.requestId],
      ["Status", renderStatusPill(selection.data.taskStatus)],
      ["Phase", renderBadge(taskPhase(selection.data), taskPhaseTone(selection.data))],
      ["Robot", selection.data.robotId || "--"],
      ["Source", selection.data.sourcePoint || "--"],
      ["Current leg", selection.data.currentTargetPoint || selection.data.targetPoint],
      ["Final target", selection.data.targetPoint || "--"],
      ["Path", taskPathSummary(selection.data)],
      ["Duration", `${selection.data.estimatedDuration ?? "--"} ticks`],
      ["Cost", formatNumber(selection.data.estimatedCost)],
    ]);

    const controls = el("div", { className: "toolbar-actions" }, [
      el("button", { className: "button button--danger", onclick: () => actions.cancelTask?.(selection.data) }, "Cancel"),
      el("button", { className: "button button--ghost", onclick: () => actions.continueTask?.(selection.data) }, "Continue"),
      el("button", { className: "button button--ghost", onclick: () => actions.retargetTask?.(selection.data) }, "Retarget"),
    ]);
    card.appendChild(controls);
  } else if (selection.kind === "robot") {
    appendDefinitionList(card, [
      ["Robot ID", selection.data.robotId],
      ["Model", selection.data.robotModel],
      ["State", renderStatusPill(selection.data.state)],
      ["Current point", selection.data.currentPoint],
      ["Battery", formatPercent((selection.data.batteryLevel || 0) / 100, 0)],
      ["Current task", selection.data.currentTaskId || "--"],
      ["Busy ticks", String(selection.data.busyTicks ?? 0)],
    ]);
  } else if (selection.kind === "segment") {
    appendDefinitionList(card, [
      ["Segment ID", selection.data.segmentId],
      ["From", selection.data.fromPoint],
      ["To", selection.data.toPoint],
      ["Length", formatNumber(selection.data.length)],
      ["Speed limit", formatNumber(selection.data.speedLimit)],
      ["Direction", selection.data.direction],
      ["Load", formatPercent(selection.meta?.load ?? 0)],
    ]);
  } else if (selection.kind === "run") {
    appendDefinitionList(card, [
      ["Run ID", selection.data.id],
      ["Status", renderStatusPill(selection.data.status)],
      ["Algorithm", selection.data.algorithm],
      ["Map", selection.data.mapName || selection.data.mapPath],
      ["Scenario", selection.data.scenarioName || selection.data.scenarioPath],
      ["Created", formatDateTime(selection.data.createdAt)],
      ["Tick", String(selection.data.currentTick)],
    ]);
  } else if (selection.kind === "point") {
    appendDefinitionList(card, [
      ["Point ID", selection.data.pointId],
      ["Type", selection.data.pointType],
      ["Area", selection.data.areaId || "--"],
      ["Coordinates", `${formatNumber(selection.data.x, 0)}, ${formatNumber(selection.data.y, 0)}`],
    ]);
  }

  container.appendChild(card);
}

function appendDefinitionList(root, items) {
  const list = el("dl");
  for (const [label, value] of items) {
    list.appendChild(el("dt", {}, label));
    const dd = el("dd");
    if (value instanceof Node) {
      dd.appendChild(value);
    } else {
      dd.textContent = value ?? "--";
    }
    list.appendChild(dd);
  }
  root.appendChild(list);
}

function inspectorTitle(selection) {
  switch (selection.kind) {
    case "task":
      return `Task ${selection.data.taskId}`;
    case "robot":
      return `Robot ${selection.data.robotId}`;
    case "segment":
      return `Segment ${selection.data.segmentId}`;
    case "run":
      return `Run ${selection.data.id}`;
    case "point":
      return `Point ${selection.data.pointId}`;
    default:
      return "Selection";
  }
}
