package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/usage"
)

const marginX = 1

func (m Model) View() tea.View {
	v := tea.NewView(m.frame())
	v.AltScreen = m.altScreen
	v.WindowTitle = "uc · usage"
	return v
}

func (m Model) frame() string {
	if m.width == 0 {
		return ""
	}
	if m.inline {
		return m.miniFrame()
	}
	w := m.width - 2*marginX
	header := m.header(w)
	footer := m.footer()
	bodyH := max(m.height-lipgloss.Height(header)-lipgloss.Height(footer)-1, 1)
	body := lipgloss.NewStyle().Height(bodyH).MaxHeight(bodyH).Render(m.body(w, bodyH))
	page := lipgloss.JoinVertical(lipgloss.Left, header, body, "", footer)
	return lipgloss.NewStyle().Padding(0, marginX).MaxWidth(m.width).Render(page)
}

func (m Model) footer() string {
	if m.settings {
		return m.settingsHelp()
	}
	return m.help.View(m.keys)
}

func (m Model) header(w int) string {
	return "\n" + m.headerCore(w) + "\n"
}

func (m Model) headerCore(w int) string {
	t := m.theme
	title := t.badge("◆ uc", t.accent) + t.fg(t.subtle).Render("  usage limits")
	lines := []string{spread(title, m.status(), w)}
	if best := m.bestLine(w); best != "" {
		lines = append(lines, best)
	}
	return strings.Join(lines, "\n")
}

func (m Model) status() string {
	t := m.theme
	switch {
	case m.loading:
		return m.spin.View() + t.fg(t.subtle).Render(" refreshing")
	case m.warning:
		return t.fg(t.warn).Render("cache not saved")
	case m.err != nil:
		return t.fg(t.bad).Render("● refresh failed")
	case m.fetchedAt.IsZero():
		return ""
	case !m.live:
		return t.fg(t.subtle).Render("updated " + m.fetchedAt.Format("15:04"))
	}
	next := max(m.fetchedAt.Add(m.every).Sub(m.now), 0)
	return t.fg(t.subtle).Render(fmt.Sprintf("updated %s · next in %s", m.fetchedAt.Format("15:04"), countdown(next)))
}

func (m Model) bestLine(w int) string {
	t := m.theme
	var parts []string
	for _, p := range m.picks() {
		parts = append(parts, m.pickText(p))
	}
	if len(parts) == 0 {
		return ""
	}
	label := t.fg(t.accent).Render("★ ") + t.fg(t.subtle).Render("use first  ")
	line := label + strings.Join(parts, t.fg(t.faint).Render("   │   "))
	if lipgloss.Width(line) <= w {
		return line
	}
	return label + strings.Join(parts, "\n"+strings.Repeat(" ", lipgloss.Width(label)))
}

func (m Model) body(w, h int) string {
	t := m.theme
	switch {
	case m.settings:
		return strings.Join(m.settingsLines(), "\n")
	case m.err != nil && len(m.reports) == 0:
		return t.fg(t.bad).Width(w).Render("Could not load accounts: " + usage.TerminalText(m.err.Error()) + "\nPress r to retry.")
	case m.loading && len(m.reports) == 0:
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, m.spin.View()+t.fg(t.subtle).Render(" fetching usage…"))
	case len(m.reports) == 0:
		return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, t.fg(t.subtle).Render("No accounts found. Log in with claude or codex first."))
	case m.compact:
		return fit(m.rows(w), m.selected, h)
	}
	blocks, sel := m.grid(w)
	return fit(blocks, sel, h)
}

func countdown(d time.Duration) string {
	secs := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%dm %02ds", secs/60, secs%60)
}
