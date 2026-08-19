package typly

import (
	"bytes"
	"image/gif"
	"testing"
)

func TestEncodeGIFRoundTrip(t *testing.T) {
	r := newTestRenderer(t)
	frames, err := Animate(r, []string{"Go"})
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := EncodeGIF(&buf, frames); err != nil {
		t.Fatal(err)
	}

	decoded, err := gif.DecodeAll(&buf)
	if err != nil {
		t.Fatalf("decode gif: %v", err)
	}
	if len(decoded.Image) != len(frames) {
		t.Errorf("decoded %d frames, want %d", len(decoded.Image), len(frames))
	}
	if decoded.LoopCount != 0 {
		t.Errorf("loop count = %d, want 0 (infinite)", decoded.LoopCount)
	}
	for i, f := range frames {
		if decoded.Delay[i] != f.Delay {
			t.Errorf("frame %d delay = %d, want %d", i, decoded.Delay[i], f.Delay)
		}
	}
}

func TestEncodeGIFEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := EncodeGIF(&buf, nil); err == nil {
		t.Error("expected error for no frames")
	}
}
