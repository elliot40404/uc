package tray

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/elliot40404/uc/internal/usage"
)

const maxTip = 127

func Providers(reports []usage.Report) []string {
	var out []string
	for _, r := range reports {
		if !slices.Contains(out, r.Provider) {
			out = append(out, r.Provider)
		}
	}
	return out
}

func Picks(reports []usage.Report, now time.Time) map[string]usage.Pick {
	out := map[string]usage.Pick{}
	for _, p := range Providers(reports) {
		if pick, ok := usage.Best(reports, p, now); ok {
			out[p] = pick
		}
	}
	return out
}

func Line(r usage.Report, picked bool) string {
	name := r.Provider + " " + r.Account
	if picked {
		name = "★ " + name
	}
	return menuText(name) + "\t" + menuText(state(r))
}

func Tooltip(reports []usage.Report, picks map[string]usage.Pick) string {
	lines := []string{"uc"}
	for _, p := range Providers(reports) {
		line := p + ": nothing available"
		if pick, ok := picks[p]; ok {
			line = pick.Report.Provider + " " + pick.Report.Account + ": " + windows(pick.Report)
		}
		lines = append(lines, usage.TerminalText(line))
	}
	return clip(strings.Join(lines, "\n"), maxTip)
}

func IsPick(r usage.Report, picks map[string]usage.Pick) bool {
	p, ok := picks[r.Provider]
	return ok && p.Report.Dir == r.Dir
}

func state(r usage.Report) string {
	if r.Status != usage.StatusOK {
		return string(r.Status)
	}
	s := windows(r)
	if r.Stale {
		s += " (cached)"
	}
	return s
}

func windows(r usage.Report) string {
	parts := make([]string, len(r.Windows))
	for i, w := range r.Windows {
		parts[i] = fmt.Sprintf("%s %.0f%%", w.Name, w.UsedPct)
	}
	return strings.Join(parts, "  ")
}

func menuText(s string) string {
	return strings.ReplaceAll(usage.TerminalText(s), "&", "&&")
}

func clip(s string, n int) string {
	r := []rune(s)
	if utf16Len(r) <= n {
		return s
	}
	for utf16Len(r) > n-1 {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

func utf16Len(r []rune) int {
	return len(utf16.Encode(r))
}
