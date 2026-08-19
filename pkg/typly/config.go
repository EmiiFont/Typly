// Package typly generates typing-animation frames and encodes them as
// animated GIFs or MP4 videos.
package typly

import (
	"fmt"
	"image/color"
	"math"
	"strings"
)

// Config controls rendering and animation behavior.
type Config struct {
	Width            int         // canvas width in pixels
	Height           int         // canvas height in pixels
	FontSize         float64     // font size in points
	FontData         []byte      // primary TTF/OTF font; nil means embedded Noto Sans
	FallbackFontData [][]byte    // optional fallback fonts, tried in order
	Foreground       color.Color // text and cursor color
	Background       color.Color // canvas background color
	Cursor           string      // cursor glyph, e.g. "|"
	ColorEmoji       bool        // use embedded Twemoji assets when available
	Blinks           int         // cursor blink cycles at the end of each sentence
	FPS              int         // playback frame rate
}

// DefaultConfig mirrors the defaults of the original Typly:
// 1280x720, 72pt, black on white, "|" cursor, 12 fps.
func DefaultConfig() Config {
	return Config{
		Width:      1280,
		Height:     720,
		FontSize:   72,
		Foreground: color.Black,
		Background: color.White,
		Cursor:     "|",
		ColorEmoji: true,
		Blinks:     3,
		FPS:        12,
	}
}

// Validate checks the configuration and normalizes dimensions to even
// values, which video encoders such as libx264 require.
func (c *Config) Validate() error {
	if c.Width <= 0 || c.Height <= 0 {
		return fmt.Errorf("typly: invalid canvas size %dx%d", c.Width, c.Height)
	}
	if c.FontSize <= 0 || math.IsNaN(c.FontSize) || math.IsInf(c.FontSize, 0) {
		return fmt.Errorf("typly: invalid font size %v", c.FontSize)
	}
	if c.FPS <= 0 {
		return fmt.Errorf("typly: invalid fps %d", c.FPS)
	}
	if c.Blinks < 0 {
		return fmt.Errorf("typly: invalid blink count %d", c.Blinks)
	}
	if strings.TrimSpace(c.Cursor) == "" {
		c.Cursor = "|"
	}
	c.Width = evenDimension(c.Width)
	c.Height = evenDimension(c.Height)
	return nil
}

// evenDimension rounds n up to the nearest even number.
func evenDimension(n int) int {
	if n%2 != 0 {
		return n + 1
	}
	return n
}

// ParseHexColor parses colors of the form "#RGB", "#RRGGBB", with or
// without the leading '#'.
func ParseHexColor(s string) (color.Color, error) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	var r, g, b uint8
	switch len(s) {
	case 3:
		if _, err := fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b); err != nil {
			return nil, fmt.Errorf("typly: invalid hex color %q", s)
		}
		r, g, b = r*17, g*17, b*17
	case 6:
		if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
			return nil, fmt.Errorf("typly: invalid hex color %q", s)
		}
	default:
		return nil, fmt.Errorf("typly: invalid hex color %q", s)
	}
	return color.RGBA{R: r, G: g, B: b, A: 0xff}, nil
}
