package discover

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestScan(t *testing.T) {
	home := t.TempDir()
	for _, d := range []string{".claude", ".claude-work", ".codex - Copy", "other"} {
		os.Mkdir(filepath.Join(home, d), 0o700)
	}
	os.WriteFile(filepath.Join(home, ".claude.json"), nil, 0o600)
	extra := filepath.Join(t.TempDir(), "myclaude")
	os.Mkdir(extra, 0o700)

	got := Scan("claude", home, ".claude", extra)
	want := []string{"default", "myclaude", "work"}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i, a := range got {
		if a.Name != want[i] || a.Provider != "claude" {
			t.Errorf("got %+v, want name %q", a, want[i])
		}
	}

	codex := Scan("codex", home, ".codex", "")
	if len(codex) != 1 || codex[0].Name != "Copy" {
		t.Errorf("got %+v", codex)
	}
}

func TestScanDedupesEnvDir(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude-personal")
	os.Mkdir(dir, 0o700)
	if got := Scan("claude", home, ".claude", dir); len(got) != 1 {
		t.Errorf("want 1 account, got %+v", got)
	}
}

func TestScanKeepsCaseSensitiveAccounts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows paths are case-insensitive")
	}
	home := t.TempDir()
	if err := os.Mkdir(filepath.Join(home, ".claude-Work"), 0o700); err != nil {
		t.Fatal(err)
	}
	if isDir(filepath.Join(home, ".claude-work")) {
		t.Skip("file system is case-insensitive")
	}
	if err := os.Mkdir(filepath.Join(home, ".claude-work"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got := Scan("claude", home, ".claude", ""); len(got) != 2 {
		t.Fatalf("want two accounts, got %+v", got)
	}
}
