//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"

	"github.com/elliot40404/uc/internal/tray"
)

func (t *trayApp) open() {
	if err := openDashboard(); err != nil {
		t.menu.setStatus(tray.MenuText("Could not open uc: " + err.Error()))
	}
}

func openDashboard() error {
	path, err := ucPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(path)
	cmd.Dir, _ = os.UserHomeDir()
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_CONSOLE}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func ucPath() (string, error) {
	if self, err := os.Executable(); err == nil {
		near := filepath.Join(filepath.Dir(self), "uc.exe")
		if _, err := os.Stat(near); err == nil {
			return near, nil
		}
	}
	return exec.LookPath("uc.exe")
}
