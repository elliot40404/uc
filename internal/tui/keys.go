package tui

import "charm.land/bubbles/v2/key"

type keyMap struct {
	up      key.Binding
	down    key.Binding
	refresh key.Binding
	compact key.Binding
	emails  key.Binding
	help    key.Binding
	quit    key.Binding
}

func newKeys() keyMap {
	return keyMap{
		up:      key.NewBinding(key.WithKeys("up", "k", "shift+tab"), key.WithHelp("↑/k", "up")),
		down:    key.NewBinding(key.WithKeys("down", "j", "tab"), key.WithHelp("↓/j", "down")),
		refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		compact: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compact")),
		emails:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "emails")),
		help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "more")),
		quit:    key.NewBinding(key.WithKeys("q", "esc", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.up, k.down, k.refresh, k.compact, k.emails, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.up, k.down}, {k.refresh, k.compact, k.emails}, {k.help, k.quit}}
}
