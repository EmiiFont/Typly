package typly

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/rivo/uniseg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// These fonts are embedded so common scripts and symbols render without
// requiring a font installation on the user's machine. Color emoji use the
// embedded Twemoji PNG assets instead of the monochrome symbol fallback.
//
//go:embed assets/NotoSans-Regular.ttf
var defaultFontData []byte

//go:embed assets/NotoSansSymbols2-Regular.ttf
var symbolFontData []byte

// paletteShades is the number of interpolated colors between background
// and foreground. Anti-aliased text only produces colors on that line,
// so a linear ramp reproduces it with no visible quantization.
const paletteShades = 64

// Renderer draws a single animation frame: wrapped text plus an
// optional trailing cursor, on a solid background.
type Renderer struct {
	cfg     Config
	faces   []font.Face
	fonts   []*sfnt.Font
	buffers []sfnt.Buffer
	margin  int
}

// NewRenderer parses the configured font (or the embedded Noto Sans
// default) and prepares a reusable renderer. The caller should Close it
// when done.
func NewRenderer(cfg Config) (*Renderer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	data := cfg.FontData
	if data == nil {
		data = defaultFontData
	}
	fontData := append([][]byte{data}, cfg.FallbackFontData...)
	fontData = append(fontData, symbolFontData)
	faces := make([]font.Face, 0, len(fontData))
	fonts := make([]*sfnt.Font, 0, len(fontData))
	buffers := make([]sfnt.Buffer, 0, len(fontData))
	for _, data := range fontData {
		parsed, err := parseFont(data)
		if err != nil {
			for _, f := range faces {
				_ = f.Close()
			}
			return nil, fmt.Errorf("typly: parse font: %w", err)
		}
		face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
			Size:    cfg.FontSize,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			for _, f := range faces {
				_ = f.Close()
			}
			return nil, fmt.Errorf("typly: create font face: %w", err)
		}
		faces = append(faces, face)
		fonts = append(fonts, parsed)
		buffers = append(buffers, sfnt.Buffer{})
	}
	return &Renderer{
		cfg:     cfg,
		faces:   faces,
		fonts:   fonts,
		buffers: buffers,
		margin:  int(cfg.FontSize / 2),
	}, nil
}

// parseFont accepts both individual fonts and TrueType collections (TTC),
// which are commonly used for CJK font families.
func parseFont(data []byte) (*sfnt.Font, error) {
	if parsed, err := sfnt.Parse(data); err == nil {
		return parsed, nil
	}
	collection, err := sfnt.ParseCollection(data)
	if err != nil {
		return nil, err
	}
	return collection.Font(0)
}

// Close releases font face resources.
func (r *Renderer) Close() error {
	for _, face := range r.faces {
		if err := face.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Config returns the validated configuration the renderer was built with.
func (r *Renderer) Config() Config { return r.cfg }

// Frame renders the given text with an optional trailing cursor and
// returns an RGBA image ready for PNG encoding or GIF quantization.
func (r *Renderer) Frame(text string, showCursor bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, r.cfg.Width, r.cfg.Height))
	draw.Draw(img, img.Bounds(), image.NewUniform(r.cfg.Background), image.Point{}, draw.Src)

	lines := r.wrap(text + r.cfg.Cursor)
	if !showCursor && len(lines) > 0 {
		last := strings.TrimSuffix(lines[len(lines)-1], r.cfg.Cursor)
		if last == "" {
			lines = lines[:len(lines)-1]
		} else {
			lines[len(lines)-1] = last
		}
	}

	metrics := r.faces[0].Metrics()
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(r.cfg.Foreground),
		Face: r.faces[0],
	}
	y := fixed.I(r.margin) + metrics.Ascent
	for _, line := range lines {
		x := fixed.I(r.margin)
		for _, run := range r.runs(line) {
			if run.emoji != nil {
				left := x.Floor()
				top := y.Floor() - int(r.cfg.FontSize)
				draw.Draw(img, image.Rect(left, top, left+run.emoji.Bounds().Dx(), top+run.emoji.Bounds().Dy()), run.emoji, image.Point{}, draw.Over)
				x += run.advance
				continue
			}
			drawer.Face = r.faces[run.faceIndex]
			drawer.Dot = fixed.Point26_6{X: x, Y: y}
			drawer.DrawString(string(run.runes))
			x += run.advance
		}
		y += metrics.Height
	}
	return img
}

