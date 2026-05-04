const fragmentCache = new Map();

export function qs(selector, root = document) {
  return root.querySelector(selector);
}

export function clearNode(node) {
  while (node.firstChild) {
    node.removeChild(node.firstChild);
  }
}

export function el(tagName, attributes = {}, children = []) {
  const node = document.createElement(tagName);
  for (const [key, value] of Object.entries(attributes)) {
    if (value == null) continue;
    if (key === "className") {
      node.className = value;
      continue;
    }
    if (key === "dataset") {
      Object.assign(node.dataset, value);
      continue;
    }
    if (key.startsWith("on") && typeof value === "function") {
      node.addEventListener(key.slice(2).toLowerCase(), value);
      continue;
    }
    node.setAttribute(key, String(value));
  }

  const list = Array.isArray(children) ? children : [children];
  for (const child of list) {
    if (child == null) continue;
    if (typeof child === "string") {
      node.appendChild(document.createTextNode(child));
      continue;
    }
    node.appendChild(child);
  }
  return node;
}

export async function loadFragment(name) {
  if (!fragmentCache.has(name)) {
    const response = await fetch(`/ui/fragments/${name}.html`);
    if (!response.ok) {
      throw new Error(`Unable to load fragment ${name}`);
    }
    fragmentCache.set(name, await response.text());
  }
  return fragmentCache.get(name);
}

export async function mountFragment(container, name) {
  container.innerHTML = await loadFragment(name);
  return container;
}
