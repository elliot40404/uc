//go:build windows

package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
)

func logCrashes() {
	dir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	dir = filepath.Join(dir, "uc")
	if os.MkdirAll(dir, 0o700) != nil {
		return
	}
	f, err := os.OpenFile(filepath.Join(dir, "uctray.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	debug.SetCrashOutput(f, debug.CrashOptions{})
}
