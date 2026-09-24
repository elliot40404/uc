package term

import (
	"os"

	xterm "github.com/charmbracelet/x/term"
)

func IsTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func ColorOK(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || !IsTerminal(f) {
		return false
	}
	return enableVT(f)
}

func Width(f *os.File, fallback int) int {
	w, _, err := xterm.GetSize(f.Fd())
	if err != nil || w <= 0 {
		return fallback
	}
	return w
}
