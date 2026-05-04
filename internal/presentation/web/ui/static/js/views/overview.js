import { renderMetricCards, renderSummaryList } from "../components/cards.js";
import { renderDataTable } from "../components/table.js";
import { renderBadge, renderStatusPill } from "../components/status.js";
import { formatNumber, formatPercent, formatTick, truncate } from "../utils/format.js";

export function createOverviewView(ctx) {
  return {
    id: "overview",
    title: "Overview",
    description: "Current run summary, quick KPIs and recent scheduler activity.",
    fragment: "overview",
    async refresh(root) {
      const runId = ctx.store.getState().currentRunId;
      if (!runId) return;

      const overview = await ctx.api.getOverview(runId);
      ctx.store.patch({ overview });
      ctx.setSelection({ kind: "run", data: overview.run });

      const nextBadge = renderBadge(`Mode: ${overview.run.mode}`, "accent");
      const currentBadge = root.querySelector("#overviewModeBadge") || root.querySelector(".panel__header .badge");
      currentBadge.replaceWith(nextBadge);
      nextBadge.id = "overviewModeBadge";

      renderSummaryList(root.querySelector("#overviewHero"), [
        {
          label: "Status",
          value: renderStatusPill(overview.run.status),
          meta: `${overview.run.algorithm} | ${formatTick(overview.run.currentTick)}`,
        },
        {
          label: "Map",
          value: overview.run.mapName || overview.run.mapPath,
          meta: overview.run.mapPath,
        },
        {
          label: "Scenario",
          value: overview.run.scenarioName || overview.run.scenarioPath,
          meta: overview.run.scenarioPath,
        },
        {
          label: "Queue",
          value: `${overview.snapshot.waitingTasks} waiting`,
          meta: `${overview.snapshot.activeTasks} active | ${overview.snapshot.availableRobots} idle robots`,
        },
      ]);

      const latest = overview.latestMetrics || {};
      renderMetricCards(root.querySelector("#overviewMetricsGrid"), [
        { label: "Throughput", value: formatNumber(latest.throughput), meta: "tasks/tick" },
        { label: "Observed wait", value: formatNumber(latest.avgWaitTime), meta: "queued + started tasks" },
        { label: "Avg execution", value: formatNumber(latest.avgExecutionTime), meta: "ticks" },
        { label: "Robot utilization", value: formatPercent(latest.robotUtilization), meta: "fleet load" },
        { label: "Cancel rate", value: formatPercent(latest.cancelRate), meta: "run quality" },
        { label: "Deadline success", value: formatPercent(latest.deadlineSuccessRate), meta: "SLA proxy" },
      ]);

      const alertsContainer = root.querySelector("#overviewAlerts");
      alertsContainer.innerHTML = "";
      if (overview.alerts?.length) {
        for (const alert of overview.alerts) {
          const node = document.createElement("div");
          node.className = "summary-item";
          node.textContent = alert;
          alertsContainer.appendChild(node);
        }
      } else {
        alertsContainer.className = "empty-state";
        alertsContainer.textContent = "No alerts yet.";
      }

      renderDataTable(root.querySelector("#overviewDecisions"), {
        rows: overview.recentDecisions || [],
        emptyText: "No decisions yet.",
        rowKey: "taskId",
        columns: [
          { key: "tick", label: "Tick" },
          { key: "action", label: "Action" },
          { key: "taskId", label: "Task", render: (value) => truncate(value || "--") },
          { key: "robotId", label: "Robot" },
          { key: "reasonCode", label: "Reason" },
          { key: "score", label: "Score", render: (value) => formatNumber(value) },
        ],
      });

      renderEventFeed(root.querySelector("#overviewEventFeed"), ctx.store.getState().liveEvents || []);
    },
  };
}

function renderEventFeed(container, events) {
  container.innerHTML = "";
  if (!events.length) {
    container.className = "event-feed empty-state";
    container.textContent = "Waiting for events.";
    return;
  }

  container.className = "event-feed";
  for (const event of events.slice(0, 8)) {
    const item = document.createElement("article");
    item.className = "event-feed__item";
    item.innerHTML = `
      <div class="event-feed__meta">${formatTick(event.currentTick)} | ${event.kind}</div>
      <strong>${event.message}</strong>
      <div class="event-feed__meta">${event.runStatus} | ${event.activeTasks} active / ${event.waitingTasks} waiting</div>
    `;
    container.appendChild(item);
  }
}
