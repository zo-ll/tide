//go:build !linux

package tide

// Non-Linux terminal control via the stty(1) subprocess. macOS/BSD expose
// different ioctl request numbers and termios struct layouts than Linux, so
// rather than track each platform's constants we defer to stty, which every
// POSIX system ships. These mirror the signatures of the syscall-based
// implementation in text_linux.go so terminal.go is portable unchanged.
//
// Because the shared *syscall.Termios signature cannot carry stty's opaque
// "-g" state string, the saved settings are held in package-level maps keyed
// by file descriptor. getTermios captures them; the first setTermios after a
// capture enters raw mode; the next restores. This matches how terminal.go
// pairs EnterRaw with LeaveRaw.

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var (
	sttySaved = map[uintptr]string{} // fd -> `stty -g` snapshot
	sttyRaw   = map[uintptr]bool{}   // fd -> raw mode currently applied
)

// getTermios captures the terminal's current settings via `stty -g`.
// The returned *syscall.Termios is an opaque non-nil handle whose fields are
// ignored on this platform (raw mode is configured through stty in
// setTermios). Returns nil when f is nil or not a terminal.
func getTermios(f *os.File) (*syscall.Termios, error) {
	if f == nil || !IsTerminal(f) {
		return nil, nil
	}
	out, err := sttyCapture(f, "-g")
	if err != nil {
		return nil, err
	}
	sttySaved[f.Fd()] = strings.TrimSpace(out)
	return &syscall.Termios{}, nil
}

// setTermios applies terminal settings via stty. The first call after a
// getTermios capture enters raw mode (`stty raw -echo`); the next restores the
// captured settings. State (nil) or non-terminals are no-ops.
func setTermios(f *os.File, state *syscall.Termios) error {
	if f == nil || state == nil || !IsTerminal(f) {
		return nil
	}
	fd := f.Fd()
	if !sttyRaw[fd] {
		sttyRaw[fd] = true
		return sttyRun(f, "raw", "-echo")
	}
	saved := sttySaved[fd]
	delete(sttyRaw, fd)
	delete(sttySaved, fd)
	if saved == "" {
		return nil
	}
	return sttyRun(f, saved)
}

// terminalSize queries the terminal size via `stty size`.
// Returns rows, cols. Falls back to 80x24 on error.
func terminalSize(f *os.File) (rows, cols int) {
	rows, cols = 24, 80
	if f == nil {
		return rows, cols
	}
	out, err := sttyCapture(f, "size")
	if err != nil {
		return rows, cols
	}
	parts := strings.Fields(strings.TrimSpace(out))
	if len(parts) != 2 {
		return rows, cols
	}
	if r, e := strconv.Atoi(parts[0]); e == nil && r > 0 {
		rows = r
	}
	if c, e := strconv.Atoi(parts[1]); e == nil && c > 0 {
		cols = c
	}
	return rows, cols
}

// sttyCapture runs `stty args...` with f as its controlling terminal and
// returns stdout.
func sttyCapture(f *os.File, args ...string) (string, error) {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = f
	out, err := cmd.Output()
	return string(out), err
}

// sttyRun runs `stty args...` with f as its controlling terminal, discarding
// output.
func sttyRun(f *os.File, args ...string) error {
	cmd := exec.Command("stty", args...)
	cmd.Stdin = f
	return cmd.Run()
}
