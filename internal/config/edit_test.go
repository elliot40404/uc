package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/elliot40404/uc/internal/discover"
)

func TestMergeKeepsHidden(t *testing.T) {
	home := filepath.FromSlash("/h")
	c := Config{Accounts: []Entry{
		{Provider: "claude", Dir: filepath.Join(home, ".claude-old"), Hide: true},
		{Provider: "codex", Dir: filepath.FromSlash("/x/side"), Name: "side", Hide: true},
	}}
	found := []discover.Account{
		{Provider: "claude", Name: "default", Dir: filepath.Join(home, ".claude")},
		{Provider: "claude", Name: "old", Dir: filepath.Join(home, ".claude-old")},
	}
	got := c.Merge(found)
	if len(got) != 3 || got[0].Hide || !got[1].Hide || !got[2].Hide || got[2].Name != "side" {
		t.Fatalf("got %+v", got)
	}
	if len(c.Apply(found)) != 1 {
		t.Fatalf("apply should drop hidden: %+v", c.Apply(found))
	}
}

func TestEditAddsAndUpdatesEntries(t *testing.T) {
	dir := filepath.FromSlash("/h/.claude")
	orig := Config{Accounts: []Entry{{Provider: "codex", Dir: dir}}}
	c := orig.SetHidden("claude", dir, true)
	if len(c.Accounts) != 2 || !c.Accounts[1].Hide || len(orig.Accounts) != 1 {
		t.Fatalf("bad add: %+v, orig %+v", c.Accounts, orig.Accounts)
	}
	c = c.SetName("claude", dir, "  main  ")
	if len(c.Accounts) != 2 || c.Accounts[1].Name != "main" || !c.Accounts[1].Hide {
		t.Fatalf("bad rename: %+v", c.Accounts)
	}
	c = c.SetName("codex", dir, strings.Repeat("é", 40))
	if n := len([]rune(c.Accounts[0].Name)); n != MaxNameLen {
		t.Fatalf("name length %d", n)
	}
}

func TestEditDoesNotShareBacking(t *testing.T) {
	orig := Config{Accounts: make([]Entry, 1, 4)}
	orig.Accounts[0] = Entry{Provider: "codex", Dir: "/a"}
	_ = orig.SetHidden("codex", "/a", true)
	if orig.Accounts[0].Hide {
		t.Fatal("edit changed the original slice")
	}
}
