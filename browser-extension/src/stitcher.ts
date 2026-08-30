// Deterministic stitcher — uses actual DOM scroll offsets, no image overlap
// detection needed. Canvas is scaled by devicePixelRatio.

import type { Frame, PageMeta } from "./shared";

export async function stitch(frames: Frame[], info: PageMeta): Promise<HTMLCanvasElement> {
  if (frames.length === 0) throw new Error("no frames captured");

  const sorted = [...frames].sort((a, b) => a.scrollY - b.scrollY);
  const minY = sorted[0].scrollY;
  const last = sorted[sorted.length - 1];

  // Real captured height: last frame's offset + one viewport. This is the
  // truth even if the page grew from lazy loading mid-capture.
  const cssHeight = last.scrollY + info.viewportHeight - minY;

  const canvas = document.createElement("canvas");
  canvas.width = Math.round(info.viewportWidth * info.dpr);
  canvas.height = Math.round(cssHeight * info.dpr);
  const ctx = canvas.getContext("2d")!;

  for (const frame of sorted) {
    const img = await toImage(frame.dataUrl);
    // Crop the right-side scrollbar: source starts at x=0 with reduced width.
    const sw = img.width - info.scrollbarOverlap * info.dpr;
    const dy = (frame.scrollY - minY) * info.dpr;
    ctx.drawImage(img, 0, 0, sw, img.height, 0, dy, sw, img.height);
  }
  return canvas;
}

async function toImage(dataUrl: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error("image decode failed"));
    img.src = dataUrl;
  });
}
