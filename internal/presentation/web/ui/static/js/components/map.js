import { clearNode } from "../utils/dom.js";

const SVG_NS = "http://www.w3.org/2000/svg";
const MIN_SCALE = 1;
const MAX_SCALE = 3.5;
const ZOOM_STEP = 1.18;

export function renderMap(
  svg,
  mapData,
  { robots = [], tasks = [], segmentLoads = [], selection = null, onSelect = () => {} } = {},
) {
  clearNode(svg);
  if (!mapData?.points?.length) {
    return;
  }

  const bounds = computeBounds(mapData.points);
  ensurePanZoom(svg, bounds);

  const pointIndex = Object.fromEntries(mapData.points.map((point) => [point.pointId, point]));
  const segmentLoadIndex = new Map(segmentLoads.map((item) => [item.segmentId, Number(item.load || 0)]));
  const layer = svgElement("g", { class: "map-layer" });

  for (const segment of mapData.segments || []) {
    const from = pointIndex[segment.fromPoint];
    const to = pointIndex[segment.toPoint];
    if (!from || !to) continue;

    const load = segmentLoadIndex.get(segment.segmentId) || 0;
    const line = svgElement("line", {
      x1: from.x,
      y1: from.y,
      x2: to.x,
      y2: to.y,
      stroke: segmentStroke(load),
      "stroke-width": selection?.kind === "segment" && selection.data?.segmentId === segment.segmentId ? 12 : 7,
      "stroke-linecap": "round",
      opacity: 0.9,
      "data-map-interactive": "true",
    });
    line.style.cursor = "pointer";
    line.addEventListener("click", () => onSelect({ kind: "segment", data: segment, meta: { load } }));
    layer.appendChild(line);
  }

  const activeTaskIds = new Set(
    tasks.filter((task) => ["assigned", "in_progress"].includes(task.taskStatus)).map((task) => task.taskId),
  );

  for (const task of tasks) {
    if (!activeTaskIds.has(task.taskId) || !task.plannedPath?.length) continue;
    const routePath = [];
    for (const pointId of task.plannedPath) {
      const point = pointIndex[pointId];
      if (!point) continue;
      routePath.push(`${routePath.length === 0 ? "M" : "L"} ${point.x} ${point.y}`);
    }
    const route = svgElement("path", {
      d: routePath.join(" "),
      stroke: selection?.kind === "task" && selection.data?.taskId === task.taskId ? "#bf4c39" : "#25624a",
      "stroke-width": selection?.kind === "task" && selection.data?.taskId === task.taskId ? 6 : 3,
      "stroke-dasharray": "10 8",
      fill: "none",
      opacity: 0.7,
    });
    layer.appendChild(route);
  }

  for (const point of mapData.points) {
    const circle = svgElement("circle", {
      cx: point.x,
      cy: point.y,
      r: selection?.kind === "point" && selection.data?.pointId === point.pointId ? 12 : 9,
      fill: "#1e2422",
      "data-map-interactive": "true",
    });
    circle.style.cursor = "pointer";
    circle.addEventListener("click", () => onSelect({ kind: "point", data: point }));
    layer.appendChild(circle);

    const label = svgElement("text", {
      x: point.x + 14,
      y: point.y - 12,
      "font-size": 18,
      "font-weight": 700,
      "data-map-interactive": "true",
    });
    label.textContent = point.pointId;
    label.style.cursor = "pointer";
    label.addEventListener("click", () => onSelect({ kind: "point", data: point }));
    layer.appendChild(label);
  }

  for (const robot of robots) {
    const point = pointIndex[robot.currentPoint];
    if (!point) continue;
    const robotNode = svgElement("rect", {
      x: point.x - 10,
      y: point.y + 12,
      width: 20,
      height: 20,
      rx: 5,
      fill: robot.state === "busy" ? "#bf4c39" : robot.state === "fault" ? "#c28b2f" : "#25624a",
      stroke: selection?.kind === "robot" && selection.data?.robotId === robot.robotId ? "#1e2422" : "white",
      "stroke-width": 2,
      "data-map-interactive": "true",
    });
    robotNode.style.cursor = "pointer";
    robotNode.addEventListener("click", () => onSelect({ kind: "robot", data: robot }));
    layer.appendChild(robotNode);
  }

  svg.appendChild(layer);
}

export function bindMapControls(root) {
  const shell = root.querySelector("#mapShell");
  const svg = root.querySelector("#mapCanvas");
  const zoomIn = root.querySelector("#mapZoomIn");
  const zoomOut = root.querySelector("#mapZoomOut");
  const reset = root.querySelector("#mapReset");
  if (!shell || !svg || shell.dataset.mapBound === "true") {
    return;
  }

  shell.dataset.mapBound = "true";
  zoomIn?.addEventListener("click", () => zoomMap(svg, 1 / ZOOM_STEP));
  zoomOut?.addEventListener("click", () => zoomMap(svg, ZOOM_STEP));
  reset?.addEventListener("click", () => resetMap(svg));
}

function computeBounds(points) {
  const padding = 96;
  const xs = points.map((point) => point.x);
  const ys = points.map((point) => point.y);
  const minX = Math.min(...xs) - padding;
  const maxX = Math.max(...xs) + padding;
  const minY = Math.min(...ys) - padding;
  const maxY = Math.max(...ys) + padding;
  return {
    minX,
    minY,
    width: Math.max(360, maxX - minX),
    height: Math.max(260, maxY - minY),
  };
}

