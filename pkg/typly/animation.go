package typly

import (
	"fmt"
	"image"
	"strings"

	"github.com/rivo/uniseg"
)

// Frame is a single animation frame with its display duration.
type Frame struct {
	Image *image.RGBA
	Delay int // hundredths of a second
}

// blinkDelay is how long each cursor blink state is held, in
// hundredths of a second (roughly typed.js's cadence).
const blinkDelay = 50

// Animate renders the full animation for the given sentences:
// each sentence is typed grapheme by grapheme, the cursor blinks at
// the end, and the sentence is then "deleted" by replaying its typing
// frames in reverse. Frames are shared between the typing and deleting
// phases, so memory cost is one image per typed character.
func Animate(r *Renderer, sentences []string) ([]Frame, error) {
	if len(sentences) == 0 {
		return nil, fmt.Errorf("typly: no sentences to animate")
	}
	typeDelay := 100 / r.Config().FPS
	if typeDelay < 2 {
		typeDelay = 2 // GIF's practical minimum delay
	}

	var frames []Frame
	for _, sentence := range sentences {
		clusters := graphemeClusters(sentence)
		if len(clusters) == 0 {
			continue
		}

		// Typing phase: one frame per typed character.
		typed := make([]*image.RGBA, len(clusters))
		for i := range clusters {
			typed[i] = r.Frame(strings.Join(clusters[:i+1], ""), true)
			frames = append(frames, Frame{Image: typed[i], Delay: typeDelay})
		}

		// Blink phase at the end of the sentence.
		cursorOff := r.Frame(sentence, false)
		cursorOn := typed[len(typed)-1]
		for b := 0; b < r.Config().Blinks; b++ {
			frames = append(frames,
				Frame{Image: cursorOff, Delay: blinkDelay},
				Frame{Image: cursorOn, Delay: blinkDelay},
			)
		}

		// Deleting phase: replay typing frames in reverse, ending on
		// an empty canvas so the next sentence (or the loop restart)
		// begins cleanly.
		for i := len(typed) - 2; i >= 0; i-- {
			frames = append(frames, Frame{Image: typed[i], Delay: typeDelay})
		}
		frames = append(frames, Frame{Image: r.Frame("", true), Delay: blinkDelay})
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("typly: no sentences to animate")
	}
	return frames, nil
}

func graphemeClusters(text string) []string {
	var clusters []string
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		clusters = append(clusters, g.Str())
	}
	return clusters
}
