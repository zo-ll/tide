package tide

// This file holds the text measurement and wrapping engine shared by the
// terminal toolkit (DisplayWidth, Wrap) and the widget renderers, plus the
// termios/ioctl helpers used by Terminal. It is the only surviving logic
// from tide's original line-editor implementation; the Editor itself was
// removed in favor of the async Overlay/Input/Picker surface.

import (
	"os"
	"syscall"
	"unsafe"
)

func dim(text string) string {
	return "\x1b[2m" + text + "\x1b[0m"
}

// displayWidth returns the visible cell width of s, honoring combining marks
// and East-Asian wide/fullwidth runes.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 32 || (r >= 0x7f && r < 0xa0):
		return 0
	case isCombining(r):
		return 0
	case isWideRune(r):
		return 2
	default:
		return 1
	}
}

func isCombining(r rune) bool {
	switch {
	case r >= 0x0300 && r <= 0x036f:
	case r >= 0x1ab0 && r <= 0x1aff:
	case r >= 0x1dc0 && r <= 0x1dff:
	case r >= 0x20d0 && r <= 0x20ff:
	case r >= 0xfe20 && r <= 0xfe2f:
	default:
		return false
	}
	return true
}

func isWideRune(r rune) bool {
	switch {
	case r >= 0x1100 && r <= 0x115f:
	case r >= 0x2329 && r <= 0x232a:
	case r >= 0x2e80 && r <= 0xa4cf:
	case r >= 0xac00 && r <= 0xd7a3:
	case r >= 0xf900 && r <= 0xfaff:
	case r >= 0xfe10 && r <= 0xfe19:
	case r >= 0xfe30 && r <= 0xfe6f:
	case r >= 0xff00 && r <= 0xff60:
	case r >= 0xffe0 && r <= 0xffe6:
	case r >= 0x1f300 && r <= 0x1faff:
	case r >= 0x20000 && r <= 0x3fffd:
	default:
		return false
	}
	return true
}

func isWordSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

// wrapLineFromRunes wraps runes to width display cells, preferring word
// (whitespace) boundaries and hard-wrapping long words. It backs Wrap and
// WrapPlain.
func wrapLineFromRunes(runes []rune, width int) []string {
	if width <= 1 || len(runes) == 0 {
		return []string{string(runes)}
	}
	var lines []string
	for len(runes) > 0 {
		used := 0
		cut := 0
		lastSpace := -1
		for cut < len(runes) {
			rw := runeWidth(runes[cut])
			if rw < 0 {
				rw = 0
			}
			if cut > 0 && used+rw > width {
				break
			}
			used += rw
			if isWordSpace(runes[cut]) {
				lastSpace = cut
			}
			cut++
			if used >= width {
				break
			}
		}
		if cut >= len(runes) {
			lines = append(lines, string(runes))
			break
		}
		if lastSpace > 0 {
			lines = append(lines, string(trimRightSpaces(runes[:lastSpace])))
			runes = trimLeftSpaces(runes[lastSpace+1:])
			continue
		}
		if cut == 0 {
			cut = 1
		}
		lines = append(lines, string(runes[:cut]))
		runes = runes[cut:]
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func trimLeftSpaces(runes []rune) []rune {
	for len(runes) > 0 && isWordSpace(runes[0]) {
		runes = runes[1:]
	}
	return runes
}

func trimRightSpaces(runes []rune) []rune {
	for len(runes) > 0 && isWordSpace(runes[len(runes)-1]) {
		runes = runes[:len(runes)-1]
	}
	return runes
}

// getTermios reads termios for the fd underlying f.
// Returns nil and no error when f is nil or not a terminal.
func getTermios(f *os.File) (*syscall.Termios, error) {
	if f == nil {
		return nil, nil
	}
	var t syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&t)))
	if errno == syscall.ENOTTY {
		return nil, nil
	}
	if errno != 0 {
		return nil, errno
	}
	return &t, nil
}

// setTermios writes termios for the fd underlying f.
func setTermios(f *os.File, state *syscall.Termios) error {
	if f == nil || state == nil {
		return nil
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(state)))
	if errno == syscall.ENOTTY {
		return nil
	}
	if errno != 0 {
		return errno
	}
	return nil
}

// terminalSize queries the terminal size via TIOCGWINSZ ioctl.
// Returns rows, cols. Falls back to 80x24 on error.
func terminalSize(f *os.File) (rows, cols int) {
	if f == nil {
		return 24, 80
	}
	var ws struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Row == 0 || ws.Col == 0 {
		return 24, 80
	}
	return int(ws.Row), int(ws.Col)
}
