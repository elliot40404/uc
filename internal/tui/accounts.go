package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/usage"
)

func (m Model) items() []config.Item {
	return m.cfg.Merge(m.found)
}

func (m Model) settingRows() int {
	return rowCount + len(m.items())
}

func (m Model) selectedItem() (config.Item, bool) {
	items, i := m.items(), m.setRow-rowCount
	if i < 0 || i >= len(items) {
		return config.Item{}, false
	}
	return items[i], true
}

func (m Model) toggleHidden() (Model, tea.Cmd) {
	it, ok := m.selectedItem()
	if !ok {
		return m, nil
	}
	m.cfg = m.cfg.SetHidden(it.Provider, it.Dir, !it.Hide)
	return m.persist().refetch()
}

func (m Model) refetch() (Model, tea.Cmd) {
	if m.loading {
		m.reload = true
		return m, nil
	}
	return m.startLoad()
}

func (m Model) accountLines() []string {
	t := m.theme
	items := m.items()
	if len(items) == 0 {
		return []string{t.fg(t.subtle).Render("  no accounts found")}
	}
	nameW := 0
	for _, it := range items {
		nameW = max(nameW, lipgloss.Width(usage.TerminalText(it.Name)))
	}
	lines := make([]string, len(items))
	for i, it := range items {
		lines[i] = m.accountLine(it, rowCount+i == m.setRow, nameW)
	}
	return lines
}

func (m Model) accountLine(it config.Item, selected bool, nameW int) string {
	t := m.theme
	cursor, name := "  ", t.fg(t.text)
	if selected {
		cursor, name = t.fg(t.accent).Render("› "), name.Bold(true)
	}
	if it.Hide {
		name = name.Foreground(t.subtle)
	}
	label := name.Width(nameW + 2).Render(usage.TerminalText(it.Name))
	if selected && m.renaming {
		label = m.input.View() + "  "
	}
	line := cursor + t.fg(t.provider(it.Provider)).Width(8).Render(it.Provider) + label +
		t.fg(t.subtle).Render(usage.TerminalText(config.ShortHome(it.Dir, m.home)))
	if it.Hide {
		line += t.fg(t.warn).Render("  hidden")
	}
	return line
}
