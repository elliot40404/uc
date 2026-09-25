package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

var (
	keyN     = tea.KeyPressMsg{Code: 'n', Text: "n"}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyClear = tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl}
)

func typed(s string) []tea.KeyPressMsg {
	var out []tea.KeyPressMsg
	for _, r := range s {
		out = append(out, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return out
}

func TestRenameSavesTypedName(t *testing.T) {
	m, saved := withSaver(withAccounts(loaded(t, 120, 40)))
	m = press(t, m, keyS, keyUp, keyN)
	if !m.renaming || m.input.Value() != "work" {
		t.Fatalf("want rename box with current name, got %v %q", m.renaming, m.input.Value())
	}
	if out := plain(m.frame()); !strings.Contains(out, "enter save") {
		t.Errorf("missing rename help:\n%s", out)
	}
	m = press(t, m, keyClear)
	m = press(t, m, typed("qsh job")...)
	if m.quitting || !m.settings || len(*saved) != 0 {
		t.Fatalf("typing triggered other keys: quitting=%v settings=%v saves=%d", m.quitting, m.settings, len(*saved))
	}
	next, cmd := m.Update(keyEnter)
	m = next.(Model)
	if m.renaming || cmd == nil || len(*saved) != 1 || (*saved)[0].Accounts[0].Name != "qsh job" {
		t.Fatalf("bad rename save: renaming=%v %+v", m.renaming, *saved)
	}
	if out := plain(m.frame()); !strings.Contains(out, "qsh job") {
		t.Errorf("new name not shown:\n%s", out)
	}
}

func TestRenameEscCancels(t *testing.T) {
	m, saved := withSaver(withAccounts(loaded(t, 120, 40)))
	m = press(t, m, keyS, keyUp, keyN, keyClear)
	m = press(t, m, typed("x")...)
	m = press(t, m, keyEsc)
	if m.renaming || !m.settings || len(*saved) != 0 {
		t.Fatalf("esc should cancel only: renaming=%v settings=%v saves=%d", m.renaming, m.settings, len(*saved))
	}
}

func TestRenameEmptyResetsName(t *testing.T) {
	m, saved := withSaver(withAccounts(loaded(t, 120, 40)))
	m = press(t, m, keyS, keyUp, keyN, keyClear, keyEnter)
	if len(*saved) != 1 || (*saved)[0].Accounts[0].Name != "" {
		t.Fatalf("empty name should clear override: %+v", *saved)
	}
}

func TestRenameOnlyOnAccountRows(t *testing.T) {
	m := press(t, withAccounts(loaded(t, 120, 40)), keyS, keyN)
	if m.renaming {
		t.Fatal("n on a setting row should do nothing")
	}
}
