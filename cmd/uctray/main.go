//go:build windows

package main

import (
	"errors"

	"fyne.io/systray"
	"golang.org/x/sys/windows"
)

const mutexName = `Local\uc-tray`

func main() {
	lock, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(mutexName))
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		return
	}
	defer windows.CloseHandle(lock)
	t := newTray()
	systray.Run(t.start, nil)
}
