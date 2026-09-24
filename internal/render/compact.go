package render

import (
	"io"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func Compact(w io.Writer, reports []usage.Report, now time.Time, color bool) error {
	reports = usage.TerminalReports(reports)
	extras, emails := hasExtras(reports), hasEmails(reports)
	head := []string{"PROVIDER", "ACCOUNT"}
	if emails {
		head = append(head, "EMAIL")
	}
	head = append(head, "PLAN", "5H", "7D")
	if extras {
		head = append(head, "OTHER")
	}
	rows := [][]cell{headerRow(append(head, "STATUS"))}
	for _, r := range reports {
		rows = append(rows, compactRow(r, now, extras, emails))
	}
	return writeLines(w, grid(rows, color))
}

func headerRow(names []string) []cell {
	out := make([]cell, len(names))
	for i, n := range names {
		out[i] = text(bold, n)
	}
	return out
}

func compactRow(r usage.Report, now time.Time, extras, emails bool) []cell {
	row := []cell{text("", r.Provider), text("", r.Account)}
	if emails {
		row = append(row, text("", orDash(r.Email)))
	}
	row = append(row, text("", orDash(r.Plan)))
	row = append(row, compactWindow(r, usage.Session, now), compactWindow(r, usage.Week, now))
	if extras {
		row = append(row, extraCell(r))
	}
	return append(row, status(r))
}

func compactWindow(r usage.Report, name string, now time.Time) cell {
	w, ok := r.Window(name)
	if !ok {
		return text(dim, "-")
	}
	c := append(bar(w.UsedPct, 5, "▰", "▱"), seg{"", " "})
	c = append(c, percent(w.UsedPct)...)
	c = append(c, seg{"", "  "})
	return append(c, resets(w, now, "")...)
}

func extraCell(r usage.Report) cell {
	var c cell
	for i, w := range r.Extra() {
		if i > 0 {
			c = append(c, seg{"", ", "})
		}
		c = append(c, seg{"", w.Name + " "})
		c = append(c, percent(w.UsedPct)...)
	}
	if len(c) == 0 {
		return text(dim, "-")
	}
	return c
}

func hasEmails(reports []usage.Report) bool {
	for _, r := range reports {
		if r.Email != "" {
			return true
		}
	}
	return false
}

func hasExtras(reports []usage.Report) bool {
	for _, r := range reports {
		if len(r.Extra()) > 0 {
			return true
		}
	}
	return false
}

func writeLines(w io.Writer, lines []string) error {
	_, err := io.WriteString(w, strings.Join(lines, "\n")+"\n")
	return err
}
