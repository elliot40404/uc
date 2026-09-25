//go:build windows

package main

import "fyne.io/systray"

func buildMenu(refresh func()) {
	on(systray.AddMenuItem("Refresh now", ""), refresh)
	auto := systray.AddMenuItemCheckbox("Start with Windows", "", autostartOn())
	on(auto, func() { toggleAutostart(auto) })
	systray.AddSeparator()
	on(systray.AddMenuItem("Quit", ""), systray.Quit)
}

func on(item *systray.MenuItem, f func()) {
	go func() {
		for range item.ClickedCh {
			f()
		}
	}()
}
