export class APIClient {
  constructor(base = "/api") {
    this.base = base;
  }

  async request(path, options = {}) {
    const response = await fetch(`${this.base}${path}`, {
      headers: { "Content-Type": "application/json", ...(options.headers || {}) },
      ...options,
    });

    if (!response.ok) {
      let message = response.statusText;
      try {
        const payload = await response.json();
        message = payload.error || message;
      } catch {
        // ignore
      }
      throw new Error(message);
    }

    if (response.headers.get("Content-Type")?.includes("application/json")) {
      return response.json();
    }
    return response.text();
  }

  async health() {
    const response = await fetch("/health");
    if (!response.ok) {
      throw new Error(response.statusText);
    }
    return response.json();
  }
  catalog() { return this.request("/catalog"); }
  listRuns() { return this.request("/runs"); }
  createRun(payload) { return this.request("/runs", { method: "POST", body: JSON.stringify(payload) }); }
  getRun(id) { return this.request(`/runs/${id}`); }
  getOverview(id) { return this.request(`/runs/${id}/overview`); }
  getRequests(id) { return this.request(`/runs/${id}/requests`); }
  getTasks(id) { return this.request(`/runs/${id}/tasks`); }
  getRobots(id) { return this.request(`/runs/${id}/robots`); }
  getMetrics(id) { return this.request(`/runs/${id}/metrics`); }
  getMap(id) { return this.request(`/runs/${id}/map`); }
  getDecisions(id) { return this.request(`/runs/${id}/decisions`); }
  getSegmentLoads(id) { return this.request(`/runs/${id}/segment-load`); }
  stepRun(id) { return this.request(`/runs/${id}/step`, { method: "POST" }); }
  startRun(id) { return this.request(`/runs/${id}/start`, { method: "POST" }); }
  stopRun(id) { return this.request(`/runs/${id}/stop`, { method: "POST" }); }
  async exportRun(id, format = "json") {
    const response = await fetch(`${this.base}/runs/${encodeURIComponent(id)}/export?format=${encodeURIComponent(format)}`);
    if (!response.ok) {
      throw new Error(response.statusText);
    }
    return response.text();
  }
  addTask(id, payload) { return this.request(`/runs/${id}/ics/addTask`, { method: "POST", body: JSON.stringify(payload) }); }
  cancelTask(id, taskId) { return this.request(`/runs/${id}/ics/cancelTask?taskId=${encodeURIComponent(taskId)}`, { method: "POST" }); }
  continueTask(id, taskId) { return this.request(`/runs/${id}/ics/continueTask?taskId=${encodeURIComponent(taskId)}`, { method: "POST" }); }
  retargetTask(id, taskId, targetPoint) {
    return this.request(`/runs/${id}/ics/updateOrderPointInfo?taskId=${encodeURIComponent(taskId)}`, {
      method: "POST",
      body: JSON.stringify({ targetPoint }),
    });
  }

  eventStream(id) {
    return new EventSource(`${this.base}/runs/${encodeURIComponent(id)}/stream`);
  }
}
