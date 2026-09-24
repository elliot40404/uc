package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/discover"
)

func writeConfig(t *testing.T, body string) string {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMissingIsEmpty(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.json"), "/h")
	if err != nil || len(c.Accounts) != 0 || !c.Defaults.Live || !c.Defaults.AltScreen {
		t.Fatalf("got %+v, %v", c, err)
	}
}

func TestLoadErrors(t *testing.T) {
	cases := map[string]string{
		`{"accounts":[{"provider":"gemini","dir":"x"}]}`: "provider must be",
		`{"accounts":[{"provider":"claude"}]}`:           "dir is required",
		`{"acounts":[]}`:                                 "unknown field",
		`{`:                                              "unexpected EOF",
	}
	for body, want := range cases {
		_, err := Load(writeConfig(t, body), "/h")
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("body %s: got %v, want %q", body, err, want)
		}
	}
}

func TestDefaults(t *testing.T) {
	c, err := Load(writeConfig(t, `{"defaults":{"mini":true,"every":"2m"},"accounts":[]}`), "/h")
	if err != nil || !c.Defaults.Mini || !c.Defaults.Live || !c.Defaults.AltScreen || c.Defaults.Refresh != 2*time.Minute {
		t.Fatalf("got %+v, %v", c.Defaults, err)
	}
	c, err = Load(writeConfig(t, `{"defaults":{"live":false,"alt_screen":false}}`), "/h")
	if err != nil || c.Defaults.Live || c.Defaults.AltScreen {
		t.Fatalf("explicit false defaults: %+v, %v", c.Defaults, err)
	}
	c, _ = Load(filepath.Join(t.TempDir(), "none.json"), "/h")
	if c.Defaults.Refresh != DefaultRefresh {
		t.Errorf("missing config refresh = %v", c.Defaults.Refresh)
	}
	if _, err := Load(writeConfig(t, `{"defaults":{"every":"soon"}}`), "/h"); err == nil || !strings.Contains(err.Error(), "defaults.every") {
		t.Errorf("want every error, got %v", err)
	}
}

func TestApply(t *testing.T) {
	home := filepath.FromSlash("/h")
	c, err := Load(writeConfig(t, `{"accounts":[
		{"provider":"claude","dir":"~/.claude-backup","hide":true},
		{"provider":"claude","dir":"~/.claude-work","name":"job"},
		{"provider":"codex","dir":"~/.claude-work","name":"wrong provider"},
		{"provider":"codex","dir":"/x/codex2"}
	]}`), home)
	if err != nil {
		t.Fatal(err)
	}
	found := []discover.Account{
		{Provider: "claude", Name: "backup", Dir: filepath.Join(home, ".claude-backup")},
		{Provider: "claude", Name: "work", Dir: filepath.Join(home, ".claude-work")},
	}
	got := c.Apply(found)
	want := []string{"claude/job", "codex/wrong provider", "codex/codex2"}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i, a := range got {
		if a.Provider+"/"+a.Name != want[i] {
			t.Errorf("got %+v, want %s", a, want[i])
		}
	}
}

func TestDuplicateAccountsRejected(t *testing.T) {
	entries := `[ {"provider":"codex","dir":"~/account"}, {"provider":"codex","dir":"~/account"} ]`
	if _, err := Load(writeConfig(t, `{"accounts":`+entries+`}`), t.TempDir()); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("want duplicate error, got %v", err)
	}
	if runtime.GOOS == "windows" {
		entries = `[ {"provider":"codex","dir":"~/Account"}, {"provider":"codex","dir":"~/account"} ]`
		if _, err := Load(writeConfig(t, `{"accounts":`+entries+`}`), t.TempDir()); err == nil {
			t.Fatal("case-insensitive duplicate should fail")
		}
	}
}

func TestCaseSensitiveConfigPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows paths are case-insensitive")
	}
	home := t.TempDir()
	c, err := Load(writeConfig(t, `{"accounts":[
		{"provider":"claude","dir":"~/.claude-Work","name":"one"},
		{"provider":"claude","dir":"~/.claude-work","name":"two"}
	]}`), home)
	if err != nil || len(c.Apply(nil)) != 2 {
		t.Fatalf("want separate paths, got %+v, %v", c, err)
	}
}

func TestPath(t *testing.T) {
	home := filepath.FromSlash("/h")
	t.Setenv("XDG_CONFIG_HOME", "")
	if got := Path(home); got != filepath.Join(home, ".config", "uc", "config.json") {
		t.Errorf("got %q", got)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.FromSlash("/x"))
	if got := Path(home); got != filepath.Join(filepath.FromSlash("/x"), "uc", "config.json") {
		t.Errorf("got %q", got)
	}
}

func TestExpandHome(t *testing.T) {
	home := filepath.FromSlash("/h")
	if got := ExpandHome("~/.codex", home); got != filepath.Join(home, ".codex") {
		t.Errorf("got %q", got)
	}
	if got := ExpandHome("~x", home); got != "~x" {
		t.Errorf("got %q", got)
	}
}

func TestStarterRoundTrip(t *testing.T) {
	home := t.TempDir()
	found := []discover.Account{{Provider: "codex", Name: "Copy", Dir: filepath.Join(home, ".codex - Copy")}}
	path := filepath.Join(t.TempDir(), "uc", "config.json")
	if err := Write(path, Starter(found, home)); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"dir": "~/.codex - Copy"`) {
		t.Errorf("bad starter:\n%s", data)
	}
	c, err := Load(path, home)
	if err != nil || len(c.Apply(found)) != 1 || c.Accounts[0].Dir != found[0].Dir {
		t.Errorf("round trip failed: %+v %v", c, err)
	}
	if err := Write(path, Starter(found, home)); err == nil {
		t.Error("want error when config exists")
	}
}
