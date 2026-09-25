package tui

import (
	"charm.land/bubbles/v2/key"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/usage"
)

var settingLabels = [rowCount]string{"live refresh", "compact rows", "show emails", "mini by default", "alt screen", "refresh every"}

func (m Model) settingsLines() ([]string, int) {
	t := m.theme
	lines := []string{t.fg(t.accent).Bold(true).Render("SETTINGS"), m.saveStatus(), ""}
	for r := range rowCount {
		lines = append(lines, m.settingLine(r))
	}
	lines = append(lines, "", t.fg(t.accent).Bold(true).Render("ACCOUNTS"))
	sel := 3 + m.setRow
	if m.setRow >= rowCount {
		sel = len(lines) + m.setRow - rowCount
	}
	return append(lines, m.accountLines()...), sel
}

func (m Model) saveStatus() string {
	t := m.theme
	if m.saveErr != nil {
		return t.fg(t.bad).Render("not saved: " + usage.TerminalText(m.saveErr.Error()))
	}
	return t.fg(t.subtle).Render("changes save to config.json right away")
}

func (m Model) settingLine(r int) string {
	t := m.theme
	cursor, label := "  ", t.fg(t.text).Render(settingLabels[r])
	if r == m.setRow {
		cursor, label = t.fg(t.accent).Render("› "), t.fg(t.text).Bold(true).Render(settingLabels[r])
	}
	switch r {
	case rowEvery:
		return cursor + "    " + label + "   " + t.fg(t.accent).Render("‹ "+config.FormatEvery(m.every)+" ›")
	case rowMini:
		return cursor + m.checkbox(m.cfg.Defaults.Mini) + " " + label + t.fg(t.subtle).Render("  next launch")
	}
	return cursor + m.checkbox(m.settingOn(r)) + " " + label
}

func (m Model) settingOn(r int) bool {
	switch r {
	case rowLive:
		return m.live
	case rowCompact:
		return m.compact
	case rowEmails:
		return m.emails
	case rowAltScreen:
		return m.altScreen
	}
	return false
}

func (m Model) checkbox(on bool) string {
	if on {
		return m.theme.fg(m.theme.good).Render("[x]")
	}
	return m.theme.fg(m.theme.subtle).Render("[ ]")
}

func (m Model) settingsHelp() string {
	if m.renaming {
		return m.help.ShortHelpView([]key.Binding{m.keys.confirm, m.keys.cancel})
	}
	if m.setRow >= rowCount {
		return m.help.ShortHelpView([]key.Binding{m.keys.up, m.keys.down, m.keys.hide, m.keys.rename, m.keys.back})
	}
	return m.help.ShortHelpView([]key.Binding{m.keys.up, m.keys.down, m.keys.toggle, m.keys.left, m.keys.back})
}
