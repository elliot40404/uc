package tui

import (
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/usage"
)

func NewMini(fetch Fetch, o Options) Model {
	o.Compact = true
	m := New(fetch, o)
	m.inline = true
	return m
}

func Print(w io.Writer, reports []usage.Report, at time.Time, width int, dark bool, o Options) error {
	m := Model{theme: newTheme(dark), reports: usage.TerminalReports(reports), fetchedAt: at, now: at, width: width, compact: true, emails: o.Emails, inline: true, selected: -1}
	_, err := lipgloss.Fprintln(w, m.frame())
	return err
}

func (m Model) miniFrame() string {
	w := m.width - 2*marginX
	parts := []string{m.headerCore(w), ""}
	switch {
	case m.err != nil && len(m.reports) == 0:
		parts = append(parts, m.theme.fg(m.theme.bad).Render("Could not load accounts: "+usage.TerminalText(m.err.Error())))
	case len(m.reports) == 0 && m.loading:
		parts = append(parts, m.spin.View()+m.theme.fg(m.theme.subtle).Render(" fetching usage…"))
	default:
		parts = append(parts, m.rows(w)...)
	}
	if m.live && !m.quitting {
		parts = append(parts, "", m.help.ShortHelpView([]key.Binding{m.keys.up, m.keys.down, m.keys.refresh, m.keys.emails, m.keys.quit}))
	}
	page := strings.Join(parts, "\n")
	return lipgloss.NewStyle().Padding(0, marginX).MaxWidth(m.width).Render(page)
}
