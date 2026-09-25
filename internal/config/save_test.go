package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestSaveRoundTrip(t *testing.T) {
	home := t.TempDir()
	path := writeConfig(t, `{"accounts":[{"provider":"codex","dir":"~/.codex-work","name":"work","hide":true}]}`)
	c, err := Load(path, home)
	if err != nil {
		t.Fatal(err)
	}
	c.Defaults.Live, c.Defaults.AltScreen, c.Defaults.Mini = false, false, true
	c.Defaults.Refresh = 2 * time.Minute
	if err := Save(path, c, home); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"dir": "~/.codex-work"`) || !strings.Contains(string(data), `"every": "2m"`) {
		t.Errorf("bad saved config:\n%s", data)
	}
	got, err := Load(path, home)
	if err != nil || got.Defaults.Live || got.Defaults.AltScreen || !got.Defaults.Mini || got.Defaults.Refresh != 2*time.Minute {
		t.Fatalf("defaults lost: %+v %v", got.Defaults, err)
	}
	if len(got.Accounts) != 1 || !got.Accounts[0].Hide || got.Accounts[0].Name != "work" {
		t.Fatalf("accounts lost: %+v", got.Accounts)
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uc", "config.json")
	if err := Save(path, Config{Defaults: baseDefaults()}, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("want only config.json, got %v", entries)
	}
	info, _ := os.Stat(path)
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("perm = %v", info.Mode().Perm())
	}
}

func TestStarterKeepsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := Write(path, Starter(nil, t.TempDir()), t.TempDir()); err != nil {
		t.Fatal(err)
	}
	c, err := Load(path, t.TempDir())
	c.Defaults.Every = ""
	if err != nil || c.Defaults != baseDefaults() {
		t.Fatalf("got %+v %v", c.Defaults, err)
	}
}

func TestFormatEvery(t *testing.T) {
	cases := map[time.Duration]string{
		30 * time.Second: "30s",
		time.Minute:      "1m",
		90 * time.Second: "1m30s",
		time.Hour:        "1h",
		90 * time.Minute: "1h30m",
	}
	for d, want := range cases {
		if got := FormatEvery(d); got != want {
			t.Errorf("%v: got %q, want %q", d, got, want)
		}
	}
}
