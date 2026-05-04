import { renderBarChart, renderSparkline } from "../components/charts.js";
import { renderDataTable } from "../components/table.js";
import { clearNode, el } from "../utils/dom.js";
import { formatNumber, formatPercent } from "../utils/format.js";

const REJECT_REASON_LABELS = {
  active_limit_reached: "active limit",
  no_available_robot: "no robot",
  route_overloaded: "route overload",
  zone_overloaded: "zone overload",
};

const REJECT_REASON_COLORS = {
  active_limit_reached: "#5d6fbe",
  no_available_robot: "#bf4c39",
  route_overloaded: "#c28b2f",
  zone_overloaded: "#2f8f6b",
};

export function createAnalyticsView(ctx) {
  return {
    id: "analytics",
    title: "Analytics",
    description: "Run metrics over time and comparison across recent runs.",
    fragment: "analytics",
    async refresh(root) {
      const state = ctx.store.getState();
      const runId = state.currentRunId;
      if (!runId) return;
      const metrics = await ctx.api.getMetrics(runId);
      const currentRun = state.runs.find((item) => item.id === runId) || await ctx.api.getRun(runId);

      const charts = root.querySelector("#analyticsCharts");
      charts.innerHTML = "";
      renderSparkline(charts, {
        title: "Throughput",
        points: metrics.map((item) => ({ label: item.tick, value: item.throughput })),
        className: "chart-card--large",
      });
      renderSparkline(charts, {
        title: "Observed wait",
        points: metrics.map((item) => ({ label: item.tick, value: item.avgWaitTime })),
        color: "#c28b2f",
        className: "chart-card--large",
      });
      renderSparkline(charts, {
        title: "Avg execution",
        points: metrics.map((item) => ({ label: item.tick, value: item.avgExecutionTime })),
        color: "#5d6fbe",
        className: "chart-card--large",
      });
      renderSparkline(charts, {
        title: "Robot utilization",
        points: metrics.map((item) => ({ label: item.tick, value: item.robotUtilization })),
        color: "#bf4c39",
        valueFormatter: (value) => formatPercent(value),
        className: "chart-card--large",
      });

      const comparisonItems = await buildComparison(ctx.api, state.runs, currentRun);
      renderBarChart(root.querySelector("#comparisonChart"), {
        title: "Throughput by algorithm",
        items: comparisonItems.map((item) => ({
          label: `${String(item.algorithm).toUpperCase()} · ${item.scenario}`,
          value: item.throughput,
        })),
      });
      renderBarChart(root.querySelector("#waitP95Chart"), {
        title: "Observed wait P95 by algorithm",
        items: comparisonItems.map((item) => ({
          label: String(item.algorithm).toUpperCase(),
          value: item.waitP95,
        })),
      });
      renderBarChart(root.querySelector("#rejectReasonChart"), {
        title: "Admission rejects by reason",
        items: flattenRejectReasons(comparisonItems),
      });
      renderComparisonHighlights(root.querySelector("#comparisonHighlights"), comparisonItems);
      renderDataTable(root.querySelector("#comparisonTable"), {
        rows: comparisonItems,
        rowKey: "runId",
        selectedKey: runId,
        onRowClick: (item) => ctx.changeRun(item.runId),
        emptyText: "Run more experiments on this scenario to compare algorithms.",
        columns: [
          { key: "algorithm", label: "Algorithm" },
          { key: "status", label: "Status" },
          { key: "completedTasks", label: "Done" },
          { key: "throughput", label: "Throughput", render: (value) => formatNumber(value) },
          { key: "waitP50", label: "Wait P50", render: (value) => formatNumber(value) },
          { key: "waitP95", label: "Wait P95", render: (value) => formatNumber(value) },
          { key: "highPriorityWait", label: "High-prio wait", render: (value) => formatNumber(value) },
          { key: "deadlineMisses", label: "Misses" },
          { key: "rejects", label: "Rejects" },
          { key: "currentTick", label: "Drain tick" },
        ],
      });

      renderDataTable(root.querySelector("#analyticsMetricsTable"), {
        rows: metrics.slice(-20).reverse(),
        rowKey: "tick",
        columns: [
          { key: "tick", label: "Tick" },
          { key: "throughput", label: "Throughput", render: (value) => formatNumber(value) },
          { key: "avgWaitTime", label: "Observed wait", render: (value) => formatNumber(value) },
          { key: "avgExecutionTime", label: "Avg exec", render: (value) => formatNumber(value) },
          { key: "cancelRate", label: "Cancel", render: (value) => formatPercent(value) },
          { key: "deadlineSuccessRate", label: "Deadline", render: (value) => formatPercent(value) },
          { key: "robotUtilization", label: "Utilization", render: (value) => formatPercent(value) },
        ],
      });
    },
  };
}

