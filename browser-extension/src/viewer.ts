// Viewer page — receives frames from background service worker, stitches via
// canvas, and provides PNG/JPG/PDF save + clipboard copy.

import type { CaptureJob } from "./shared";
import { jsPDF } from "jspdf";
import { stitch } from "./stitcher";

const jobId = location.hash.slice(1);
const port = chrome.runtime.connect({ name: "viewer" });
port.postMessage({ wanted: jobId });

port.onMessage.addListener((job: CaptureJob) => {
  if (!job.meta.complete) return handleError(job.meta.error);
  void render(job);
});

function handleError(msg?: string): void {
  document.getElementById("error")!.textContent = msg ?? "Capture failed.";
}

async function render(job: CaptureJob): Promise<void> {
  const canvas = await stitch(job.frames, job.info);
  const view = document.getElementById("c") as HTMLCanvasElement;
  view.width = canvas.width;
  view.height = canvas.height;
  view.getContext("2d")!.drawImage(canvas, 0, 0);

  document.getElementById("title")!.textContent = job.info.title || "Captured page";

  const png = canvas.toDataURL("image/png");
  const jpg = canvas.toDataURL("image/jpeg", 0.92);
  const base = fileBase(job.info.title || "page");

  document.getElementById("savePng")!.addEventListener("click", () => {
    chrome.downloads.download({ url: png, filename: `${base}.png` });
  });
  document.getElementById("saveJpg")!.addEventListener("click", () => {
    chrome.downloads.download({ url: jpg, filename: `${base}.jpg` });
  });
  document.getElementById("savePdf")!.addEventListener("click", () => {
    exportPDF(canvas, `${base}.pdf`);
  });
  document.getElementById("copy")!.addEventListener("click", () => copy(canvas));
}

function exportPDF(canvas: HTMLCanvasElement, name: string): void {
  const doc = new jsPDF({
    orientation: canvas.width < canvas.height ? "portrait" : "landscape",
    unit: "pt",
    format: [canvas.width, canvas.height],
  });
  doc.addImage(canvas.toDataURL("image/jpeg", 0.95), "JPEG", 0, 0, canvas.width, canvas.height);
  doc.save(name);
}

async function copy(canvas: HTMLCanvasElement): Promise<void> {
  const blob = await new Promise<Blob | null>((r) => canvas.toBlob(r, "image/png"));
  if (!blob) return;
  await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
}

function fileBase(t: string): string {
  return t.replace(/[^\w.-]+/g, "_").slice(0, 60);
}
