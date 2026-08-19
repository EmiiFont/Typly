package typly

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"io"
)

// EncodeGIF writes the animation as an infinitely looping animated GIF.
func EncodeGIF(w io.Writer, frames []Frame) error {
	if len(frames) == 0 {
		return fmt.Errorf("typly: no frames to encode")
	}
	g := &gif.GIF{
		Image:     make([]*image.Paletted, len(frames)),
		Delay:     make([]int, len(frames)),
		LoopCount: 0, // loop forever
	}
	for i, f := range frames {
		g.Image[i] = quantize(f.Image)
		g.Delay[i] = f.Delay
	}
	if err := gif.EncodeAll(w, g); err != nil {
		return fmt.Errorf("typly: encode gif: %w", err)
	}
	return nil
}

// quantize converts a full-color frame to a compact GIF palette. The first
// two entries preserve the configured background and text colors; Plan9
// supplies enough colors for Twemoji artwork while staying within GIF's
// 256-color limit.
func quantize(src *image.RGBA) *image.Paletted {
	p := make(color.Palette, 0, 256)
	p = append(p, src.At(0, 0), color.Black)
	for _, c := range palette.Plan9 {
		if len(p) == 256 {
			break
		}
		p = append(p, c)
	}
	dst := image.NewPaletted(src.Bounds(), p)
	cache := make([]int16, 1<<15)
	for i := range cache {
		cache[i] = -1
	}
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			r, g, b, _ := src.At(x, y).RGBA()
			key := int((r>>11)<<10 | (g>>11)<<5 | b>>11)
			idx := cache[key]
			if idx < 0 {
				idx = int16(nearestColor(p, uint8(r>>8), uint8(g>>8), uint8(b>>8)))
				cache[key] = idx
			}
			dst.SetColorIndex(x, y, uint8(idx))
		}
	}
	return dst
}

func nearestColor(p color.Palette, r, g, b uint8) int {
	best, bestDistance := 0, int64(1<<62)
	for i, candidate := range p {
		cr, cg, cb, _ := candidate.RGBA()
		dr, dg, db := int64(r)-int64(cr>>8), int64(g)-int64(cg>>8), int64(b)-int64(cb>>8)
		distance := dr*dr + dg*dg + db*db
		if distance < bestDistance {
			best, bestDistance = i, distance
		}
	}
	return best
}