// wrap greedily breaks text into lines that fit the canvas width minus
// margins. Words wider than a full line are hard-broken by Unicode rune.
func (r *Renderer) wrap(text string) []string {
	maxWidth := fixed.I(r.cfg.Width - 2*r.margin)
	var lines []string
	var current string
	flush := func() {
		if current != "" {
			lines = append(lines, current)
			current = ""
		}
	}
	for _, word := range strings.Fields(text) {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if r.measure(candidate) <= maxWidth {
			current = candidate
			continue
		}
		flush()
		// Hard-break words that do not fit on a line by themselves.
		wordRunes := []rune(word)
		for r.measure(string(wordRunes)) > maxWidth {
			cut := len(wordRunes)
			for cut > 1 && r.measure(string(wordRunes[:cut])) > maxWidth {
				cut--
			}
			lines = append(lines, string(wordRunes[:cut]))
			wordRunes = wordRunes[cut:]
		}
		current = string(wordRunes)
	}
	flush()
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

type textRun struct {
	faceIndex int
	runes     []rune
	advance   fixed.Int26_6
	emoji     image.Image
}

// runs selects the first font containing each rune. Keeping runs grouped
// avoids replacing unsupported Unicode with the primary font's tofu glyph.
func (r *Renderer) runs(text string) []textRun {
	var runs []textRun
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		cluster := graphemes.Str()
		if r.cfg.ColorEmoji {
			if emoji := emojiForCluster(cluster, int(r.cfg.FontSize)); emoji != nil {
				runs = append(runs, textRun{emoji: emoji, advance: fixed.I(int(r.cfg.FontSize))})
				continue
			}
		}
		for _, char := range graphemes.Runes() {
			faceIndex := 0
			for i := range r.faces {
				glyph, err := r.fonts[i].GlyphIndex(&r.buffers[i], char)
				if err == nil && glyph != 0 {
					faceIndex = i
					break
				}
			}
			advance := fixed.Int26_6(0)
			_, _, _, advance, _ = r.faces[faceIndex].Glyph(fixed.Point26_6{}, char)
			if len(runs) > 0 && runs[len(runs)-1].faceIndex == faceIndex {
				runs[len(runs)-1].runes = append(runs[len(runs)-1].runes, char)
				runs[len(runs)-1].advance += advance
			} else {
				runs = append(runs, textRun{faceIndex: faceIndex, runes: []rune{char}, advance: advance})
			}
		}
	}
	return runs
}

func (r *Renderer) measure(text string) fixed.Int26_6 {
	var width fixed.Int26_6
	for _, run := range r.runs(text) {
		width += run.advance
	}
	return width
}

// buildPalette returns a linear ramp from bg (index 0) to fg (last index).
func buildPalette(bg, fg color.Color) color.Palette {
	br, bgc, bb, _ := bg.RGBA()
	fr, fgc, fb, _ := fg.RGBA()
	p := make(color.Palette, paletteShades)
	for i := range p {
		t := uint32(i)
		n := uint32(paletteShades - 1)
		lerp := func(a, b uint32) uint8 {
			return uint8(((a*(n-t) + b*t) / n) >> 8)
		}
		p[i] = color.RGBA{
			R: lerp(br, fr),
			G: lerp(bgc, fgc),
			B: lerp(bb, fb),
			A: 0xff,
		}
	}
	return p
}
