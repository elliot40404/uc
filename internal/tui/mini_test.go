package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/elliot40404/uc/internal/usage"
)

func TestPrintMini(t *testing.T) {
	var b strings.Builder
	if err := Print(&b, sample(), now, 160, true, Options{Emails: true}); err != nil {
		t.Fatal(err)
	}
	out := plain(b.String())
	for _, want := range []string{"◆ uc", "updated 02:00", "use first", "default", "06:00 · in 4h"} {
		if !strings.Contains(out, want) {
			t.Errorf("mini missing %q:\n%s", want, out)
		}
	}
	for _, bad := range []string{"next in", "quit", "▌", "╭"} {
		if strings.Contains(out, bad) {
			t.Errorf("mini should not contain %q:\n%s", bad, out)
		}
	}
}

func TestLiveMiniScreenAndDropsHelpOnQuit(t *testing.T) {
	fetch := func(context.Context) ([]usage.Report, time.Time, error) { return sample(), now, nil }
	m := NewMini(fetch, Options{Every: 5 * time.Minute, Emails: true, Live: true, AltScreen: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 160, Height: 40})
	next, _ = next.Update(reportsMsg{sample(), now, nil})
	m = next.(Model)
	if !m.View().AltScreen {
		t.Error("mini should use the alt screen by default")
	}
	if out := plain(m.frame()); !strings.Contains(out, "q quit") || !strings.Contains(out, "next in") {
		t.Errorf("live mini should show help and countdown:\n%s", out)
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if out := plain(next.(Model).frame()); strings.Contains(out, "q quit") {
		t.Errorf("help should be gone after quit:\n%s", out)
	}
	if NewMini(fetch, Options{AltScreen: false}).View().AltScreen {
		t.Error("mini should allow the normal screen")
	}
}

func TestMiniRowExplainsProblem(t *testing.T) {
	var b strings.Builder
	Print(&b, sample(), now, 200, true, Options{})
	out := plain(b.String())
	for _, want := range []string{"● rate limited  Usage server is rate limiting", "● expired  Login expired"} {
		if !strings.Contains(out, want) {
			t.Errorf("mini missing %q:\n%s", want, out)
		}
	}
}

func TestStaleRowsShowCachedTime(t *testing.T) {
	var b strings.Builder
	Print(&b, sample(), now, 220, true, Options{})
	out := plain(b.String())
	if !strings.Contains(out, "● cached 01:00") || !strings.Contains(out, "50%") {
		t.Errorf("stale row should show numbers and cached time:\n%s", out)
	}
}
