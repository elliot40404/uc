package render

import (
	"io"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

const barWidth = 20

func Grouped(w io.Writer, reports []usage.Report, now time.Time, color bool) error {
	reports = usage.TerminalReports(reports)
	accounts := make([][]cell, len(reports))
	emails := hasEmails(reports)
	var windows [][]cell
	for i, r := range reports {
		accounts[i] = []cell{text(bold, "  "+r.Account)}
		if emails {
			accounts[i] = append(accounts[i], text("", orDash(r.Email)))
		}
		accounts[i] = append(accounts[i], text(dim, orDash(r.Plan)), status(r))
		for _, win := range shown(r) {
			windows = append(windows, windowRow(win, now))
		}
	}
	return writeLines(w, interleave(reports, grid(accounts, color), grid(windows, color), color))
}

func shown(r usage.Report) []usage.Window {
	if r.Status != usage.StatusOK {
		return nil
	}
	return r.Windows
}

func windowRow(w usage.Window, now time.Time) []cell {
	return []cell{
		text("", "    "+w.Name),
		bar(w.UsedPct, barWidth, "█", "░"),
		percent(w.UsedPct),
		resets(w, now, "resets "),
	}
}

func interleave(reports []usage.Report, accounts, windows []string, color bool) []string {
	var out []string
	provider := ""
	for i, r := range reports {
		if r.Provider != provider {
			if provider != "" {
				out = append(out, "")
			}
			provider = r.Provider
			out = append(out, text(bold, strings.ToUpper(provider)).render(color))
		}
		out = append(out, accounts[i])
		n := len(shown(r))
		out = append(out, windows[:n]...)
		windows = windows[n:]
	}
	return out
}
