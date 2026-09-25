package tui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	up        key.Binding
	down      key.Binding
	refresh   key.Binding
	compact   key.Binding
	emails    key.Binding
	settings  key.Binding
	help      key.Binding
	quit      key.Binding
	interrupt key.Binding
	back      key.Binding
	toggle    key.Binding
	left      key.Binding
	right     key.Binding
	hide      key.Binding
	rename    key.Binding
	confirm   key.Binding
	cancel    key.Binding
}

func newKeys() keyMap {
	return keyMap{
		up:        key.NewBinding(key.WithKeys("up", "k", "shift+tab"), key.WithHelp("↑/k", "up")),
		down:      key.NewBinding(key.WithKeys("down", "j", "tab"), key.WithHelp("↓/j", "down")),
		refresh:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		compact:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compact")),
		emails:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "emails")),
		settings:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
		help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),
		quit:      key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
		interrupt: key.NewBinding(key.WithKeys("ctrl+c")),
		back:      key.NewBinding(key.WithKeys("esc", "s", "q"), key.WithHelp("esc", "back")),
		toggle:    key.NewBinding(key.WithKeys("space", "enter"), key.WithHelp("space", "toggle")),
		left:      key.NewBinding(key.WithKeys("left"), key.WithHelp("←/→", "change")),
		right:     key.NewBinding(key.WithKeys("right")),
		hide:      key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "hide/unhide")),
		rename:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "rename")),
		confirm:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save")),
		cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.up, k.down, k.refresh, k.compact, k.emails, k.settings, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.up, k.down}, {k.refresh, k.compact, k.emails, k.settings}, {k.help, k.quit}}
}
