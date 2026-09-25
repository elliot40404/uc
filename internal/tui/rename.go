package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/elliot40404/uc/internal/config"
)

func (m Model) startRename() (Model, tea.Cmd) {
	it, ok := m.selectedItem()
	if !ok {
		return m, nil
	}
	in := textinput.New()
	in.Prompt = ""
	in.Placeholder = "default name"
	in.CharLimit = config.MaxNameLen
	in.SetWidth(config.MaxNameLen)
	in.SetValue(it.Name)
	m.input, m.renaming, m.renameOf = in, true, it
	return m, m.input.Focus()
}

func (m Model) onRenameKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.interrupt):
		m.quitting = true
		return m, tea.Quit
	case key.Matches(msg, m.keys.cancel):
		m.renaming = false
		return m, nil
	case key.Matches(msg, m.keys.confirm):
		m.renaming = false
		m.cfg = m.cfg.SetName(m.renameOf.Provider, m.renameOf.Dir, m.input.Value())
		return m.persist().refetch()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
