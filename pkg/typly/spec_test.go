package typly

import "testing"

func TestRenderSpecDefaultsAndConfig(t *testing.T) {
	spec := RenderSpec{Sentences: []string{"Hello 🌍"}}
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
	if spec.Width != 1280 || spec.Height != 720 || spec.Emoji != "color" {
		t.Errorf("defaults not applied: %+v", spec)
	}
	cfg, err := spec.Config()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.ColorEmoji || cfg.Foreground == nil || cfg.Background == nil {
		t.Errorf("config conversion failed: %+v", cfg)
	}
}

func TestRenderSpecRejectsOversizedJobs(t *testing.T) {
	tests := []RenderSpec{
		{Sentences: []string{"x"}, Width: 4000},
		{Sentences: []string{"x"}, FontSize: 7},
		{Sentences: []string{"x"}, FPS: 61},
		{Sentences: []string{"x"}, Emoji: "rainbow"},
		{Sentences: []string{"x"}, Foreground: "#not-color"},
	}
	for i := range tests {
		if err := tests[i].Validate(); err == nil {
			t.Errorf("case %d unexpectedly succeeded", i)
		}
	}
}

func TestRenderSpecCountsGraphemeFrames(t *testing.T) {
	spec := DefaultRenderSpec()
	spec.Sentences = []string{"A🌍", "é"}
	if got, want := spec.EstimateFrames(), 2+2*3+2+1+2*3+1; got != want {
		t.Errorf("EstimateFrames() = %d, want %d", got, want)
	}
}

func TestSentencesFromText(t *testing.T) {
	got := SentencesFromText(" first ; ;second ")
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Errorf("SentencesFromText() = %#v", got)
	}
}
