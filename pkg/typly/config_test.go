package typly

import (
	"image/color"
	"math"
	"testing"
)

func TestParseHexColor(t *testing.T) {
	tests := []struct {
		in      string
		want    color.RGBA
		wantErr bool
	}{
		{in: "#000000", want: color.RGBA{0, 0, 0, 0xff}},
		{in: "#FFFFFF", want: color.RGBA{0xff, 0xff, 0xff, 0xff}},
		{in: "1e90ff", want: color.RGBA{0x1e, 0x90, 0xff, 0xff}},
		{in: "#F00", want: color.RGBA{0xff, 0, 0, 0xff}},
		{in: " #abc ", want: color.RGBA{0xaa, 0xbb, 0xcc, 0xff}},
		{in: "", wantErr: true},
		{in: "#12345", wantErr: true},
		{in: "#gggggg", wantErr: true},
	}
	for _, tt := range tests {
		got, err := ParseHexColor(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseHexColor(%q): expected error, got %v", tt.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseHexColor(%q): unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseHexColor(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestConfigValidate(t *testing.T) {
	t.Run("rounds odd dimensions up to even", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Width, cfg.Height = 1279, 719
		if err := cfg.Validate(); err != nil {
			t.Fatal(err)
		}
		if cfg.Width != 1280 || cfg.Height != 720 {
			t.Errorf("got %dx%d, want 1280x720", cfg.Width, cfg.Height)
		}
	})

	t.Run("defaults blank cursor", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Cursor = "  "
		if err := cfg.Validate(); err != nil {
			t.Fatal(err)
		}
		if cfg.Cursor != "|" {
			t.Errorf("cursor = %q, want %q", cfg.Cursor, "|")
		}
	})

	t.Run("rejects bad values", func(t *testing.T) {
		bad := []func(*Config){
			func(c *Config) { c.Width = 0 },
			func(c *Config) { c.Height = -1 },
			func(c *Config) { c.FontSize = 0 },
			func(c *Config) { c.FPS = 0 },
			func(c *Config) { c.Blinks = -1 },
			func(c *Config) { c.FontSize = math.NaN() },
			func(c *Config) { c.FontSize = math.Inf(1) },
		}
		for i, mutate := range bad {
			cfg := DefaultConfig()
			mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Errorf("case %d: expected error", i)
			}
		}
	})
}
