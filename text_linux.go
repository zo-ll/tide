//go:build linux

package tide

// Linux terminal control via direct syscalls. macOS/BSD use different ioctl
// request numbers and are served by the stty-based fallback in text_other.go.
// The three functions here (getTermios, setTermios, terminalSize) share their
// signatures with that fallback so terminal.go compiles against either.

import (
	"os"
	"syscall"
	"unsafe"
)

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
