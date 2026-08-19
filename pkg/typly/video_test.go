package typly

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFrameRepeats(t *testing.T) {
	tests := []struct {
		delayCS, fps, want int
	}{
		{delayCS: 8, fps: 12, want: 1},   // one typing frame at 12fps
		{delayCS: 50, fps: 12, want: 6},  // blink hold ~= half a second
		{delayCS: 50, fps: 30, want: 15}, // scales with fps
		{delayCS: 1, fps: 12, want: 1},   // never less than one frame
		{delayCS: 0, fps: 12, want: 1},
	}
	for _, tt := range tests {
		if got := frameRepeats(tt.delayCS, tt.fps); got != tt.want {
			t.Errorf("frameRepeats(%d, %d) = %d, want %d", tt.delayCS, tt.fps, got, tt.want)
		}
	}
}

func TestEncodeMP4Validation(t *testing.T) {
	if err := EncodeMP4("out.mp4", nil, 12); err == nil {
		t.Error("expected error for no frames")
	}
	r := newTestRenderer(t)
	frames, err := Animate(r, []string{"x"})
	if err != nil {
		t.Fatal(err)
	}
	if err := EncodeMP4("out.mp4", frames, 0); err == nil {
		t.Error("expected error for invalid fps")
	}
}

func TestEncodeMP4Integration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	r := newTestRenderer(t)
	frames, err := Animate(r, []string{"Go"})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "typly.mp4")
	if err := EncodeMP4(path, frames, r.Config().FPS); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("mp4 file is empty")
	}
}
