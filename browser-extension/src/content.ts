// Content script — injected into the target page. Handles DOM-side capture
// ops: fixed-element fixing, scrolling step, exact restoration.

import type { PageMeta } from "./shared";

type Msg = { cmd: "viewportInfo" | "fixRegion" | "step" | "restore" };

// Exact original inline styles of every mutated element, keyed by node.
const savedStyles = new Map<HTMLElement, string>();
let savedScroll: { x: number; y: number } | null = null;

function doc(): HTMLElement {
  return document.documentElement;
}

function viewportInfo(): PageMeta {
  return {
    title: document.title,
    viewportWidth: doc().clientWidth,
    viewportHeight: doc().clientHeight,
    totalHeight: Math.max(doc().scrollHeight, document.body.scrollHeight),
    dpr: window.devicePixelRatio,
    scrollbarOverlap: window.innerWidth - doc().clientWidth,
  };
}

function isVisible(node: HTMLElement): boolean {
  const cs = getComputedStyle(node);
  return cs.display !== "none" && cs.visibility !== "hidden";
}

// Rewrite fixed/sticky elements to absolute so they don't repeat in every
// frame, save their exact inline styles, then scroll to the page top.
function fixRegion(): void {
  savedScroll = { x: window.scrollX, y: window.scrollY };
  for (const node of Array.from(document.querySelectorAll<HTMLElement>("*"))) {
    if (!isVisible(node)) continue;
    const cs = getComputedStyle(node);
    if (cs.position === "fixed" || cs.position === "sticky") {
      savedStyles.set(node, node.style.cssText);
      node.style.setProperty("position", "absolute", "important");
    }
  }
  window.scrollTo(0, 0);
}

function restore(): void {
  for (const [node, cssText] of savedStyles) {
    node.style.cssText = cssText;
  }
  savedStyles.clear();
  if (savedScroll) {
    window.scrollTo(savedScroll.x, savedScroll.y);
    savedScroll = null;
  }
}

chrome.runtime.onMessage.addListener((msg: Msg, _sender, respond) => {
  let reply: PageMeta | { kind: string; y: number };
  switch (msg.cmd) {
    case "viewportInfo":
      reply = viewportInfo();
      break;
    case "fixRegion":
      fixRegion();
      reply = { kind: "stepDone", y: window.scrollY };
      break;
    case "step":
      window.scrollTo(0, window.scrollY + doc().clientHeight);
      reply = { kind: "stepDone", y: window.scrollY };
      break;
    case "restore":
      restore();
      reply = { kind: "stepDone", y: window.scrollY };
      break;
  }
  respond(reply);
  return true;
});