function ensurePanZoom(svg, bounds) {
  if (!svg.__mapState) {
    svg.__mapState = {
      base: bounds,
      view: initialView(bounds),
      drag: null,
    };
    svg.addEventListener("wheel", (event) => {
      if (!svg.__mapState?.base) return;
      event.preventDefault();
      const factor = event.deltaY < 0 ? 1 / ZOOM_STEP : ZOOM_STEP;
      zoomMap(svg, factor, { x: event.offsetX, y: event.offsetY });
    }, { passive: false });
    svg.addEventListener("pointerdown", (event) => {
      if (event.button !== 0) return;
      if (event.target !== svg && event.target?.getAttribute?.("data-map-interactive") === "true") {
        return;
      }
      svg.setPointerCapture(event.pointerId);
      svg.__mapState.drag = {
        pointerId: event.pointerId,
        x: event.clientX,
        y: event.clientY,
        view: { ...svg.__mapState.view },
      };
      svg.classList.add("is-dragging");
    });
    svg.addEventListener("pointermove", (event) => {
      const drag = svg.__mapState?.drag;
      if (!drag || drag.pointerId !== event.pointerId) return;
      const rect = svg.getBoundingClientRect();
      if (!rect.width || !rect.height) return;
      const dx = event.clientX - drag.x;
      const dy = event.clientY - drag.y;
      const scaleX = drag.view.width / rect.width;
      const scaleY = drag.view.height / rect.height;
      svg.__mapState.view = clampView(svg.__mapState.base, {
        x: drag.view.x - dx * scaleX,
        y: drag.view.y - dy * scaleY,
        width: drag.view.width,
        height: drag.view.height,
      });
      applyViewBox(svg);
    });
    svg.addEventListener("pointerup", () => endDrag(svg));
    svg.addEventListener("pointercancel", () => endDrag(svg));
    svg.addEventListener("pointerleave", () => endDrag(svg));
  }

  svg.__mapState.base = bounds;
  const shouldReset =
    !svg.__mapState.view ||
    !Number.isFinite(svg.__mapState.view.x) ||
    !Number.isFinite(svg.__mapState.view.y) ||
    svg.__mapState.view.width === 0 ||
    svg.__mapState.view.height === 0;
  if (shouldReset) {
    svg.__mapState.view = initialView(bounds);
  } else {
    svg.__mapState.view = clampView(bounds, normalizeView(bounds, svg.__mapState.view));
  }
  applyViewBox(svg);
}

function zoomMap(svg, factor, anchor = null) {
  const state = svg.__mapState;
  if (!state?.base || !state?.view) return;

  const rect = svg.getBoundingClientRect();
  const baseScale = state.base.width / state.view.width;
  const nextScale = clamp(baseScale / factor, MIN_SCALE, MAX_SCALE);
  const nextWidth = state.base.width / nextScale;
  const nextHeight = state.base.height / nextScale;

  const anchorX = anchor?.x ?? rect.width / 2;
  const anchorY = anchor?.y ?? rect.height / 2;
  const rx = rect.width ? anchorX / rect.width : 0.5;
  const ry = rect.height ? anchorY / rect.height : 0.5;
  const anchorWorldX = state.view.x + state.view.width * rx;
  const anchorWorldY = state.view.y + state.view.height * ry;

  state.view = clampView(state.base, {
    x: anchorWorldX - nextWidth * rx,
    y: anchorWorldY - nextHeight * ry,
    width: nextWidth,
    height: nextHeight,
  });
  applyViewBox(svg);
}

function resetMap(svg) {
  const state = svg.__mapState;
  if (!state?.base) return;
  state.view = initialView(state.base);
  applyViewBox(svg);
}

function applyViewBox(svg) {
  const { view } = svg.__mapState;
  svg.setAttribute("viewBox", `${view.x} ${view.y} ${view.width} ${view.height}`);
}

function clampView(base, view) {
  const maxX = base.minX + base.width - view.width;
  const maxY = base.minY + base.height - view.height;
  return {
    x: clamp(view.x, base.minX, maxX),
    y: clamp(view.y, base.minY, maxY),
    width: clamp(view.width, base.width / MAX_SCALE, base.width / MIN_SCALE),
    height: clamp(view.height, base.height / MAX_SCALE, base.height / MIN_SCALE),
  };
}

function initialView(bounds) {
  return {
    x: bounds.minX,
    y: bounds.minY,
    width: bounds.width,
    height: bounds.height,
  };
}

function normalizeView(bounds, view) {
  return {
    x: Number.isFinite(view.x) ? view.x : bounds.minX,
    y: Number.isFinite(view.y) ? view.y : bounds.minY,
    width: Number.isFinite(view.width) ? view.width : bounds.width,
    height: Number.isFinite(view.height) ? view.height : bounds.height,
  };
}

function endDrag(svg) {
  if (!svg.__mapState) return;
  const pointerId = svg.__mapState.drag?.pointerId;
  if (pointerId != null && svg.hasPointerCapture?.(pointerId)) {
    svg.releasePointerCapture(pointerId);
  }
  svg.__mapState.drag = null;
  svg.classList.remove("is-dragging");
}

function svgElement(tagName, attributes) {
  const node = document.createElementNS(SVG_NS, tagName);
  for (const [key, value] of Object.entries(attributes)) {
    node.setAttribute(key, String(value));
  }
  return node;
}

function segmentStroke(load) {
  if (load >= 0.75) return "#bf4c39";
  if (load >= 0.4) return "#c28b2f";
  return "#7ea691";
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}