async function buildComparison(api, runs, currentRun) {
  const sameScenarioRuns = runs.filter((run) => (
    run.mapPath === currentRun.mapPath &&
    run.scenarioPath === currentRun.scenarioPath
  ));
  const selectedRuns = (sameScenarioRuns.length ? sameScenarioRuns : runs).slice(0, 6);
  const [metricLists, taskLists, decisionLists, requestLists] = await Promise.all([
    Promise.all(selectedRuns.map((run) => api.getMetrics(run.id))),
    Promise.all(selectedRuns.map((run) => api.getTasks(run.id))),
    Promise.all(selectedRuns.map((run) => api.getDecisions(run.id))),
    Promise.all(selectedRuns.map((run) => api.getRequests(run.id))),
  ]);

  const results = [];
  for (let index = 0; index < selectedRuns.length; index += 1) {
    const run = selectedRuns[index];
    const metrics = metricLists[index];
    const tasks = taskLists[index];
    const decisions = decisionLists[index];
    const requests = requestLists[index];
    const latest = metrics[metrics.length - 1];
    if (!latest) continue;
    const computed = computeRunComparison(run, tasks, decisions, requests);
    results.push({
      runId: run.id,
      algorithm: run.algorithm,
      status: run.status,
      scenario: run.scenarioName || run.scenarioPath,
      currentTick: run.currentTick,
      completedTasks: latest.completedTasks,
      throughput: latest.throughput,
      waitTime: latest.avgWaitTime,
      waitP50: computed.waitP50,
      waitP95: computed.waitP95,
      highPriorityWait: computed.highPriorityWait,
      avgExec: latest.avgExecutionTime,
      utilization: latest.robotUtilization,
      deadlineMisses: computed.deadlineMisses,
      rejects: computed.rejects,
      rejectReasons: computed.rejectReasons,
    });
  }

  return results.sort((left, right) => {
    if (left.throughput !== right.throughput) {
      return right.throughput - left.throughput;
    }
    if (left.waitTime !== right.waitTime) {
      return left.waitTime - right.waitTime;
    }
    return left.currentTick - right.currentTick;
  });
}

function renderComparisonHighlights(container, items) {
  clearNode(container);
  if (!items.length) {
    container.appendChild(el("div", { className: "empty-state" }, "No comparable runs yet."));
    return;
  }

  const throughputLeader = [...items].sort((left, right) => right.throughput - left.throughput)[0];
  const waitLeader = [...items].sort((left, right) => left.waitP95 - right.waitP95)[0];
  const speedLeader = [...items].sort((left, right) => left.currentTick - right.currentTick)[0];
  const deadlineLeader = [...items].sort((left, right) => left.deadlineMisses - right.deadlineMisses)[0];

  container.appendChild(metricCard("Best throughput", String(throughputLeader.algorithm).toUpperCase(), `${formatNumber(throughputLeader.throughput)} tasks/tick`));
  container.appendChild(metricCard("Best wait tail", String(waitLeader.algorithm).toUpperCase(), `P95 ${formatNumber(waitLeader.waitP95)} ticks`));
  container.appendChild(metricCard("Fewest deadline misses", String(deadlineLeader.algorithm).toUpperCase(), `${deadlineLeader.deadlineMisses} missed tasks`));
  container.appendChild(metricCard("Fastest completion", String(speedLeader.algorithm).toUpperCase(), `Tick ${speedLeader.currentTick}`));
}

function metricCard(label, value, meta) {
  return el("article", { className: "metric-card" }, [
    el("span", { className: "metric-card__label" }, label),
    el("strong", { className: "metric-card__value" }, value),
    el("span", { className: "metric-card__meta" }, meta),
  ]);
}

function computeRunComparison(run, tasks, decisions, requests) {
  const requestIndex = Object.fromEntries(requests.map((item) => [item.requestId, item]));
  const waits = tasks.map((task) => observedWait(task, run.currentTick));
  const highPriorityWaits = tasks
    .filter((task) => (requestIndex[task.requestId]?.priority ?? 0) >= 8)
    .map((task) => observedWait(task, run.currentTick));
  const deadlineMisses = tasks.filter((task) => {
    const request = requestIndex[task.requestId];
    if (!request || request.deadline <= 0) return false;
    if (task.taskStatus === "completed") {
      return task.finishedAt > request.deadline;
    }
    return run.currentTick > request.deadline;
  }).length;
  const rejectReasons = decisions.reduce((acc, item) => {
    if (!Object.hasOwn(REJECT_REASON_LABELS, item.reasonCode)) {
      return acc;
    }
    acc[item.reasonCode] = (acc[item.reasonCode] || 0) + 1;
    return acc;
  }, {});
  const rejects = Object.values(rejectReasons).reduce((sum, count) => sum + count, 0);

  return {
    waitP50: percentile(waits, 0.5),
    waitP95: percentile(waits, 0.95),
    highPriorityWait: average(highPriorityWaits),
    deadlineMisses,
    rejects,
    rejectReasons,
  };
}

function flattenRejectReasons(items) {
  const chartItems = [];
  for (const item of items) {
    for (const [reasonCode, count] of Object.entries(item.rejectReasons || {})) {
      chartItems.push({
        label: `${String(item.algorithm).toUpperCase()} · ${REJECT_REASON_LABELS[reasonCode] ?? reasonCode}`,
        value: count,
        color: REJECT_REASON_COLORS[reasonCode],
      });
    }
  }
  return chartItems.sort((left, right) => right.value - left.value);
}

function observedWait(task, currentTick) {
  if (task.startedAt >= 0) {
    return Math.max(0, task.startedAt - task.createdAt);
  }
  if (task.finishedAt >= 0) {
    return Math.max(0, task.finishedAt - task.createdAt);
  }
  return Math.max(0, currentTick - task.createdAt);
}

function percentile(values, ratio) {
  if (!values.length) return 0;
  const ordered = [...values].sort((left, right) => left - right);
  const index = Math.min(ordered.length - 1, Math.max(0, Math.ceil(ordered.length * ratio) - 1));
  return ordered[index];
}

function average(values) {
  if (!values.length) return 0;
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}
