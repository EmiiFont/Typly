package typly

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func newTestRenderer(t *testing.T, mutate ...func(*Config)) *Renderer {
	t.Helper()
	cfg := DefaultConfig()
	cfg.Width, cfg.Height = 320, 180
	cfg.FontSize = 24
	for _, m := range mutate {
		m(&cfg)
	}
	r, err := NewRenderer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	return r
}

// inkPixels counts pixels that are not the background palette index.
func inkPixels(img *image.RGBA) int {
	n := 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)) != (color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}) {
				n++
			}
		}
	}
	return n
}

func TestFrameBackgroundOnly(t *testing.T) {
	r := newTestRenderer(t)
	img := r.Frame("", false)
	if got := inkPixels(img); got != 0 {
		t.Errorf("empty frame without cursor has %d ink pixels, want 0", got)
	}
	b := img.Bounds()
	if b.Dx() != 320 || b.Dy() != 180 {
		t.Errorf("bounds = %v, want 320x180", b)
	}
}

func TestFrameCursorToggle(t *testing.T) {
	r := newTestRenderer(t)
	on := r.Frame("Hi", true)
	off := r.Frame("Hi", false)
	onInk := inkPixels(on)
	offInk := inkPixels(off)
	if offInk == 0 {
		t.Fatal("frame without cursor rendered no text at all")
	}
	if onInk <= offInk {
		t.Errorf("cursor-on ink (%d) should exceed cursor-off ink (%d)", onInk, offInk)
	}
}

func TestFrameCursorOnlyOnEmptyText(t *testing.T) {
	r := newTestRenderer(t)
	img := r.Frame("", true)
	if got := inkPixels(img); got == 0 {
		t.Error("empty text with cursor should render the cursor glyph")
	}
}

func TestWrapFitsWidth(t *testing.T) {
	r := newTestRenderer(t)
	text := "The quick brown fox jumps over the lazy dog again and again"
	lines := r.wrap(text)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping into multiple lines, got %d", len(lines))
	}
	if got := strings.Join(lines, " "); got != text {
		t.Errorf("wrap altered content:\n got %q\nwant %q", got, text)
	}
}

func TestWrapHardBreaksLongWord(t *testing.T) {
	r := newTestRenderer(t)
	long := strings.Repeat("W", 60)
	lines := r.wrap(long)
	if len(lines) < 2 {
		t.Fatalf("expected a hard break, got %d line(s)", len(lines))
	}
	if got := strings.Join(lines, ""); got != long {
		t.Errorf("hard break lost characters: got %d, want %d", len(got), len(long))
	}
}

func TestWrapHardBreaksUnicodeByRune(t *testing.T) {
	r := newTestRenderer(t)
	text := strings.Repeat("你好世界🌍", 12)
	lines := r.wrap(text)
	if len(lines) < 2 {
		t.Fatalf("expected a Unicode hard break, got %d line(s)", len(lines))
	}
	if got := strings.Join(lines, ""); got != text {
		t.Errorf("Unicode hard break altered text: got %q, want %q", got, text)
	}
}

func TestFrameRendersUnicodeAndEmoji(t *testing.T) {
	r := newTestRenderer(t)
	img := r.Frame("Hello, 世界 🌍 café", false)
	if got := inkPixels(img); got == 0 {
		t.Fatal("Unicode frame rendered no pixels")
	}
	// The symbol fallback must contribute visible pixels for the globe.
	emoji := r.Frame("🌍", false)
	if got := inkPixels(emoji); got == 0 {
		t.Fatal("emoji frame rendered no pixels")
	}
	runs := r.runs("🌍")
	if len(runs) != 1 || runs[0].emoji == nil {
		t.Fatalf("emoji was not routed to the color asset: %+v", runs)
	}
}

func TestWrapEmpty(t *testing.T) {
	r := newTestRenderer(t)
	if lines := r.wrap(""); len(lines) != 1 || lines[0] != "" {
		t.Errorf("wrap(\"\") = %q, want one empty line", lines)
	}
}

func TestNewRendererBadFont(t *testing.T) {
	cfg := DefaultConfig()
	cfg.FontData = []byte("not a font")
	if _, err := NewRenderer(cfg); err == nil {
		t.Error("expected error for invalid font data")
	}
}

func TestPaletteEndpoints(t *testing.T) {
	r := newTestRenderer(t)
	img := r.Frame("🌍", false)
	p := quantize(img).Palette
	if len(p) > 256 {
		t.Fatalf("palette size = %d, want at most 256", len(p))
	}
	br, bg, bb, _ := p[0].RGBA()
	if br>>8 != 0xff || bg>>8 != 0xff || bb>>8 != 0xff {
		t.Errorf("palette[0] = %v, want white background", p[0])
	}
	if inkPixels(img) == 0 {
		t.Error("color emoji frame has no visible pixels")
	}
}
