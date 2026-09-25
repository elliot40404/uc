package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/usage"
)

func press(t *testing.T, m Model, keys ...tea.KeyPressMsg) Model {
	t.Helper()
	for _, k := range keys {
		next, _ := m.Update(k)
		m = next.(Model)
	}
	return m
}

var (
	keyS     = tea.KeyPressMsg{Code: 's', Text: "s"}
	keyDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	keySpace = tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	keyRight = tea.KeyPressMsg{Code: tea.KeyRight}
	keyLeft  = tea.KeyPressMsg{Code: tea.KeyLeft}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
)

func withSaver(m Model) (Model, *[]config.Config) {
	var saved []config.Config
	m.save = func(c config.Config) error {
		saved = append(saved, c)
		return nil
	}
	return m, &saved
}

func TestSettingsToggleAppliesAndSaves(t *testing.T) {
	m, saved := withSaver(loaded(t, 120, 40))
	m = press(t, m, keyS, keySpace, keyDown, keySpace, keyDown, keySpace, keyDown, keySpace, keyDown, keySpace)
	if m.live || !m.compact || m.emails || m.altScreen {
		t.Fatalf("session not updated: live=%v compact=%v emails=%v alt=%v", m.live, m.compact, m.emails, m.altScreen)
	}
	if len(*saved) != 5 {
		t.Fatalf("want 5 saves, got %d", len(*saved))
	}
	d := (*saved)[4].Defaults
	if d.Live || !d.Compact || d.ShowEmails || !d.Mini || d.AltScreen {
		t.Fatalf("bad saved defaults: %+v", d)
	}
	if m.View().AltScreen {
		t.Error("alt screen should turn off now")
	}
}

func TestSettingsEveryStepsAndClamps(t *testing.T) {
	m, saved := withSaver(loaded(t, 120, 40))
	m = press(t, m, keyS, keyDown, keyDown, keyDown, keyDown, keyDown, keyRight)
	if m.every != 10*time.Minute || (*saved)[0].Defaults.Refresh != 10*time.Minute {
		t.Fatalf("every = %v", m.every)
	}
	m = press(t, m, keyLeft, keyLeft, keyLeft, keyLeft, keyLeft)
	if m.every != time.Minute || len(*saved) != 4 {
		t.Fatalf("want clamp at 1m with 4 saves, got %v %d", m.every, len(*saved))
	}
}

func TestSettingsArrowsIgnoredOffEveryRow(t *testing.T) {
	m, saved := withSaver(loaded(t, 120, 40))
	m = press(t, m, keyS, keyRight)
	if m.every != 5*time.Minute || len(*saved) != 0 {
		t.Fatalf("arrow changed every on wrong row: %v", m.every)
	}
}

func TestSettingsCloseAndQuit(t *testing.T) {
	m := press(t, loaded(t, 120, 40), keyS, keyEsc)
	if m.settings || m.quitting {
		t.Fatalf("esc should close settings only: settings=%v quitting=%v", m.settings, m.quitting)
	}
	m = press(t, m, keyS)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if cmd == nil {
		t.Fatal("ctrl+c should quit from settings")
	}
}

func TestSettingsSaveError(t *testing.T) {
	m := loaded(t, 120, 40)
	m.save = func(config.Config) error { return errors.New("disk full") }
	m = press(t, m, keyS, keySpace)
	if m.saveErr == nil || m.live {
		t.Fatalf("want save error and applied change, got %v live=%v", m.saveErr, m.live)
	}
}

func TestStepEvery(t *testing.T) {
	cases := []struct {
		cur  time.Duration
		dir  int
		want time.Duration
	}{
		{3 * time.Minute, 1, 5 * time.Minute},
		{3 * time.Minute, -1, 2 * time.Minute},
		{30 * time.Minute, 1, 30 * time.Minute},
		{time.Hour, -1, 30 * time.Minute},
	}
	for _, c := range cases {
		if got := stepEvery(c.cur, c.dir); got != c.want {
			t.Errorf("stepEvery(%v, %d) = %v, want %v", c.cur, c.dir, got, c.want)
		}
	}
}

func TestSettingsViewFull(t *testing.T) {
	for _, size := range [][2]int{{60, 20}, {80, 24}} {
		m := press(t, loaded(t, size[0], size[1]), keyS)
		frame := m.frame()
		if lipgloss.Height(frame) > size[1] || lipgloss.Width(frame) > size[0] {
			t.Errorf("%v: frame %dx%d too big", size, lipgloss.Width(frame), lipgloss.Height(frame))
		}
		out := plain(frame)
		for _, want := range []string{"SETTINGS", "› [x] live refresh", "[ ] compact rows", "mini by default  next launch", "refresh every   ‹ 5m ›", "esc back"} {
			if !strings.Contains(out, want) {
				t.Errorf("%v: missing %q:\n%s", size, want, out)
			}
		}
		if strings.Contains(out, "personal") {
			t.Errorf("settings should replace account cards:\n%s", out)
		}
	}
}

func TestSettingsViewMiniAndError(t *testing.T) {
	fetch := func(context.Context, config.Config) ([]usage.Report, time.Time, error) { return sample(), now, nil }
	m := NewMini(fetch, Options{Every: 5 * time.Minute, Live: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = next.(Model)
	m.save = func(config.Config) error { return errors.New("disk full") }
	m = press(t, m, keyS, keySpace)
	out := plain(m.frame())
	for _, want := range []string{"SETTINGS", "[ ] live refresh", "not saved: disk full", "esc back"} {
		if !strings.Contains(out, want) {
			t.Errorf("mini settings missing %q:\n%s", want, out)
		}
	}
	if m = press(t, m, keyEsc); !strings.Contains(plain(m.frame()), "s settings") {
		t.Errorf("mini help should stay after live is off:\n%s", plain(m.frame()))
	}
}

func TestFetchGetsSavedConfig(t *testing.T) {
	var got config.Config
	m := New(func(_ context.Context, c config.Config) ([]usage.Report, time.Time, error) {
		got = c
		return nil, now, nil
	}, Options{Every: 5 * time.Minute, Live: true})
	m = press(t, m, keyS, keySpace)
	m.load()()
	if got.Defaults.Live {
		t.Fatalf("fetch should see saved config, got %+v", got.Defaults)
	}
}
