package tide

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Size struct {
	Rows int
	Cols int
}

type Terminal struct {
	In        *os.File
	Out       io.Writer
	rawState  string
	raw       bool
	alt       bool
	mouse     bool
	altScroll bool
}

func Open(in *os.File, out io.Writer) (*Terminal, error) {
	if in == nil {
		return nil, fmt.Errorf("nil input terminal")
	}
	if out == nil {
		return nil, fmt.Errorf("nil output terminal")
	}
	return &Terminal{In: in, Out: out}, nil
}

func (t *Terminal) Close() error {
	var err error
	if t.mouse {
		if e := t.DisableMouse(); err == nil && e != nil {
			err = e
		}
	}
	if t.altScroll {
		if e := t.DisableAltScroll(); err == nil && e != nil {
			err = e
		}
	}
	if t.alt {
		if e := t.LeaveAltScreen(); err == nil && e != nil {
			err = e
		}
	}
	if t.raw {
		if e := t.LeaveRaw(); err == nil && e != nil {
			err = e
		}
	}
	return err
}

func (t *Terminal) EnterRaw() error {
	if t.raw {
		return nil
	}
	state, err := sttyCapture(t.In, "-g")
	if err != nil {
		return err
	}
	if err := sttyRun(t.In, "raw", "-echo"); err != nil {
		return err
	}
	t.rawState = strings.TrimSpace(state)
	t.raw = true
	_, _ = io.WriteString(t.Out, "\x1b[?2004h")
	return nil
}

func (t *Terminal) LeaveRaw() error {
	if !t.raw {
		return nil
	}
	_, _ = io.WriteString(t.Out, "\x1b[?2004l")
	t.raw = false
	if t.rawState == "" {
		return nil
	}
	return sttyRun(t.In, t.rawState)
}

func (t *Terminal) EnterAltScreen() error {
	if t.alt {
		return nil
	}
	_, err := io.WriteString(t.Out, "\x1b[?1049h\x1b[H\x1b[2J")
	t.alt = err == nil
	return err
}

func (t *Terminal) LeaveAltScreen() error {
	if !t.alt {
		return nil
	}
	t.alt = false
	_, err := io.WriteString(t.Out, "\x1b[?1049l")
	return err
}

func (t *Terminal) EnableMouse() error {
	if t.mouse {
		return nil
	}
	if t.altScroll {
		if err := t.DisableAltScroll(); err != nil {
			return err
		}
	}
	_, err := io.WriteString(t.Out, "\x1b[?1000h\x1b[?1006h")
	t.mouse = err == nil
	return err
}

func (t *Terminal) DisableMouse() error {
	if !t.mouse {
		return nil
	}
	t.mouse = false
	_, err := io.WriteString(t.Out, "\x1b[?1000l\x1b[?1006l")
	return err
}

// EnableAltScroll asks terminals to translate wheel movement in the alternate
// screen into cursor-key events. Unlike full mouse tracking, it preserves the
// terminal's native click-and-drag text selection behavior.
func (t *Terminal) EnableAltScroll() error {
	if t.altScroll {
		return nil
	}
	if t.mouse {
		if err := t.DisableMouse(); err != nil {
			return err
		}
	}
	_, err := io.WriteString(t.Out, "\x1b[?1007h")
	t.altScroll = err == nil
	return err
}

func (t *Terminal) DisableAltScroll() error {
	if !t.altScroll {
		return nil
	}
	t.altScroll = false
	_, err := io.WriteString(t.Out, "\x1b[?1007l")
	return err
}

func (t *Terminal) Size() Size {
	cols := terminalWidth(t.In)
	rows := 24
	if out, err := sttyCapture(t.In, "size"); err == nil {
		parts := strings.Fields(strings.TrimSpace(out))
		if len(parts) == 2 {
			if r, convErr := strconv.Atoi(parts[0]); convErr == nil && r > 0 {
				rows = r
			}
			if c, convErr := strconv.Atoi(parts[1]); convErr == nil && c > 0 {
				cols = c
			}
		}
	}
	return Size{Rows: rows, Cols: cols}
}

func (t *Terminal) NewEditor(prompt string, complete Completer) *Editor {
	return New(t.In, t.Out, prompt, complete)
}

func HideCursor(w io.Writer)           { _, _ = io.WriteString(w, "\x1b[?25l") }
func ShowCursor(w io.Writer)           { _, _ = io.WriteString(w, "\x1b[?25h") }
func ClearScreen(w io.Writer)          { _, _ = io.WriteString(w, "\x1b[H\x1b[2J") }
func MoveTo(w io.Writer, row, col int) { _, _ = fmt.Fprintf(w, "\x1b[%d;%dH", row, col) }
func ClearLine(w io.Writer)            { _, _ = io.WriteString(w, "\x1b[2K") }

func Dim(text string) string               { return dim(text) }
func Warn(text string) string              { return "\x1b[33m" + text + "\x1b[0m" }
func Command(text string) string           { return "\x1b[36m" + text + "\x1b[0m" }
func DisplayWidth(text string) int         { return displayWidth(text) }
func Wrap(text string, width int) []string { return wrapLineFromRunes([]rune(text), width) }
