export function taskPhase(task) {
  const status = String(task?.taskStatus || "").toLowerCase();
  const source = task?.sourcePoint;
  const target = task?.targetPoint;
  const currentTarget = task?.currentTargetPoint;

  if (status === "completed") return "Completed";
  if (status === "cancelled") return "Cancelled";
  if (status === "failed") return "Failed";

  if (source && target && source === target) {
    if (status === "queued" || status === "created") return "Queued direct move";
    if (status === "assigned" || status === "in_progress") return "Direct delivery";
  }

  if (currentTarget && source && currentTarget === source) {
    if (status === "paused") return "Paused before pickup";
    if (status === "queued" || status === "created") return "Waiting for pickup leg";
    if (status === "assigned" || status === "in_progress") return "Heading to pickup";
  }

  if (currentTarget && target && currentTarget === target) {
    if (status === "paused") return "Paused before dropoff";
    if (status === "queued" || status === "created") return "Waiting for dropoff leg";
    if (status === "assigned" || status === "in_progress") return "Heading to dropoff";
  }

  if (status === "queued" || status === "created") return "Queued";
  if (status === "assigned" || status === "in_progress") return "In transit";
  if (status === "paused") return "Paused";
  return "Unknown";
}

export function taskPhaseTone(task) {
  const phase = taskPhase(task).toLowerCase();
  if (phase.includes("completed")) return "success";
  if (phase.includes("cancelled") || phase.includes("failed")) return "danger";
  if (phase.includes("dropoff") || phase.includes("delivery") || phase.includes("transit")) return "accent";
  if (phase.includes("pickup") || phase.includes("queued") || phase.includes("paused")) return "warning";
  return "neutral";
}

export function taskPathSummary(task) {
  if (task?.plannedPath?.length) {
    return task.plannedPath.join(" -> ");
  }
  const source = task?.sourcePoint || "?";
  const target = task?.targetPoint || "?";
  if (source === target) return source;
  return `${source} => ${target}`;
}
