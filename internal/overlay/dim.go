package overlay

import "image"

// DimImage returns a copy of src with every pixel darkened by the given
// factor. A factor of 0 produces a fully black image; 255 leaves the
// image unchanged. A typical overlay uses a factor around 128 (50 %).
//
// DimImage creates a new image that is a copy of src with its red, green, and blue channels scaled by factor (0–255) while preserving each pixel's alpha channel.
// The scaling is applied as channel * factor / 255; factor 255 leaves colors unchanged and factor 0 makes RGB values 0 while keeping alpha intact.
func DimImage(src *image.RGBA, factor uint8) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	f := uint16(factor)
	for i := 0; i < len(src.Pix); i += 4 {
		dst.Pix[i+0] = uint8(uint16(src.Pix[i+0]) * f / 255) // R
		dst.Pix[i+1] = uint8(uint16(src.Pix[i+1]) * f / 255) // G
		dst.Pix[i+2] = uint8(uint16(src.Pix[i+2]) * f / 255) // B
		dst.Pix[i+3] = src.Pix[i+3]                           // A
	}
	return dst
}

// ComposeSelection composites the original (un-dimmed) pixels from src
// into dst within the given selection rectangle, producing the
// "bright window" effect typical of screenshot overlay tools.
//
// Pixels outside the selection remain as they are in dst (typically a
// dimmed version of src). The selection rect is clamped to the image
// ComposeSelection copies the rectangular region sel from src into dst, constrained to the overlap of their bounds.
// 
// The selection rectangle is first intersected with src.Bounds() and dst.Bounds(); if the resulting rectangle is empty
// the function returns without modifying dst. For a non-empty selection, the corresponding RGBA pixels from src
// overwrite dst inside the clamped rectangle.
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
