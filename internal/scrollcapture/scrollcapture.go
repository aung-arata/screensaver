package scrollcapture

import (
	"fmt"
	"image"
	"image/draw"
)

const (
	defaultDelayMs   = 250
	defaultStepPx    = 900
	defaultMaxFrames = 20
	minNewContentPx  = 30
	minOverlapHeight = 20
	overlapEdgeGuard = 10
	frameDiffStepX   = 8
	frameDiffStepY   = 8
	overlapStepX     = 6
	overlapStepY     = 2

	// nearIdenticalThreshold is the average absolute sampled RGB-channel
	// difference (0..255 scale) below which two frames are treated as identical.
	nearIdenticalThreshold = 2.0
	// maxOverlapScore is the maximum average absolute sampled RGB-channel
	// difference (0..255 scale) accepted for a valid overlap match.
	maxOverlapScore = 20.0
)

// Config controls long-page auto-scroll capture behavior.
type Config struct {
	DelayMs   int
	StepPx    int
	MaxFrames int
}

// Capture performs a Windows-only long-page capture for the given selected
// region. It auto-scrolls between frames and returns the final vertically
// stitched image.
func Capture(region image.Rectangle, cfg Config) (image.Image, error) {
	if region.Dx() <= 0 || region.Dy() <= 0 {
		return nil, fmt.Errorf("invalid region: %v", region)
	}
	return capturePlatform(region, normalizeConfig(cfg))
}

func normalizeConfig(cfg Config) Config {
	if cfg.DelayMs <= 0 {
		cfg.DelayMs = defaultDelayMs
	}
	if cfg.StepPx == 0 {
		cfg.StepPx = defaultStepPx
	}
	if cfg.MaxFrames <= 0 {
		cfg.MaxFrames = defaultMaxFrames
	}
	return cfg
}

func stitchFrames(frames []*image.RGBA) *image.RGBA {
	if len(frames) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	width := frames[0].Bounds().Dx()
	totalHeight := 0
	for _, f := range frames {
		totalHeight += f.Bounds().Dy()
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, totalHeight))
	y := 0
	for _, f := range frames {
		h := f.Bounds().Dy()
		draw.Draw(dst, image.Rect(0, y, width, y+h), f, f.Bounds().Min, draw.Src)
		y += h
	}
	return dst
}

func cropRows(src *image.RGBA, yStart int) *image.RGBA {
	b := src.Bounds()
	h := b.Dy()
	if yStart <= 0 {
		return cloneRGBA(src)
	}
	if yStart >= h {
		return image.NewRGBA(image.Rect(0, 0, b.Dx(), 0))
	}

	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), h-yStart))
	draw.Draw(
		dst,
		dst.Bounds(),
		src,
		image.Pt(b.Min.X, b.Min.Y+yStart),
		draw.Src,
	)
	return dst
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	return dst
}

func framesAreNearIdentical(a, b *image.RGBA) bool {
	diff, ok := sampleAverageDiff(a, b, frameDiffStepX, frameDiffStepY, 0, a.Bounds().Dy(), 0)
	if !ok {
		return false
	}
	return diff <= nearIdenticalThreshold
}

func findBestVerticalOverlap(prev, curr *image.RGBA) (int, bool) {
	bp := prev.Bounds()
	bc := curr.Bounds()
	if bp.Dx() != bc.Dx() || bp.Dy() != bc.Dy() {
		return 0, false
	}
	h := bp.Dy()
	if h < minOverlapHeight {
		return 0, false
	}

	minOverlap := 50
	if h/4 < minOverlap {
		minOverlap = h / 4
	}
	if minOverlap < 10 {
		minOverlap = 10
	}
	maxOverlap := h - overlapEdgeGuard
	if maxOverlap <= minOverlap {
		return 0, false
	}

	bestOverlap := 0
	bestScore := 1e9
	for overlap := minOverlap; overlap <= maxOverlap; overlap++ {
		startPrev := h - overlap
		score, ok := sampleAverageDiff(prev, curr, overlapStepX, overlapStepY, startPrev, h, 0)
		if !ok {
			continue
		}
		if score < bestScore {
			bestScore = score
			bestOverlap = overlap
		}
	}

	if bestOverlap == 0 {
		return 0, false
	}
	if bestScore > maxOverlapScore {
		return 0, false
	}
	return bestOverlap, true
}

func sampleAverageDiff(a, b *image.RGBA, stepX, stepY, aYStart, aYEnd, bYStart int) (float64, bool) {
	ba := a.Bounds()
	bb := b.Bounds()
	if ba.Dx() != bb.Dx() || ba.Dy() != bb.Dy() {
		return 0, false
	}
	if stepX <= 0 {
		stepX = 1
	}
	if stepY <= 0 {
		stepY = 1
	}

	h := ba.Dy()
	w := ba.Dx()
	if aYStart < 0 {
		aYStart = 0
	}
	if aYEnd > h {
		aYEnd = h
	}
	if aYEnd <= aYStart || w == 0 {
		return 0, false
	}
	if bYStart < 0 || bYStart+(aYEnd-aYStart) > h {
		return 0, false
	}

	total := int64(0)
	samples := int64(0)
	for ayIdx := aYStart; ayIdx < aYEnd; ayIdx += stepY {
		byIdx := bYStart + (ayIdx - aYStart)
		ay := ba.Min.Y + ayIdx
		by := bb.Min.Y + byIdx
		for x := 0; x < w; x += stepX {
			ax := ba.Min.X + x
			bx := bb.Min.X + x
			ia := a.PixOffset(ax, ay)
			ib := b.PixOffset(bx, by)
			total += int64(absDiff(a.Pix[ia], b.Pix[ib]))
			total += int64(absDiff(a.Pix[ia+1], b.Pix[ib+1]))
			total += int64(absDiff(a.Pix[ia+2], b.Pix[ib+2]))
			samples += 3
		}
	}
	if samples == 0 {
		return 0, false
	}
	return float64(total) / float64(samples), true
}

func absDiff(a, b byte) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
