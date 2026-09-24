package tui

import (
	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/usage"
)

func (m Model) picks() []usage.Pick {
	var out []usage.Pick
	for _, provider := range config.Providers {
		if p, ok := usage.Best(m.reports, provider, m.now); ok {
			out = append(out, p)
		}
	}
	return out
}

func (m Model) picked() map[int]bool {
	out := map[int]bool{}
	for _, p := range m.picks() {
		for i, r := range m.reports {
			if sameAccount(r, p.Report) {
				out[i] = true
			}
		}
	}
	return out
}

func (m Model) pickText(p usage.Pick) string {
	t := m.theme
	r := p.Report
	s := t.fg(t.provider(r.Provider)).Bold(true).Render(r.Provider) + t.fg(t.subtle).Render(" › ") +
		t.fg(t.text).Bold(true).Render(r.Account) + "  " +
		t.fg(t.level(100-p.WeekLeft)).Render(render.WeekLeft(p)) +
		t.fg(t.subtle).Render(" · "+render.WeekResets(p, m.now))
	if u := render.UsableIn(p, m.now); u != "" {
		s += t.fg(t.warn).Render(" · " + u)
	}
	return s
}

func (t theme) star(on bool) string {
	if !on {
		return "  "
	}
	return t.fg(t.accent).Render("★ ")
}

func sameAccount(a, b usage.Report) bool {
	return a.Provider == b.Provider && a.Account == b.Account && a.Dir == b.Dir
}
