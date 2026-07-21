package tide

// Style and escape-sequence helpers: synchronized output, SGR colors, and
// OSC-8 hyperlinks. These complement the cursor/attribute helpers in
// terminal.go (HideCursor, MoveTo, Dim, ...) and, like them, write raw
// terminal control sequences that unsupported terminals silently ignore.

import (
	"fmt"
	"io"
)

// BeginSync opens a synchronized output frame (DEC private mode 2026).
// Terminals that support it buffer everything written until EndSync, painting
// the whole frame at once to avoid tearing during animation. Terminals that do
// not support it silently ignore the sequence.
func BeginSync(w io.Writer) { _, _ = io.WriteString(w, "\x1b[?2026h") }

// EndSync closes a synchronized output frame opened by BeginSync, flushing the
// buffered frame to the screen.
func EndSync(w io.Writer) { _, _ = io.WriteString(w, "\x1b[?2026l") }

// Color is one of the eight base ANSI terminal colors, used with SetFg/SetBg.
type Color int

const (
	Black Color = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// SetFg sets the foreground (text) color to one of the eight base ANSI colors.
// Pair with ResetAttr to return to the terminal default.
func SetFg(w io.Writer, c Color) { _, _ = fmt.Fprintf(w, "\x1b[%dm", 30+int(c)) }

// SetBg sets the background color to one of the eight base ANSI colors.
// Pair with ResetAttr to return to the terminal default.
func SetBg(w io.Writer, c Color) { _, _ = fmt.Fprintf(w, "\x1b[%dm", 40+int(c)) }

// ResetAttr clears all SGR attributes (color, dim, etc.) back to default.
func ResetAttr(w io.Writer) { _, _ = io.WriteString(w, "\x1b[0m") }

// Hyperlink wraps text in an OSC-8 hyperlink so supporting terminals render it
// as clickable. If url is empty, text is returned unchanged (no escape noise).
// Terminals without OSC-8 support display text plainly.
func Hyperlink(url, text string) string {
	if url == "" {
		return text
	}
	return "\x1b]8;;" + url + "\x1b\\" + text + "\x1b]8;;\x1b\\"
}
