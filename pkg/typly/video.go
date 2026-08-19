package typly

import (
	"fmt"
	"image/png"
	"io"
	"os/exec"
	"strconv"
)

// EncodeMP4 encodes the animation to an H.264 MP4 file by piping
// PNG frames to the ffmpeg CLI, which must be available on PATH.
func EncodeMP4(path string, frames []Frame, fps int) error {
	if len(frames) == 0 {
		return fmt.Errorf("typly: no frames to encode")
	}
	if fps <= 0 {
		return fmt.Errorf("typly: invalid fps %d", fps)
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("typly: ffmpeg not found on PATH (required for MP4 output): %w", err)
	}

	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "image2pipe",
		"-framerate", strconv.Itoa(fps),
		"-i", "-",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		path,
	)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("typly: ffmpeg stdin: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("typly: start ffmpeg: %w", err)
	}

	writeErr := writeFrames(stdin, frames, fps)
	closeErr := stdin.Close()
	waitErr := cmd.Wait()
	if waitErr != nil {
		return fmt.Errorf("typly: ffmpeg failed: %w", waitErr)
	}
	if writeErr != nil {
		return fmt.Errorf("typly: write frames to ffmpeg: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("typly: close ffmpeg stdin: %w", closeErr)
	}
	return nil
}

// writeFrames PNG-encodes each frame to w, repeating frames so their
// per-frame delays are preserved at the fixed output frame rate.
func writeFrames(w io.Writer, frames []Frame, fps int) error {
	for _, f := range frames {
		for i := 0; i < frameRepeats(f.Delay, fps); i++ {
			if err := png.Encode(w, f.Image); err != nil {
				return err
			}
		}
	}
	return nil
}

// frameRepeats converts a delay in hundredths of a second into a
// number of frames at the given rate, at minimum one.
func frameRepeats(delayCS, fps int) int {
	n := (delayCS*fps + 50) / 100 // round to nearest
	if n < 1 {
		n = 1
	}
	return n
}
