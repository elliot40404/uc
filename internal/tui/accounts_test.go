package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/elliot40404/uc/internal/discover"
)

var keyH = tea.KeyPressMsg{Code: 'h', Text: "h"}
var keyUp = tea.KeyPressMsg{Code: tea.KeyUp}

func withAccounts(m Model) Model {
	m.home = filepath.FromSlash("/h")
	m.found = []discover.Account{
		{Provider: "claude", Name: "default", Dir: filepath.FromSlash("/h/.claude")},
		{Provider: "codex", Name: "work", Dir: filepath.FromSlash("/h/.codex-work")},
	}
	return m
}

func TestHideAccountSavesAndRefetches(t *testing.T) {
	m, saved := withSaver(withAccounts(loaded(t, 120, 40)))
	m = press(t, m, keyS, keyUp)
	next, cmd := m.Update(keyH)
	m = next.(Model)
	if len(*saved) != 1 || !(*saved)[0].Accounts[0].Hide || (*saved)[0].Accounts[0].Provider != "codex" {
		t.Fatalf("bad save: %+v", *saved)
	}
	if cmd == nil || !m.loading {
		t.Fatal("hide should start a refetch")
	}
	if out := plain(m.frame()); !strings.Contains(out, "~/.codex-work  hidden") || !strings.Contains(out, "h hide/unhide") {
		t.Errorf("missing hidden row or help:\n%s", out)
	}
	m = press(t, m, keyH)
	if (*saved)[1].Accounts[0].Hide || len((*saved)[1].Accounts) != 1 {
		t.Fatalf("unhide should reuse entry: %+v", (*saved)[1].Accounts)
	}
}

func TestHideWhileLoadingReloadsAfter(t *testing.T) {
	m, _ := withSaver(withAccounts(loaded(t, 120, 40)))
	m.loading = true
	m = press(t, m, keyS, keyUp, keyH)
	if !m.reload {
		t.Fatal("want reload queued")
	}
	next, cmd := m.Update(reportsMsg{sample(), now, nil})
	m = next.(Model)
	if m.reload || !m.loading || cmd == nil {
		t.Fatalf("want queued reload to start: reload=%v loading=%v", m.reload, m.loading)
	}
}

func TestSpaceOnAccountRowDoesNothing(t *testing.T) {
	m, saved := withSaver(withAccounts(loaded(t, 120, 40)))
	m = press(t, m, keyS, keyUp, keySpace)
	if len(*saved) != 0 {
		t.Fatalf("space on account row saved: %+v", *saved)
	}
}

func TestSettingsScrollsToSelectedAccount(t *testing.T) {
	m := withAccounts(loaded(t, 80, 14))
	m = press(t, m, keyS, keyUp)
	out := plain(m.frame())
	if !strings.Contains(out, "› codex") {
		t.Errorf("selected account should be visible:\n%s", out)
	}
}
