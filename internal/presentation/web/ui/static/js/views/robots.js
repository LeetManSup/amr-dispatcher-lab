import { renderMetricCards } from "../components/cards.js";
import { renderDataTable } from "../components/table.js";
import { renderStatusPill } from "../components/status.js";
import { formatPercent, truncate } from "../utils/format.js";

export function createRobotsView(ctx) {
  return {
    id: "robots",
    title: "Robots",
    description: "Fleet summary, robot state table and current assignments.",
    fragment: "robots",
    async refresh(root) {
      const runId = ctx.store.getState().currentRunId;
      if (!runId) return;
      const robots = await ctx.api.getRobots(runId);
      const selection = ctx.store.getState().selection;
      if (selection?.kind === "robot") {
        const freshRobot = robots.find((robot) => robot.robotId === selection.data.robotId);
        if (freshRobot) {
          ctx.setSelection({ kind: "robot", data: freshRobot });
        } else {
          ctx.setSelection(null);
        }
      }

      renderMetricCards(root.querySelector("#robotSummaryCards"), summarizeRobots(robots));

      const currentSelection = ctx.store.getState().selection;
      renderDataTable(root.querySelector("#robotsTable"), {
        rows: robots,
        rowKey: "robotId",
        selectedKey: currentSelection?.kind === "robot" ? currentSelection.data.robotId : null,
        onRowClick: (robot) => ctx.setSelection({ kind: "robot", data: robot }),
        columns: [
          { key: "robotId", label: "Robot", render: (value) => truncate(value) },
          { key: "state", label: "State", render: (value) => renderStatusPill(value) },
          { key: "currentPoint", label: "Current point" },
          { key: "batteryLevel", label: "Battery", render: (value) => formatPercent((Number(value) || 0) / 100, 0) },
          { key: "currentTaskId", label: "Task", render: (value) => value || "--" },
          { key: "busyTicks", label: "Busy ticks" },
        ],
      });
    },
  };
}

function summarizeRobots(robots) {
  const counts = robots.reduce((acc, robot) => {
    acc[robot.state] = (acc[robot.state] || 0) + 1;
    return acc;
  }, {});
  return [
    { label: "Fleet size", value: String(robots.length), meta: "registered robots" },
    { label: "Busy", value: String(counts.busy || 0), meta: "currently executing" },
    { label: "Charging", value: String(counts.charging || 0), meta: "recovering battery" },
    { label: "Fault", value: String(counts.fault || 0), meta: "requires attention" },
  ];
}
