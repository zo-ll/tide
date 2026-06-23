// Package tide (continued) — scrollable content window. Viewport holds
// a list of visual lines, a visible height, and a scroll offset with
// clamping. It represents one widget's scroll state (e.g. the transcript
// area in oi).
package tide

// Viewport is a bounded window over a list of visual lines. It tracks
// the current scroll offset and clamps it to valid bounds. The caller is
// responsible for rendering the visible slice of Lines.
type Viewport struct {
	Lines  []string
	Height int
	Offset int
}

func (v *Viewport) SetLines(lines []string) {
	v.Lines = append(v.Lines[:0], lines...)
	v.Clamp()
}

func (v *Viewport) Scroll(delta int) {
	v.Offset += delta
	v.Clamp()
}

func (v *Viewport) Bottom() {
	v.Offset = len(v.Lines) - v.Height
	v.Clamp()
}

func (v *Viewport) Visible() []string {
	v.Clamp()
	if v.Height <= 0 || len(v.Lines) == 0 {
		return nil
	}
	end := v.Offset + v.Height
	if end > len(v.Lines) {
		end = len(v.Lines)
	}
	return v.Lines[v.Offset:end]
}

func (v *Viewport) Clamp() {
	if v.Height < 0 {
		v.Height = 0
	}
	max := len(v.Lines) - v.Height
	if max < 0 {
		max = 0
	}
	if v.Offset < 0 {
		v.Offset = 0
	}
	if v.Offset > max {
		v.Offset = max
	}
}
