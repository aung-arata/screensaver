// Content script — injected into the target page. Handles DOM-side capture
// ops: fixed-element fixing, scrolling step, restoration.

import type { PageMeta } from "./shared";

const PersistAttr = "data-ssw-persist";

type Msg = { cmd: "viewportInfo" | "fixRegion" | "step" | "restore" };

function doc(): HTMLElement {
  return document.documentElement;
}

function viewportInfo(): PageMeta {
  return {
    title: document.title,
    viewportWidth: doc().clientWidth,
    totalHeight: Math.max(doc().scrollHeight, document.body.scrollHeight),
    dpr: window.devicePixelRatio,
    scrollbarOverlap: window.innerWidth - doc().clientWidth,
  };
}

function isVisible(node: HTMLElement): boolean {
  const cs = getComputedStyle(node);
  return cs.display !== "none" && cs.visibility !== "hidden";
}

function fixRegion(): void {
  for (const node of Array.from(document.querySelectorAll<HTMLElement>("*"))) {
    if (!isVisible(node)) continue;
    const cs = getComputedStyle(node);
    if (cs.position === "fixed" || cs.position === "sticky") {
      node.setAttribute(PersistAttr, JSON.stringify({ position: cs.position }));
      node.style.setProperty("position", "absolute", "important");
    }
  }
}

function restore(): void {
  for (const node of Array.from(document.querySelectorAll(`[${PersistAttr}]`))) {
    const originals = JSON.parse(node.getAttribute(PersistAttr) || "{}") as Record<string, string>;
    node.removeAttribute(PersistAttr);
    for (const k of Object.keys(originals)) {
      const v = originals[k];
      if (v === undefined) {
        (node as HTMLElement).style.removeProperty(k);
      } else {
        (node as HTMLElement).style.setProperty(k, v, "important");
      }
    }
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
