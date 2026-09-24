//go:build windows

package main

import (
	"sync"

	"fyne.io/systray"
)

type menu struct {
	mu      sync.Mutex
	open    func()
	refresh func()
	status  *systray.MenuItem
	lines   []*systray.MenuItem
}

func (m *menu) set(status string, lines []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status == nil || len(lines) != len(m.lines) {
		m.build(len(lines))
	}
	m.status.SetTitle(status)
	for i, l := range lines {
		m.lines[i].SetTitle(l)
	}
}

func (m *menu) setStatus(status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.status != nil {
		m.status.SetTitle(status)
	}
}

func (m *menu) build(n int) {
	systray.ResetMenu()
	m.status = systray.AddMenuItem("", "")
	m.status.Disable()
	m.lines = make([]*systray.MenuItem, n)
	for i := range n {
		m.lines[i] = systray.AddMenuItem("", "")
		on(m.lines[i], m.open)
	}
	systray.AddSeparator()
	on(systray.AddMenuItem("Open dashboard", ""), m.open)
	on(systray.AddMenuItem("Refresh now", ""), m.refresh)
	auto := systray.AddMenuItemCheckbox("Start with Windows", "", autostartOn())
	on(auto, func() { toggleAutostart(auto) })
	on(systray.AddMenuItem("Quit", ""), systray.Quit)
}

func on(item *systray.MenuItem, f func()) {
	go func() {
		for range item.ClickedCh {
			f()
		}
	}()
}
