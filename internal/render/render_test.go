package render

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

var now = time.Date(2026, 9, 25, 2, 0, 0, 0, time.UTC)

func init() {
	time.Local = time.UTC
}

func TestUntil(t *testing.T) {
	cases := map[time.Duration]string{
		-time.Hour:                   "now",
		45 * time.Minute:             "in 45m",
		5 * time.Hour:                "in 5h",
		4*time.Hour + 15*time.Minute: "in 4h 15m",
		48 * time.Hour:               "in 2d",
		76 * time.Hour:               "in 3d 4h",
	}
	for d, want := range cases {
		if got := Until(now.Add(d), now); got != want {
			t.Errorf("Until(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestClock(t *testing.T) {
	cases := map[time.Duration]string{
		4*time.Hour + 25*time.Minute + 40*time.Second: "06:26",
		3*24*time.Hour + 4*time.Hour:                  "Mon 06:00",
		10 * 24 * time.Hour:                           "Mon 5 Oct 02:00",
	}
	for d, want := range cases {
		if got := Clock(now.Add(d), now); got != want {
			t.Errorf("Clock(+%v) = %q, want %q", d, got, want)
		}
	}
}

func TestBar(t *testing.T) {
	cases := map[float64]string{0: "░░░░", 1: "█░░░", 50: "██░░", 100: "████", 150: "████"}
	for pct, want := range cases {
		if got := bar(pct, 4, "█", "░").render(false); got != want {
			t.Errorf("bar(%v) = %q, want %q", pct, got, want)
		}
	}
}

func sample() []usage.Report {
	return []usage.Report{
		{Provider: "claude", Account: "personal", Email: "me@x.io", Plan: "pro", Status: usage.StatusOK, Windows: []usage.Window{
			{Name: usage.Session, UsedPct: 42, ResetsAt: now.Add(90 * time.Minute)},
			{Name: usage.Week, UsedPct: 7},
			{Name: "sonnet", UsedPct: 12},
		}},
		{Provider: "claude", Account: "work", Status: usage.StatusExpired},
		{Provider: "codex", Account: "default", Plan: "plus", Status: usage.StatusOK, Windows: []usage.Window{
			{Name: usage.Session, UsedPct: 0, ResetsAt: now.Add(5 * time.Hour)},
		}},
	}
}

func render(t *testing.T, f func(*strings.Builder) error) []string {
	var b strings.Builder
	if err := f(&b); err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimRight(b.String(), "\n"), "\n")
}

func TestTextViewsDoNotEmitInputControlCharacters(t *testing.T) {
	reports := []usage.Report{{Provider: "claude", Account: "name\x1b[2J\nline", Plan: "pro\rplan", Status: usage.StatusError, Error: "bad\x1b[31m"}}
	for _, view := range []struct {
		name string
		fn   func(io.Writer, []usage.Report, time.Time, bool) error
	}{{"compact", Compact}, {"grouped", Grouped}} {
		t.Run(view.name, func(t *testing.T) {
			var out strings.Builder
			if err := view.fn(&out, reports, now, false); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(out.String(), "\x1b") || strings.Contains(out.String(), "name\nline") {
				t.Fatalf("unsanitized output %q", out.String())
			}
		})
	}
}

func TestCompact(t *testing.T) {
	lines := render(t, func(b *strings.Builder) error { return Compact(b, sample(), now, false) })
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got:\n%s", strings.Join(lines, "\n"))
	}
	for _, want := range []string{"personal  me@x.io  pro", "▰▰▱▱▱  42%  03:30 (in 1h 30m)", "▰▱▱▱▱   7%  not started", "sonnet  12%", "ok"} {
		if !strings.Contains(lines[1], want) {
			t.Errorf("row missing %q: %q", want, lines[1])
		}
	}
	if !strings.Contains(lines[2], "expired, run claude once") {
		t.Errorf("bad expired row: %q", lines[2])
	}
	if strings.Index(lines[0], "5H") != strings.Index(lines[1], "▰") {
		t.Errorf("columns not aligned:\n%s", strings.Join(lines, "\n"))
	}
}

func TestGrouped(t *testing.T) {
	lines := render(t, func(b *strings.Builder) error { return Grouped(b, sample(), now, false) })
	want := []string{
		"CLAUDE",
		"  personal  me@x.io  pro   ok",
		"    5h      ████████░░░░░░░░░░░░   42%  resets 03:30 (in 1h 30m)",
		"    7d      █░░░░░░░░░░░░░░░░░░░    7%  not started",
		"    sonnet  ██░░░░░░░░░░░░░░░░░░   12%  not started",
		"  work      -        -     expired, run claude once in this account",
		"",
		"CODEX",
		"  default   -        plus  ok",
		"    5h      ░░░░░░░░░░░░░░░░░░░░    0%  resets 07:00 (in 5h)",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

func TestColorWrapsSegments(t *testing.T) {
	lines := render(t, func(b *strings.Builder) error { return Grouped(b, sample()[2:], now, true) })
	if !strings.Contains(lines[2], green+"░") && !strings.Contains(lines[2], dim+"░") {
		t.Errorf("want colored bar, got %q", lines[2])
	}
}

func TestNoEmailColumnWhenHidden(t *testing.T) {
	reports := usage.WithoutEmails(sample())
	compact := strings.Join(render(t, func(b *strings.Builder) error { return Compact(b, reports, now, false) }), "\n")
	grouped := strings.Join(render(t, func(b *strings.Builder) error { return Grouped(b, reports, now, false) }), "\n")
	if strings.Contains(compact, "EMAIL") || strings.Contains(compact, "me@x.io") {
		t.Errorf("compact should drop email column:\n%s", compact)
	}
	if !strings.Contains(grouped, "  personal  pro   ok") {
		t.Errorf("grouped should drop email column:\n%s", grouped)
	}
}
