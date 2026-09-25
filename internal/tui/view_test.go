package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/usage"
)

var now = time.Date(2026, 9, 25, 2, 0, 0, 0, time.Local)

var ansi = regexp.MustCompile(`\x1b\[[0-9;:]*m`)

func sample() []usage.Report {
	win := func(name string, pct float64, in time.Duration) usage.Window {
		return usage.Window{Name: name, UsedPct: pct, ResetsAt: now.Add(in)}
	}
	return []usage.Report{
		{Provider: "claude", Account: "default", Email: "me@work.io", Plan: "team", Status: usage.StatusOK,
			Windows: []usage.Window{win(usage.Session, 6, 4*time.Hour), win(usage.Week, 33, 76*time.Hour)}},
		{Provider: "claude", Account: "personal", Email: "me@gmail.com", Plan: "pro", Status: usage.StatusOK,
			Windows: []usage.Window{win(usage.Session, 91, time.Hour), {Name: usage.Week, UsedPct: 62}}},
		{Provider: "claude", Account: "old", Status: usage.StatusExpired},
		{Provider: "claude", Account: "busy", Status: usage.StatusLimited},
		{Provider: "claude", Account: "slow", Status: usage.StatusOK, Stale: true, UpdatedAt: now.Add(-time.Hour),
			Windows: []usage.Window{win(usage.Session, 20, time.Hour), win(usage.Week, 50, 6*24*time.Hour)}},
		{Provider: "codex", Account: "default", Email: "me@x.io", Plan: "plus", Status: usage.StatusOK,
			Windows: []usage.Window{win(usage.Session, 0, 5*time.Hour), win(usage.Week, 9, 60*time.Hour)}},
	}
}

func loaded(t *testing.T, w, h int) Model {
	fetch := func(context.Context, config.Config) ([]usage.Report, time.Time, error) { return sample(), now, nil }
	m := New(fetch, Options{Every: 5 * time.Minute, Emails: true, Live: true, AltScreen: true})
	m.now = now
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	next, _ = next.Update(reportsMsg{sample(), now, nil})
	return next.(Model)
}

func plain(s string) string {
	return ansi.ReplaceAllString(s, "")
}

func TestFrameFitsScreen(t *testing.T) {
	for _, size := range [][2]int{{60, 20}, {120, 30}, {200, 50}} {
		m := loaded(t, size[0], size[1])
		for _, compact := range []bool{false, true} {
			m.compact = compact
			frame := m.frame()
			if got := lipgloss.Height(frame); got > size[1] {
				t.Errorf("%v compact=%v: height %d > %d", size, compact, got, size[1])
			}
			for _, line := range strings.Split(plain(frame), "\n") {
				if lipgloss.Width(line) > size[0] {
					t.Errorf("%v compact=%v: line too wide (%d): %q", size, compact, lipgloss.Width(line), line)
				}
			}
			if os.Getenv("UC_PREVIEW") != "" {
				t.Logf("\n%s", frame)
			}
		}
	}
}

func TestFirstViewShowsLoadingBeforeResize(t *testing.T) {
	fetch := func(context.Context, config.Config) ([]usage.Report, time.Time, error) { return nil, now, nil }
	for _, m := range []Model{New(fetch, Options{Live: true, AltScreen: true}), NewMini(fetch, Options{Live: true, AltScreen: true})} {
		if out := plain(m.View().Content); !strings.Contains(out, "fetching usage") {
			t.Errorf("initial view must not be blank: %q", out)
		}
		next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
		if out := plain(next.(Model).View().Content); !strings.Contains(out, "fetching usage") {
			t.Errorf("resized view must show loading: %q", out)
		}
	}
}

func TestFrameContent(t *testing.T) {
	out := plain(loaded(t, 120, 40).frame())
	for _, want := range []string{"◆ uc", "use first  claude › default  67% of 7d left · resets in 3d 4h", "codex › default  91% of 7d left", "default ★ use first",
		"me@work.io", "↻ 06:00 · in 4h", "Login expired. Run claude once", "rate limiting. Try again", "not started", "q quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("frame missing %q:\n%s", want, out)
		}
	}
}

func TestCompactShowsCountdown(t *testing.T) {
	m := loaded(t, 160, 40)
	m.compact = true
	out := plain(m.frame())
	for _, want := range []string{"06:00 · in 4h", "Mon 06:00 · in 3d 4h", "not started"} {
		if !strings.Contains(out, want) {
			t.Errorf("compact missing %q", want)
		}
	}
}

