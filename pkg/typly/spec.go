package typly

import (
	"fmt"
	"strings"
)

const (
	maxTextBytes   = 64 * 1024
	maxWidth       = 3840
	maxHeight      = 2160
	maxFrames      = 5000
	maxCursorBytes = 16
)

// RenderSpec is the JSON contract shared by the CLI, web app, mobile app,
// and server-side render jobs. It intentionally contains no filesystem or
// font bytes, so it is safe to persist and send over an API.
type RenderSpec struct {
	Sentences  []string `json:"sentences"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	FontSize   float64  `json:"fontSize"`
	FPS        int      `json:"fps"`
	Foreground string   `json:"foreground"`
	Background string   `json:"background"`
	Cursor     string   `json:"cursor"`
	Blinks     int      `json:"blinks"`
	Emoji      string   `json:"emoji"` // color or mono
}

// DefaultRenderSpec returns the product-level defaults.
func DefaultRenderSpec() RenderSpec {
	cfg := DefaultConfig()
	return RenderSpec{
		Width:      cfg.Width,
		Height:     cfg.Height,
		FontSize:   cfg.FontSize,
		FPS:        cfg.FPS,
		Foreground: "#000000",
		Background: "#FFFFFF",
		Cursor:     cfg.Cursor,
		Blinks:     cfg.Blinks,
		Emoji:      "color",
	}
}

// Validate checks user-controlled render input and applies omitted defaults.
// The limits protect the web server from accidental or malicious oversized
// jobs; callers should validate before allocating frames.
func (s *RenderSpec) Validate() error {
	defaults := DefaultRenderSpec()
	if s.Width == 0 {
		s.Width = defaults.Width
	}
	if s.Height == 0 {
		s.Height = defaults.Height
	}
	if s.FontSize == 0 {
		s.FontSize = defaults.FontSize
	}
	if s.FPS == 0 {
		s.FPS = defaults.FPS
	}
	if s.Foreground == "" {
		s.Foreground = defaults.Foreground
	}
	if s.Background == "" {
		s.Background = defaults.Background
	}
	if s.Cursor == "" {
		s.Cursor = defaults.Cursor
	}
	if s.Emoji == "" {
		s.Emoji = defaults.Emoji
	}

	textBytes := 0
	for _, sentence := range s.Sentences {
		textBytes += len(sentence)
	}
	if len(s.Sentences) == 0 || textBytes == 0 {
		return fmt.Errorf("typly: at least one non-empty sentence is required")
	}
	if textBytes > maxTextBytes {
		return fmt.Errorf("typly: text exceeds %d bytes", maxTextBytes)
	}
	if s.Width <= 0 || s.Width > maxWidth || s.Height <= 0 || s.Height > maxHeight {
		return fmt.Errorf("typly: canvas must be between 1x1 and %dx%d", maxWidth, maxHeight)
	}
	if s.FontSize < 8 || s.FontSize > 256 {
		return fmt.Errorf("typly: font size must be between 8 and 256")
	}
	if s.FPS < 1 || s.FPS > 60 {
		return fmt.Errorf("typly: fps must be between 1 and 60")
	}
	if s.Blinks < 0 || s.Blinks > 20 {
		return fmt.Errorf("typly: blinks must be between 0 and 20")
	}
	if len(s.Cursor) > maxCursorBytes {
		return fmt.Errorf("typly: cursor exceeds %d bytes", maxCursorBytes)
	}
	if s.Emoji != "color" && s.Emoji != "mono" {
		return fmt.Errorf("typly: emoji must be color or mono")
	}
	if _, err := ParseHexColor(s.Foreground); err != nil {
		return fmt.Errorf("invalid foreground: %w", err)
	}
	if _, err := ParseHexColor(s.Background); err != nil {
		return fmt.Errorf("invalid background: %w", err)
	}
	if s.EstimateFrames() > maxFrames {
		return fmt.Errorf("typly: render exceeds %d frames", maxFrames)
	}
	return nil
}

// EstimateFrames returns the number of generated frames before encoding.
func (s RenderSpec) EstimateFrames() int {
	frames := 0
	for _, sentence := range s.Sentences {
		n := len(graphemeClusters(sentence))
		if n == 0 {
			continue
		}
		frames += n + 2*s.Blinks + n
	}
	return frames
}

// Config converts the validated public spec into renderer configuration.
func (s RenderSpec) Config() (Config, error) {
	if err := s.Validate(); err != nil {
		return Config{}, err
	}
	foreground, _ := ParseHexColor(s.Foreground)
	background, _ := ParseHexColor(s.Background)
	return Config{
		Width:      s.Width,
		Height:     s.Height,
		FontSize:   s.FontSize,
		Foreground: foreground,
		Background: background,
		Cursor:     s.Cursor,
		Blinks:     s.Blinks,
		FPS:        s.FPS,
		ColorEmoji: s.Emoji == "color",
	}, nil
}

// SentencesFromText provides the same semicolon convenience as the CLI.
func SentencesFromText(text string) []string {
	var sentences []string
	for _, sentence := range strings.Split(text, ";") {
		if sentence = strings.TrimSpace(sentence); sentence != "" {
			sentences = append(sentences, sentence)
		}
	}
	return sentences
}
