import { renderDataTable } from "../components/table.js";
import { renderBadge, renderStatusPill } from "../components/status.js";
import { truncate } from "../utils/format.js";
import { taskPathSummary, taskPhase, taskPhaseTone } from "../utils/task.js";

export function createTasksView(ctx) {
  return {
    id: "tasks",
    title: "Tasks",
    description: "Queue management, task filters and ICS operations.",
    fragment: "tasks",
    async refresh(root) {
      const runId = ctx.store.getState().currentRunId;
      if (!runId) return;
      const tasks = await ctx.api.getTasks(runId);
      const selection = ctx.store.getState().selection;
      if (selection?.kind === "task") {
        const freshTask = tasks.find((task) => task.taskId === selection.data.taskId);
        if (freshTask) {
          ctx.setSelection({ kind: "task", data: freshTask });
        } else {
          ctx.setSelection(null);
        }
      }

      const statusFilter = root.querySelector("#taskStatusFilter");
      const robotFilter = root.querySelector("#taskRobotFilter");
      bindFilters(statusFilter, robotFilter, () => this.refresh(root), root);

      const filtered = tasks.filter((task) => {
        const matchesStatus = !statusFilter.value || task.taskStatus === statusFilter.value;
        const matchesRobot = !robotFilter.value || task.robotId?.toLowerCase().includes(robotFilter.value.toLowerCase());
        return matchesStatus && matchesRobot;
      });

      const currentSelection = ctx.store.getState().selection;
      renderDataTable(root.querySelector("#tasksTable"), {
        rows: filtered,
        rowKey: "taskId",
        selectedKey: currentSelection?.kind === "task" ? currentSelection.data.taskId : null,
        onRowClick: (task) => ctx.setSelection({ kind: "task", data: task }),
        columns: [
          { key: "taskId", label: "Task", render: (value) => truncate(value) },
          { key: "requestId", label: "Request", render: (value) => truncate(value) },
          { key: "taskStatus", label: "Status", render: (value) => renderStatusPill(value) },
          { key: "taskStatus", label: "Phase", render: (_, row) => renderBadge(taskPhase(row), taskPhaseTone(row)) },
          { key: "robotId", label: "Robot" },
          { key: "currentTargetPoint", label: "Current leg" },
          { key: "estimatedDuration", label: "ETA" },
          { key: "plannedPath", label: "Path", render: (_, row) => taskPathSummary(row) },
        ],
      });

      const form = root.querySelector("#addTaskForm");
      if (!form.dataset.bound) {
        form.dataset.bound = "true";
        form.addEventListener("submit", async (event) => {
          event.preventDefault();
          await ctx.addTask({
            requestId: root.querySelector("#icsRequestId").value,
            sourcePoint: root.querySelector("#icsSourcePoint").value,
            targetPoint: root.querySelector("#icsTargetPoint").value,
            businessType: root.querySelector("#icsBusinessType").value,
            priority: Number(root.querySelector("#icsPriority").value || 0),
            createdAt: 0,
            deadline: Number(root.querySelector("#icsDeadline").value || 0),
          });
        });
      }
    },
  };
}

function bindFilters(statusFilter, robotFilter, refresh, root) {
  if (root.dataset.filtersBound) return;
  root.dataset.filtersBound = "true";
  statusFilter.addEventListener("change", refresh);
  robotFilter.addEventListener("input", refresh);
}
