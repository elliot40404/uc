package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/usage"
)

const (
	minCard  = 48
	gridGap  = 2
	clockCol = 22
)

func columns(w int) int {
	return max(min((w+gridGap)/(minCard+gridGap), 3), 1)
}

func (m Model) grid(w int) ([]string, int) {
	cols := columns(w)
	picked := m.picked()
	cardW := (w - (cols-1)*gridGap) / cols
	var blocks []string
	for start := 0; start < len(m.reports); start += cols {
		end := min(start+cols, len(m.reports))
		blocks = append(blocks, m.gridRow(start, end, cardW, picked))
	}
	return blocks, m.selected / cols
}

func (m Model) gridRow(start, end, cardW int, picked map[int]bool) string {
	tallest := 0
	cards := make([]string, end-start)
	heights := make([]int, end-start)
	for i := start; i < end; i++ {
		card := m.theme.card(m.reports[i], cardW, 0, i == m.selected, picked[i], m.emails, m.now)
		cards[i-start] = card
		heights[i-start] = lipgloss.Height(card)
		tallest = max(tallest, heights[i-start])
	}
	parts := make([]string, 0, 2*(end-start))
	for i := start; i < end; i++ {
		if i > start {
			parts = append(parts, strings.Repeat(" ", gridGap))
		}
		card := cards[i-start]
		if heights[i-start] < tallest {
			card = m.theme.card(m.reports[i], cardW, tallest, i == m.selected, picked[i], m.emails, m.now)
		}
		parts = append(parts, card)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...) + "\n"
}

type rowLayout struct {
	nameW  int
	emailW int
	barW   int
	clocks bool
}

const (
	rowBase   = 2 + 7 + 2 + 2 + 6
	winBase   = 4 + 5 + 2
	minMini   = 4
	maxMini   = 20
	clockPart = clockCol + 2
)

func (m Model) layoutRows(w int) rowLayout {
	l := rowLayout{clocks: true}
	for _, r := range m.reports {
		l.nameW = max(l.nameW, lipgloss.Width(r.Account))
		l.emailW = max(l.emailW, lipgloss.Width(r.Email)+2)
	}
	room := func() int { return (w - rowBase - l.nameW - l.emailW - 2*winBase - 2*clockPart) / 2 }
	if !m.emails || room() < minMini+4 {
		l.emailW = 0
	}
	freed := 0
	if room() < minMini {
		l.clocks, freed = false, clockPart
	}
	l.barW = min(max(room()+freed, minMini), maxMini)
	return l
}

func (m Model) rows(w int) []string {
	l := m.layoutRows(w)
	picked := m.picked()
	out := make([]string, len(m.reports))
	for i, r := range m.reports {
		out[i] = m.row(r, i == m.selected, picked[i], l)
	}
	return out
}

func (m Model) row(r usage.Report, selected, picked bool, l rowLayout) string {
	t := m.theme
	marker := "  "
	name := t.fg(t.text)
	if selected {
		marker = t.fg(t.provider(r.Provider)).Render("▌ ")
		name = name.Bold(true)
	}
	parts := []string{
		marker + t.fg(t.provider(r.Provider)).Width(7).Render(r.Provider) + t.star(picked),
		name.Width(l.nameW + 2).Render(r.Account),
	}
	if l.emailW > 0 {
		parts = append(parts, t.fg(t.subtle).Width(l.emailW).Render(orDash(r.Email)))
	}
	parts = append(parts, t.fg(t.subtle).Width(6).Render(orDash(r.Plan)))
	if r.Status != usage.StatusOK {
		return strings.Join(append(parts, t.statusDot(r), t.fg(t.subtle).Render("  "+problem(r))), "")
	}
	parts = append(parts, m.miniWindow(r, usage.Session, l), m.miniWindow(r, usage.Week, l))
	if r.Stale {
		parts = append(parts, t.statusDot(r))
	}
	return strings.Join(parts, "")
}

func (m Model) miniWindow(r usage.Report, name string, l rowLayout) string {
	t := m.theme
	w, ok := r.Window(name)
	label := t.fg(t.subtle).Render(name + "  ")
	if !ok {
		return label + t.fg(t.faint).Width(l.barW+winBase-4).Render("-")
	}
	pct := t.fg(t.level(w.UsedPct)).Bold(true).Width(5).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%%", w.UsedPct))
	out := label + t.bar(w.UsedPct, l.barW) + pct + "  "
	if l.clocks {
		out += lipgloss.NewStyle().Width(clockPart).Render(m.resetShort(w))
	}
	return out
}

func fit(blocks []string, sel, height int) string {
	if len(blocks) == 0 {
		return ""
	}
	sel = min(max(sel, 0), len(blocks)-1)
	start := 0
	for start < sel && totalHeight(blocks[start:sel+1]) > height {
		start++
	}
	var out []string
	used := 0
	for _, b := range blocks[start:] {
		h := lipgloss.Height(b)
		if used+h > height && len(out) > 0 {
			break
		}
		out = append(out, b)
		used += h
	}
	return strings.Join(out, "\n")
}

func totalHeight(blocks []string) int {
	n := 0
	for _, b := range blocks {
		n += lipgloss.Height(b)
	}
	return n
}

func (m Model) resetShort(w usage.Window) string {
	t := m.theme
	if w.ResetsAt.IsZero() {
		return t.fg(t.faint).Render("not started")
	}
	return t.fg(t.text).Render(render.Clock(w.ResetsAt, m.now)) + t.fg(t.subtle).Render(" · "+render.Until(w.ResetsAt, m.now))
}
