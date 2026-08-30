// Background service worker — orchestrates full-page capture and serves
// results to viewer pages via persistent ports.

import type { CaptureJob, Frame, PageMeta, StepDone, ViewportInfo } from "./shared";
import { MAX_CAPTURED_FRAMES } from "./shared";

const jobs = new Map<string, CaptureJob>();

// Chrome limits captureVisibleTab to 2 calls/sec. 600ms between captures
// stays under the quota and doubles as scroll/render settle time.
const CAPTURE_INTERVAL_MS = 600;
const SETTLE_MS = 250;
// Conservative canvas height guard (Chrome canvas caps near 16384px/side).
const MAX_CANVAS_PX = 16000;

chrome.action.onClicked.addListener((tab) => void startCapture(tab));

chrome.runtime.onConnect.addListener((port) => {
  if (port.name !== "viewer") return;
  port.onMessage.addListener((msg: { wanted: string }) => {
    const job = jobs.get(msg.wanted);
    if (job) port.postMessage(job);
  });
});

async function startCapture(tab: chrome.tabs.Tab): Promise<void> {
  const jobId = String(Date.now());

  if (isBlocked(tab.url)) {
    jobs.set(jobId, {
      meta: { complete: false, error: "This page can't be captured (restricted URL)." },
      frames: [],
      info: emptyInfo(),
    });
    chrome.tabs.create({ url: `viewer.html#${jobId}` });
    return;
  }

  try {
    jobs.set(jobId, await capture(tab));
  } catch (e) {
    console.error("capture error", e);
    jobs.set(jobId, {
      meta: { complete: false, error: "Capture failed: " + String(e) },
      frames: [],
      info: emptyInfo(),
    });
  }
  chrome.tabs.create({ url: `viewer.html#${jobId}` });
}

async function capture(tab: chrome.tabs.Tab): Promise<CaptureJob> {
  const tabId = tab.id!;
  await chrome.scripting.executeScript({ target: { tabId }, files: ["content.js"] });
  await delay(50); // let the injected listener register

  const info = (await command(tabId, "viewportInfo")) as ViewportInfo;
  const maxFrames = Math.max(
    1,
    Math.min(MAX_CAPTURED_FRAMES, Math.floor(MAX_CANVAS_PX / (info.viewportHeight * info.dpr)))
  );

  const frames: Frame[] = [];
  try {
    await command(tabId, "fixRegion"); // also scrolls to page top
    await delay(SETTLE_MS); // re-layout after rewriting fixed elements

    // Capture the first viewport BEFORE any scrolling (fixRegion scrolled to top).
    let scrollY = 0;
    frames.push({ scrollY, dataUrl: await snap(tab) });

    // Step until the scroll position stops advancing (page bottom).
    while (frames.length < maxFrames) {
      const step = (await command(tabId, "step")) as StepDone;
      if (step.y <= scrollY) break;
      scrollY = step.y;
      await delay(CAPTURE_INTERVAL_MS);
      frames.push({ scrollY, dataUrl: await snap(tab) });
    }
  } finally {
    // Always restore: original inline styles + original scroll position.
    await command(tabId, "restore").catch(() => void 0);
  }

  return { meta: { complete: true }, frames, info };
}

async function snap(tab: chrome.tabs.Tab): Promise<string> {
  return chrome.tabs.captureVisibleTab(tab.windowId, { format: "png" });
}

type Cmd = "viewportInfo" | "fixRegion" | "step" | "restore";

async function command(tabId: number, cmd: Cmd): Promise<ViewportInfo | StepDone> {
  return (await chrome.tabs.sendMessage(tabId, { cmd })) as ViewportInfo | StepDone;
}

function delay(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function isBlocked(url?: string): boolean {
  return (
    !url ||
    url.startsWith("chrome://") ||
    url.startsWith("chrome-extension://") ||
    url.includes("chromewebstore.google.com")
  );
}

function emptyInfo(): PageMeta {
  return {
    title: "",
    viewportWidth: 0,
    viewportHeight: 0,
    totalHeight: 0,
    dpr: 1,
    scrollbarOverlap: 0,
  };
}
