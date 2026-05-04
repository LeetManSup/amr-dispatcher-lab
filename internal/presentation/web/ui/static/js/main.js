import { APIClient } from "./api/client.js";
import { renderInspector } from "./components/inspector.js";
import { renderBadge, toneForStatus } from "./components/status.js";
import { createStore } from "./core/store.js";
import { clearNode, el, mountFragment, qs } from "./utils/dom.js";
import { formatDateTime, formatTick, truncate } from "./utils/format.js";
import { createAnalyticsView } from "./views/analytics.js";
import { createExperimentsView } from "./views/experiments.js";
import { createMapTrafficView } from "./views/map-traffic.js";
import { createOverviewView } from "./views/overview.js";
import { createRobotsView } from "./views/robots.js";
import { createTasksView } from "./views/tasks.js";

const uiConfig = JSON.parse(qs("#ui-config").textContent);
const api = new APIClient(uiConfig.apiBase);

const tabDefinitions = [
  { id: "overview", title: "Overview", description: "Current run and live signals", create: createOverviewView },
  { id: "experiments", title: "Experiments", description: "Create and compare runs", create: createExperimentsView },
  { id: "map-traffic", title: "Map & Traffic", description: "Topology, routes and segment load", create: createMapTrafficView },
  { id: "tasks", title: "Tasks", description: "Queue, filters and ICS actions", create: createTasksView },
  { id: "robots", title: "Robots", description: "Fleet status and assignments", create: createRobotsView },
  { id: "analytics", title: "Analytics", description: "Metrics and run comparison", create: createAnalyticsView },
];

const store = createStore({
  currentTab: normalizeTab(location.hash.replace("#", "")) || "overview",
  currentRunId: "",
  runs: [],
  catalog: { maps: [], scenarios: [], mapItems: [], scenarioItems: [] },
  overview: null,
  selection: null,
  liveEvents: [],
  liveConnection: null,
});

const views = new Map();
let activeView = null;
let refreshToken = 0;

const dom = {
  healthBadge: qs("#healthBadge"),
  runStatusBadge: qs("#runStatusBadge"),
  runTickBadge: qs("#runTickBadge"),
  runSelector: qs("#runSelector"),
  workspaceHeader: qs("#workspaceHeader"),
  workspaceContent: qs("#workspaceContent"),
  tabNav: qs("#tabNav"),
  inspectorPane: qs("#inspectorPane"),
  inspectorContent: qs("#inspectorContent"),
  inspectorCloseButton: qs("#inspectorCloseButton"),
  stepRunButton: qs("#stepRunButton"),
  startRunButton: qs("#startRunButton"),
  stopRunButton: qs("#stopRunButton"),
  exportFormatSelect: qs("#exportFormatSelect"),
  exportRunButton: qs("#exportRunButton"),
};

const context = {
  api,
  store,
  setSelection(selection) {
    store.patch({ selection });
    renderSelection();
  },
  async changeRun(runId) {
    await setCurrentRun(runId);
  },
  async createRun(payload) {
    const run = await api.createRun(payload);
    await refreshRuns();
    await setCurrentRun(run.id);
    store.patch({ currentTab: "overview" });
    location.hash = "#overview";
    await refreshCurrentView();
  },
  async addTask(payload) {
    const runId = store.getState().currentRunId;
    if (!runId) return;
    await api.addTask(runId, payload);
    await refreshCurrentView();
  },
  async cancelTask(task) {
    const runId = store.getState().currentRunId;
    if (!runId) return;
    await api.cancelTask(runId, task.taskId);
    await refreshCurrentView();
  },
  async continueTask(task) {
    const runId = store.getState().currentRunId;
    if (!runId) return;
    await api.continueTask(runId, task.taskId);
    await refreshCurrentView();
  },
  async retargetTask(task) {
    const runId = store.getState().currentRunId;
    if (!runId) return;
    const nextTarget = window.prompt("New target point", task.currentTargetPoint || task.targetPoint || "");
    if (!nextTarget) return;
    await api.retargetTask(runId, task.taskId, nextTarget);
    await refreshCurrentView();
  },
};

