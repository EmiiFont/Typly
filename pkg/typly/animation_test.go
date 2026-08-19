package typly

import (
	"testing"
)

func TestAnimateFrameCount(t *testing.T) {
	r := newTestRenderer(t)
	sentences := []string{"Hi", "Go!"}
	frames, err := Animate(r, sentences)
	if err != nil {
		t.Fatal(err)
	}
	// Per sentence of n runes: n typing + 2*Blinks blink + (n-1) delete + 1 empty.
	want := 0
	for _, s := range sentences {
		n := len([]rune(s))
		want += n + 2*r.Config().Blinks + (n - 1) + 1
	}
	if len(frames) != want {
		t.Errorf("frame count = %d, want %d", len(frames), want)
	}
}

func TestAnimateSequence(t *testing.T) {
	r := newTestRenderer(t)
	frames, err := Animate(r, []string{"ab"})
	if err != nil {
		t.Fatal(err)
	}
	blinks := r.Config().Blinks
	// Layout: type a, type ab, (off, on)*blinks, delete back to a, empty.
	want := 2 + 2*blinks + 1 + 1
	if len(frames) != want {
		t.Fatalf("frame count = %d, want %d", len(frames), want)
	}

	// Blink "on" frames must be the final typing frame (shared image).
	full := frames[1].Image
	for b := 0; b < blinks; b++ {
		if on := frames[2+2*b+1].Image; on != full {
			t.Errorf("blink %d cursor-on frame is not the full typing frame", b)
		}
	}
	// Delete phase must reuse the typing frame images in reverse.
	if frames[2+2*blinks].Image != frames[0].Image {
		t.Error("delete frame does not reuse the typing frame image")
	}

	// Blink frames hold longer than typing frames.
	if frames[0].Delay >= frames[2].Delay {
		t.Errorf("typing delay %d should be shorter than blink delay %d",
			frames[0].Delay, frames[2].Delay)
	}
}

func TestAnimateBlinkAlternates(t *testing.T) {
	r := newTestRenderer(t)
	frames, err := Animate(r, []string{"x"})
	if err != nil {
		t.Fatal(err)
	}
	// First blink pair: cursor off then cursor on.
	off, on := frames[1].Image, frames[2].Image
	if inkPixels(off) >= inkPixels(on) {
		t.Error("cursor-off blink frame should have less ink than cursor-on")
	}
}

func TestAnimateErrors(t *testing.T) {
	r := newTestRenderer(t)
	if _, err := Animate(r, nil); err == nil {
		t.Error("expected error for no sentences")
	}
	if _, err := Animate(r, []string{"", ""}); err == nil {
		t.Error("expected error for only-empty sentences")
	}
}
