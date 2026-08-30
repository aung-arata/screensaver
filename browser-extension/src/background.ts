// Background service worker — orchestrates capture, holding frames in memory
// to be served to viewer tabs via persistent ports.

import type { CaptureJob, Frame, PageMeta, StepDone, ViewportInfo } from "./shared";
import { MAX_CAPTURED_FRAMES } from "./shared";

const jobs = new Map<string, CaptureJob>();

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
    await chrome.scripting.executeScript({ target: { tabId: tab.id! }, files: ["content.js"] });
    // small settle — chrome.scripting resolves before listener is registered in some edge cases
    await new Promise((r) => setTimeout(r, 50));
    const info = (await command(tab.id!, "viewportInfo")) as ViewportInfo;
    await command(tab.id!, "fixRegion");

    const frames: Frame[] = [];
    const viewportPx = info.viewportWidth;
    void viewportPx;

    let y = 0;
    while (y < info.totalHeight && frames.length < MAX_CAPTURED_FRAMES) {
      const step = (await command(tab.id!, "step")) as StepDone;
      const shot = await chrome.tabs.captureVisibleTab(tab.windowId, { format: "png" });
      frames.push({ scrollY: step.y, dataUrl: shot });
      y = step.y;
    }

    await command(tab.id!, "restore");

    jobs.set(jobId, { meta: { complete: true }, frames, info });
    chrome.tabs.create({ url: `viewer.html#${jobId}` });
  } catch (e) {
    console.error("capture error", e);
    jobs.set(jobId, {
      meta: { complete: false, error: "Capture failed: " + String(e) },
      frames: [],
      info: emptyInfo(),
    });
    chrome.tabs.create({ url: `viewer.html#${jobId}` });
  }
}

type Cmd = "viewportInfo" | "fixRegion" | "step" | "restore";

async function command(tabId: number, cmd: Cmd): Promise<ViewportInfo | StepDone | PageMeta> {
  return (await chrome.tabs.sendMessage(tabId, { cmd })) as ViewportInfo | StepDone | PageMeta;
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
  return { title: "", viewportWidth: 0, totalHeight: 0, dpr: 1, scrollbarOverlap: 0 };
}