for (const tab of tabDefinitions) {
  views.set(tab.id, tab.create(context));
}

window.addEventListener("hashchange", async () => {
  const nextTab = normalizeTab(location.hash.replace("#", ""));
  if (!nextTab || nextTab === store.getState().currentTab) return;
  store.patch({ currentTab: nextTab });
  renderTabNav();
  await renderCurrentView();
});

dom.inspectorCloseButton.addEventListener("click", () => dom.inspectorPane.classList.remove("is-open"));
dom.runSelector.addEventListener("change", async (event) => setCurrentRun(event.target.value));
dom.stepRunButton.addEventListener("click", async () => runAction(() => api.stepRun(store.getState().currentRunId)));
dom.startRunButton.addEventListener("click", async () => runAction(() => api.startRun(store.getState().currentRunId)));
dom.stopRunButton.addEventListener("click", async () => runAction(() => api.stopRun(store.getState().currentRunId)));
dom.exportRunButton.addEventListener("click", async () => {
  const runId = store.getState().currentRunId;
  if (!runId) return;
  const format = dom.exportFormatSelect.value || "json";
  const payload = await api.exportRun(runId, format);
  const contentType = format === "csv" ? "text/csv" : "application/json";
  const blob = new Blob([payload], { type: contentType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = `${runId}.${format}`;
  link.click();
  URL.revokeObjectURL(url);
});

await boot();

async function boot() {
  renderTabNav();
  await Promise.all([refreshHealth(), refreshCatalog(), refreshRuns()]);
  const firstRunId = store.getState().currentRunId || store.getState().runs[0]?.id || "";
  if (firstRunId) {
    await setCurrentRun(firstRunId);
  } else {
    await renderCurrentView();
  }
}

async function refreshHealth() {
  try {
    const health = await api.health();
    updateHealthBadge(health.status);
  } catch (error) {
    updateHealthBadge("error");
    notify(error.message);
  }
}

async function refreshCatalog() {
  const catalog = await api.catalog();
  store.patch({ catalog });
}

async function refreshRuns() {
  const runs = await api.listRuns();
  store.patch({
    runs,
    currentRunId: store.getState().currentRunId || runs[0]?.id || "",
  });
  syncRunSelector();
  renderSelection();
}

async function setCurrentRun(runId) {
  const currentStream = store.getState().liveConnection;
  if (currentStream) {
    currentStream.close();
  }

  if (!runId) {
    store.patch({ currentRunId: "", selection: null, liveConnection: null });
    syncRunSelector();
    await renderToolbar();
    await renderCurrentView();
    return;
  }

  store.patch({ currentRunId: runId, selection: null });
  syncRunSelector();
  await renderToolbar();
  connectLiveStream(runId);
  await refreshCurrentView();
}

async function refreshCurrentView() {
  const token = ++refreshToken;
  await renderToolbar();
  const tabId = store.getState().currentTab;
  const view = views.get(tabId);
  if (!view) return;

  if (!activeView || activeView.id !== tabId) {
    await mountFragment(dom.workspaceContent, view.fragment);
    activeView = view;
    renderWorkspaceHeader(view);
  }

  if (token !== refreshToken) return;
  await view.refresh(dom.workspaceContent);
  renderSelection();
}

async function renderCurrentView() {
  await refreshCurrentView();
}

function renderWorkspaceHeader(view) {
  clearNode(dom.workspaceHeader);
  dom.workspaceHeader.appendChild(el("div", {}, [
    el("span", { className: "panel__eyebrow" }, "Balanced Lab Interface"),
    el("h2", { className: "workspace__title" }, view.title),
    el("p", { className: "workspace__description" }, view.description),
  ]));
}

function renderTabNav() {
  clearNode(dom.tabNav);
  const currentTab = store.getState().currentTab;
  for (const tab of tabDefinitions) {
    const button = el("button", {
      type: "button",
      "aria-selected": currentTab === tab.id ? "true" : "false",
      onclick: async () => {
        store.patch({ currentTab: tab.id });
        location.hash = `#${tab.id}`;
        renderTabNav();
        await renderCurrentView();
      },
    }, [
      el("span", { className: "tab-nav__title" }, tab.title),
      el("span", { className: "tab-nav__description" }, tab.description),
    ]);
    dom.tabNav.appendChild(button);
  }
}

function syncRunSelector() {
  const { runs, currentRunId } = store.getState();
  clearNode(dom.runSelector);
  if (!runs.length) {
    dom.runSelector.appendChild(el("option", { value: "" }, "No runs"));
    dom.runSelector.value = "";
    return;
  }
  for (const run of runs) {
    const option = document.createElement("option");
    option.value = run.id;
    option.textContent = `${run.algorithm.toUpperCase()} | ${run.scenarioName || truncate(run.scenarioPath, 12, 8)} | ${run.currentTick}`;
    option.title = run.id;
    dom.runSelector.appendChild(option);
  }
  dom.runSelector.value = currentRunId || runs[0].id;
}

async function renderToolbar() {
  const runId = store.getState().currentRunId;
  if (!runId) {
    replaceNode("runStatusBadge", renderBadge("No run selected", "neutral"));
    replaceNode("runTickBadge", renderBadge("Tick --", "neutral"));
    toggleRunActions(true);
    return;
  }

  toggleRunActions(false);
  const run = store.getState().runs.find((item) => item.id === runId) || await api.getRun(runId);
  replaceNode("runStatusBadge", renderBadge(labelForRun(run), toneForStatus(run.status)));
  replaceNode("runTickBadge", renderBadge(formatTick(run.currentTick), "accent"));
}

function replaceNode(key, nextNode) {
  const currentNode = dom[key];
  currentNode.replaceWith(nextNode);
  dom[key] = nextNode;
}

function renderSelection() {
  renderInspector(dom.inspectorContent, store.getState().selection, {
    cancelTask: context.cancelTask,
    continueTask: context.continueTask,
    retargetTask: context.retargetTask,
  });
  if (window.innerWidth < 1280 && store.getState().selection?.kind) {
    dom.inspectorPane.classList.add("is-open");
  }
}

function connectLiveStream(runId) {
  const stream = api.eventStream(runId);
  stream.onmessage = async (event) => {
    const payload = JSON.parse(event.data);
    store.update((state) => ({
      ...state,
      liveEvents: [payload, ...state.liveEvents].slice(0, 20),
      liveConnection: stream,
    }));
    await refreshRuns();
    if (store.getState().currentRunId === payload.runId) {
      await refreshCurrentView();
    }
  };
  stream.onerror = () => {
    stream.close();
  };
  store.patch({ liveConnection: stream });
}

function updateHealthBadge(status) {
  replaceNode("healthBadge", renderBadge(`Health: ${status}`, status === "ok" ? "success" : "danger"));
}

function toggleRunActions(disabled) {
  dom.stepRunButton.disabled = disabled;
  dom.startRunButton.disabled = disabled;
  dom.stopRunButton.disabled = disabled;
  dom.exportFormatSelect.disabled = disabled;
  dom.exportRunButton.disabled = disabled;
}

async function runAction(action) {
  const runId = store.getState().currentRunId;
  if (!runId) return;
  await action();
  await refreshRuns();
  await refreshCurrentView();
}

function labelForRun(run) {
  return `${run.status} | ${run.algorithm.toUpperCase()} | ${formatDateTime(run.createdAt)}`;
}

function notify(message) {
  console.error(message);
}

function normalizeTab(tab) {
  return tabDefinitions.find((item) => item.id === tab)?.id;
}