func TestOnlyPicksAreStarred(t *testing.T) {
	out := plain(loaded(t, 120, 40).frame())
	if n := strings.Count(out, "★ use first"); n != 3 {
		t.Errorf("want header plus 2 starred cards, got %d:\n%s", n, out)
	}
	if strings.Contains(out, "personal ★") || strings.Contains(out, "old ★") {
		t.Errorf("wrong account starred:\n%s", out)
	}
}

func TestKeys(t *testing.T) {
	m := loaded(t, 120, 40)
	press := func(m Model, k string) Model {
		code := []rune(k)[0]
		next, _ := m.Update(tea.KeyPressMsg{Code: code, Text: k})
		return next.(Model)
	}
	if m = press(m, "k"); m.selected != len(sample())-1 {
		t.Errorf("up from top should wrap to last, got %d", m.selected)
	}
	if m = press(m, "j"); m.selected != 0 {
		t.Errorf("down from last should wrap to first, got %d", m.selected)
	}
	if m = press(m, "c"); !m.compact {
		t.Error("c should toggle compact")
	}
	m = press(m, "r")
	if !m.loading {
		t.Error("r should start a refresh")
	}
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"}); cmd == nil {
		t.Error("q should quit")
	}
}

func TestAutoRefresh(t *testing.T) {
	m := loaded(t, 120, 40)
	next, _ := m.Update(tickMsg(now.Add(time.Minute)))
	if next.(Model).loading {
		t.Error("should not refresh before interval")
	}
	next, _ = m.Update(tickMsg(now.Add(5 * time.Minute)))
	if !next.(Model).loading {
		t.Error("should refresh after interval")
	}
	m.live = false
	next, _ = m.Update(tickMsg(now.Add(5 * time.Minute)))
	if next.(Model).loading {
		t.Error("live=false should disable automatic refresh")
	}
	next, _ = next.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if !next.(Model).loading {
		t.Error("r should still refresh")
	}
	m.altScreen = false
	if m.View().AltScreen {
		t.Error("full dashboard should allow normal screen")
	}
}

func TestEmailToggle(t *testing.T) {
	m := loaded(t, 160, 40)
	for _, compact := range []bool{false, true} {
		m.compact = compact
		if !strings.Contains(plain(m.frame()), "me@work.io") {
			t.Errorf("compact=%v: emails should show by default", compact)
		}
		next, _ := m.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
		if strings.Contains(plain(next.(Model).frame()), "me@work.io") {
			t.Errorf("compact=%v: e should hide emails", compact)
		}
	}
}

func TestReportsSanitizedBeforeRendering(t *testing.T) {
	m := New(nil, Options{Every: time.Minute})
	m = m.onReports(reportsMsg{reports: []usage.Report{{Provider: "claude", Account: "name\x1b[2J", Status: usage.StatusOK}}, at: now})
	if strings.Contains(m.frame(), "\x1b[2J") {
		t.Fatal("terminal escape was rendered")
	}
}

func TestCacheWarningStillShowsReports(t *testing.T) {
	m := New(nil, Options{Every: time.Minute})
	m = m.onReports(reportsMsg{reports: sample(), at: now, err: errors.New("usage cache was not saved")})
	if len(m.reports) == 0 || m.err != nil || !strings.Contains(m.frame(), "cache not saved") {
		t.Fatal("cache warning hid usage")
	}
}

func BenchmarkFirstScreen(b *testing.B) {
	fetch := func(context.Context, config.Config) ([]usage.Report, time.Time, error) { return nil, now, nil }
	b.ReportAllocs()
	for b.Loop() {
		m := New(fetch, Options{Every: 5 * time.Minute, Live: true, AltScreen: true})
		m.width, m.height = 120, 40
		_ = m.View()
	}
}

func BenchmarkLoadedScreen(b *testing.B) {
	fetch := func(context.Context, config.Config) ([]usage.Report, time.Time, error) { return sample(), now, nil }
	m := New(fetch, Options{Every: 5 * time.Minute, Live: true, AltScreen: true})
	m.width, m.height = 120, 40
	m.reports, m.fetchedAt, m.loading = sample(), now, false
	b.ReportAllocs()
	for b.Loop() {
		_ = m.View()
	}
}

func BenchmarkAccountScale(b *testing.B) {
	for _, count := range []int{1, 10, 50} {
		b.Run(fmt.Sprintf("accounts-%d", count), func(b *testing.B) {
			reports := make([]usage.Report, count)
			for i := range reports {
				reports[i] = sample()[i%3]
				reports[i].Account = fmt.Sprintf("account-%d", i)
			}
			m := New(nil, Options{Every: 5 * time.Minute, Live: true})
			m.width, m.height = 120, 40
			m.reports, m.fetchedAt, m.loading = reports, now, false
			b.ReportAllocs()
			for b.Loop() {
				_ = m.View()
			}
		})
	}
}
