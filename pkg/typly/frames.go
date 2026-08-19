package typly

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
)

// WriteFrames saves every animation frame as a numbered PNG in dir,
// creating the directory if needed.
func WriteFrames(dir string, frames []Frame) error {
	if len(frames) == 0 {
		return fmt.Errorf("typly: no frames to write")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("typly: create frames dir: %w", err)
	}
	for i, f := range frames {
		name := filepath.Join(dir, fmt.Sprintf("frame-%04d.png", i))
		file, err := os.Create(name)
		if err != nil {
			return fmt.Errorf("typly: create %s: %w", name, err)
		}
		err = png.Encode(file, f.Image)
		if cerr := file.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return fmt.Errorf("typly: write %s: %w", name, err)
		}
	}
	return nil
}
