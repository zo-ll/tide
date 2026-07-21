package tide

// Spinner is a frame-state animator for indeterminate progress. It owns no
// goroutine, timer, or interval: the caller drives it by calling Tick() each
// render and writing the returned frame wherever it likes (e.g. via
// WriteClipped at a fixed row/col). This keeps animation tied to the app's own
// event loop rather than a background ticker.
type Spinner struct {
	pos int
}

// spinnerFrames is the braille cycle rendered one glyph per Tick.
var spinnerFrames = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// NewSpinner returns a spinner positioned at the first frame.
func NewSpinner() *Spinner { return &Spinner{} }

// Tick returns the current frame glyph and advances to the next, wrapping
// around at the end of the cycle.
func (s *Spinner) Tick() string {
	frame := spinnerFrames[s.pos]
	s.pos = (s.pos + 1) % len(spinnerFrames)
	return string(frame)
}
