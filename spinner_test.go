package tide

import "testing"

func TestSpinnerCyclesAndWraps(t *testing.T) {
	s := NewSpinner()
	frames := []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")
	for i := 0; i < len(frames); i++ {
		if got := s.Tick(); got != string(frames[i]) {
			t.Fatalf("tick %d = %q, want %q", i, got, string(frames[i]))
		}
	}
	// After a full cycle it wraps to the first frame again.
	if got := s.Tick(); got != string(frames[0]) {
		t.Fatalf("wrap tick = %q, want %q", got, string(frames[0]))
	}
}

func TestSpinnerFramesAreSingleGlyphs(t *testing.T) {
	s := NewSpinner()
	for i := 0; i < 2*len(spinnerFrames); i++ {
		if got := []rune(s.Tick()); len(got) != 1 {
			t.Fatalf("tick %d returned %d runes, want 1", i, len(got))
		}
	}
}
