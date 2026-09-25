package tray

import (
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/elliot40404/uc/internal/usage"
)

var now = time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

func report(provider, account string, session, week float64) usage.Report {
	return usage.Report{
		Provider: provider,
		Account:  account,
		Dir:      "/" + provider + "/" + account,
		Status:   usage.StatusOK,
		Windows: []usage.Window{
			{Name: usage.Session, UsedPct: session},
			{Name: usage.Week, UsedPct: week, ResetsAt: now.Add(72 * time.Hour)},
		},
	}
}

func TestTooltipShowsBestPerProvider(t *testing.T) {
	reports := []usage.Report{report("claude", "a", 10, 90), report("claude", "b", 20, 10), report("codex", "c", 0, 100)}
	got := Tooltip(reports, Picks(reports, now))
	want := "uc\nclaude b: 5h 20%  7d 10%\ncodex: nothing available"
	if got != want {
		t.Errorf("got %q", got)
	}
	if !IsPick(reports[1], Picks(reports, now)) || IsPick(reports[0], Picks(reports, now)) {
		t.Error("wrong pick")
	}
}

func TestTooltipFitsWindowsLimit(t *testing.T) {
	r := report("claude", strings.Repeat("😀", 100), 1, 1)
	got := Tooltip([]usage.Report{r}, Picks([]usage.Report{r}, now))
	if n := len(utf16.Encode([]rune(got))); n > maxTip || !strings.HasSuffix(got, "…") {
		t.Errorf("len %d, %q", n, got)
	}
}

func TestTipIsOneLine(t *testing.T) {
	if got := Tip("bad\nconfig"); got != "uc\nbad config" {
		t.Errorf("got %q", got)
	}
}
