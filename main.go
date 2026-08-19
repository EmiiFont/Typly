// Command typly generates typing-animation GIFs, MP4 videos, and PNG
// frame sequences from a list of sentences.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiiFont/typly/pkg/typly"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "typly:", err)
		os.Exit(1)
	}
}

func run(args []string, errW *os.File) error {
	fs := flag.NewFlagSet("typly", flag.ContinueOnError)
	fs.SetOutput(errW)

	var (
		text      = fs.String("text", "", "sentences to animate, separated by ';' (required)")
		out       = fs.String("out", ".", "output directory")
		gifOut    = fs.Bool("gif", true, "write typly.gif")
		mp4Out    = fs.Bool("mp4", false, "write typly.mp4 (requires ffmpeg on PATH)")
		framesOut = fs.Bool("frames", false, "write individual PNG frames to <out>/frames")
		size      = fs.String("size", "1280x720", "canvas size as WIDTHxHEIGHT")
		fps       = fs.Int("fps", 12, "typing frame rate")
		fontPath  = fs.String("font", "", "path to a TTF/OTF font (default: embedded Go Regular)")
		fontSize  = fs.Float64("font-size", 72, "font size in points")
		fg        = fs.String("fg", "#000000", "text color as hex")
		bg        = fs.String("bg", "#FFFFFF", "background color as hex")
		cursor    = fs.String("cursor", "|", "cursor glyph")
		blinks    = fs.Int("blinks", 3, "cursor blink cycles at the end of each sentence")
		emojiMode = fs.String("emoji", "color", "emoji mode: color or mono")
	)
	fs.Usage = func() {
		fmt.Fprintf(errW, "Usage: typly -text \"Hello!;World!\" [flags]\n\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	sentences := splitSentences(*text)
	if len(sentences) == 0 {
		fs.Usage()
		return fmt.Errorf("-text is required")
	}
	if !*gifOut && !*mp4Out && !*framesOut {
		return fmt.Errorf("nothing to do: enable at least one of -gif, -mp4, -frames")
	}

	cfg := typly.DefaultConfig()
	cfg.FPS = *fps
	cfg.FontSize = *fontSize
	cfg.Cursor = *cursor
	cfg.Blinks = *blinks
	switch strings.ToLower(*emojiMode) {
	case "color":
		cfg.ColorEmoji = true
	case "mono":
		cfg.ColorEmoji = false
	default:
		return fmt.Errorf("invalid -emoji %q (expected color or mono)", *emojiMode)
	}

	var err error
	if cfg.Width, cfg.Height, err = parseSize(*size); err != nil {
		return err
	}
	if cfg.Foreground, err = typly.ParseHexColor(*fg); err != nil {
		return err
	}
	if cfg.Background, err = typly.ParseHexColor(*bg); err != nil {
		return err
	}
	if *fontPath != "" {
		if cfg.FontData, err = os.ReadFile(*fontPath); err != nil {
			return fmt.Errorf("read font: %w", err)
		}
	}

	renderer, err := typly.NewRenderer(cfg)
	if err != nil {
		return err
	}
	defer renderer.Close()

	frames, err := typly.Animate(renderer, sentences)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	if *gifOut {
		path := filepath.Join(*out, "typly.gif")
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		err = typly.EncodeGIF(f, frames)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}
	if *mp4Out {
		path := filepath.Join(*out, "typly.mp4")
		if err := typly.EncodeMP4(path, frames, cfg.FPS); err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}
	if *framesOut {
		dir := filepath.Join(*out, "frames")
		if err := typly.WriteFrames(dir, frames); err != nil {
			return err
		}
		fmt.Printf("wrote %d frames to %s\n", len(frames), dir)
	}
	return nil
}

// splitSentences splits the -text flag on ';' and drops empty entries.
func splitSentences(text string) []string {
	var out []string
	for _, s := range strings.Split(text, ";") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// parseSize parses "WIDTHxHEIGHT", e.g. "1280x720".
func parseSize(s string) (int, int, error) {
	var w, h int
	if _, err := fmt.Sscanf(strings.ToLower(s), "%dx%d", &w, &h); err != nil || w <= 0 || h <= 0 || fmt.Sprintf("%dx%d", w, h) != strings.ToLower(s) {
		return 0, 0, fmt.Errorf("invalid -size %q (expected WIDTHxHEIGHT)", s)
	}
	return w, h, nil
}
