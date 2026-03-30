package overlay

import "image"

// DimImage returns a copy of src with every pixel's RGB channels scaled by
// factor (0–255) while preserving each pixel's alpha channel. A factor of 0
// produces a fully black image; 255 leaves the image unchanged. A typical
// overlay uses a factor around 128 (50%).
func DimImage(src *image.RGBA, factor uint8) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	f := uint16(factor)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			off := src.PixOffset(x, y)
			dOff := dst.PixOffset(x, y)
			dst.Pix[dOff+0] = uint8(uint16(src.Pix[off+0]) * f / 255) // R
			dst.Pix[dOff+1] = uint8(uint16(src.Pix[off+1]) * f / 255) // G
			dst.Pix[dOff+2] = uint8(uint16(src.Pix[off+2]) * f / 255) // B
			dst.Pix[dOff+3] = src.Pix[off+3]                           // A
		}
	}
	return dst
}

// ComposeSelection composites the original (un-dimmed) pixels from src into
// dst within the given selection rectangle, producing the "bright window"
// effect typical of screenshot overlay tools. Pixels outside the selection
// remain as they are in dst (typically a dimmed version of src). The selection
// rectangle is clamped to both src and dst bounds; if the overlap is empty the
// function returns without modifying dst.
func ComposeSelection(dst, src *image.RGBA, sel image.Rectangle) {
	sel = sel.Intersect(src.Bounds()).Intersect(dst.Bounds())
	if sel.Empty() {
		return
	}
	for y := sel.Min.Y; y < sel.Max.Y; y++ {
		srcOff := src.PixOffset(sel.Min.X, y)
		dstOff := dst.PixOffset(sel.Min.X, y)
		copy(dst.Pix[dstOff:dstOff+sel.Dx()*4], src.Pix[srcOff:srcOff+sel.Dx()*4])
	}
}
