// Shared types between background, content, and viewer.

export interface Frame {
  scrollY: number;
  dataUrl: string;
}

export interface PageMeta {
  title: string;
  viewportWidth: number;
  viewportHeight: number;
  totalHeight: number;
  dpr: number;
  scrollbarOverlap: number;
}

export type ViewportInfo = PageMeta & { kind: "viewportInfo" };
export type StepDone = { kind: "stepDone"; y: number };

export type CaptureReply = ViewportInfo | StepDone;

export interface CaptureJob {
  meta: {
    complete: boolean;
    error?: string;
  };
  frames: Frame[];
  info: PageMeta;
}

export const MAX_CAPTURED_FRAMES = 100;
