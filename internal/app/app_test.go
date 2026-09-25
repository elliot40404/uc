package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReportsKeepUsageWhenCacheSaveFails(t *testing.T) {
	home := t.TempDir()
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOCALAPPDATA", blocked)
	t.Setenv("XDG_CACHE_HOME", blocked)
	t.Setenv("HOME", blocked)
	t.Setenv("CLAUDE_CONFIG_DIR", "")
	t.Setenv("CODEX_HOME", "")
	e := Env{Home: home}
	reports, at, err := e.Reports(t.Context())
	if !errors.Is(err, ErrCacheSave) || at.IsZero() || len(reports) != 0 {
		t.Fatalf("reports=%+v time=%v err=%v", reports, at, err)
	}
}
