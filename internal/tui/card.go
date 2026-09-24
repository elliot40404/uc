package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/usage"
)

const (
	cardPadX   = 2
	labelWidth = 7
	pctWidth   = 5
)

func (t theme) card(r usage.Report, width, height int, selected, picked, email bool, now time.Time) string {
	inner := width - 2 - 2*cardPadX
	lines := []string{t.cardTitle(r, inner, picked)}
	if email {
		lines = append(lines, t.fg(t.subtle).Render(orDash(r.Email)))
	}
	lines = append(lines, "")
	lines = append(lines, t.cardBody(r, inner, now)...)
	border := t.faint
	if selected {
		border = t.provider(r.Provider)
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(0, cardPadX).
		Width(width).
		Height(height).
		Render(strings.Join(lines, "\n"))
}

func (t theme) cardTitle(r usage.Report, inner int, picked bool) string {
	left := t.badge(strings.ToUpper(r.Provider), t.provider(r.Provider)) + " " +
		t.fg(t.text).Bold(true).Render(r.Account)
	if picked {
		left += " " + t.fg(t.accent).Render("★ use first")
	}
	right := t.statusDot(r)
	if r.Plan != "" {
		right = t.fg(t.subtle).Render(r.Plan) + "  " + right
	}
	return spread(left, right, inner)
}

func (t theme) statusDot(r usage.Report) string {
	switch {
	case r.Stale:
		return t.fg(t.warn).Render("● cached " + r.UpdatedAt.Local().Format("15:04"))
	case r.Status == usage.StatusOK:
		return t.fg(t.good).Render("● ok")
	case r.Status == usage.StatusError:
		return t.fg(t.bad).Render("● error")
	default:
		return t.fg(t.warn).Render("● " + string(r.Status))
	}
}

func (t theme) cardBody(r usage.Report, inner int, now time.Time) []string {
	switch r.Status {
	case usage.StatusOK:
		var lines []string
		for _, w := range r.Windows {
			lines = append(lines, t.windowLine(w, inner), t.resetLine(w, now))
		}
		return lines
	case usage.StatusError:
		return []string{t.fg(t.bad).Width(inner).Render(problem(r))}
	default:
		return []string{t.fg(t.warn).Width(inner).Render(problem(r))}
	}
}

func problem(r usage.Report) string {
	switch r.Status {
	case usage.StatusExpired:
		return fmt.Sprintf("Login expired. Run %s once in this account.", r.Provider)
	case usage.StatusNoLogin:
		return fmt.Sprintf("Not logged in. Run %s to log in.", r.Provider)
	case usage.StatusLimited:
		return "Usage server is rate limiting. Try again in a few minutes."
	default:
		return r.Error
	}
}

func (t theme) windowLine(w usage.Window, inner int) string {
	label := t.fg(t.text).Bold(true).Width(labelWidth).Render(w.Name)
	pct := t.fg(t.level(w.UsedPct)).Bold(true).Width(pctWidth).Align(lipgloss.Right).Render(fmt.Sprintf("%.0f%%", w.UsedPct))
	return label + t.bar(w.UsedPct, inner-labelWidth-pctWidth-1) + " " + pct
}

func (t theme) resetLine(w usage.Window, now time.Time) string {
	pad := strings.Repeat(" ", labelWidth)
	if w.ResetsAt.IsZero() {
		return pad + t.fg(t.subtle).Render("not started")
	}
	return pad + t.fg(t.subtle).Render("↻ ") + t.fg(t.text).Render(render.Clock(w.ResetsAt, now)) +
		t.fg(t.subtle).Render(" · "+render.Until(w.ResetsAt, now))
}

func spread(left, right string, width int) string {
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + strings.Repeat(" ", gap) + right
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
