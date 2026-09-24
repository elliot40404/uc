//go:build windows

package main

import (
	"os"
	"strings"

	"fyne.io/systray"
	"golang.org/x/sys/windows/registry"
)

const (
	runKey   = `Software\Microsoft\Windows\CurrentVersion\Run`
	runValue = "uc tray"
)

func autostartOn() bool {
	want, err := command()
	if err != nil {
		return false
	}
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()
	got, _, err := k.GetStringValue(runValue)
	return err == nil && strings.EqualFold(got, want)
}

func toggleAutostart(item *systray.MenuItem) {
	if err := setAutostart(!autostartOn()); err != nil {
		return
	}
	if autostartOn() {
		item.Check()
	} else {
		item.Uncheck()
	}
}

func setAutostart(on bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		return k.DeleteValue(runValue)
	}
	cmd, err := command()
	if err != nil {
		return err
	}
	return k.SetStringValue(runValue, cmd)
}

func command() (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", err
	}
	return `"` + self + `"`, nil
}
