// Deterministic stitcher — uses actual DOM scroll offsets, no image overlap
// detection needed. Canvas is scaled by devicePixelRatio.

import type { Frame, PageMeta } from "./shared";

export async function stitch(frames: Frame[], info: PageMeta): Promise<HTMLCanvasElement> {
  if (frames.length === 0) throw new Error("no frames captured");

  const sorted = [...frames].sort((a, b) => a.scrollY - b.scrollY);
  const minY = sorted[0].scrollY;
  const maxY = sorted[sorted.length - 1].scrollY + info.totalHeight; // top of last + remaining tail

  const canvas = document.createElement("canvas");
  canvas.width = info.viewportWidth * info.dpr;
  canvas.height = maxY * info.dpr;
  const ctx = canvas.getContext("2d")!;
  ctx.imageSmoothingEnabled = false;

  for (const frame of sorted) {
    const img = await toImage(frame.dataUrl);
    // Crop scrollbar area on right; start x from scrollbarOverlap to skip the scrollbar
    const sx = info.scrollbarOverlap * info.dpr;
    const sw = img.width - info.scrollbarOverlap * info.dpr;
    const dx = 0;
    const dy = (frame.scrollY - minY) * info.dpr;
    ctx.drawImage(img, sx, 0, sw, img.height, dx, dy, sw, img.height);
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
