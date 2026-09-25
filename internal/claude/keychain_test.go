package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKeychainServiceDefault(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	if got := keychainService(filepath.Join(home, ".claude")); got != keychainBase {
		t.Fatalf("want %q, got %q", keychainBase, got)
	}
}

func TestKeychainServiceCustomDir(t *testing.T) {
	if filepath.Separator != '/' {
		t.Skip("keychain service names hash POSIX paths")
	}
	want := "Claude Code-credentials-1e91dd84"
	if got := keychainService("/Users/me/.claude-work"); got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
	if got := keychainService("/Users/me/.claude-work/"); got != want {
		t.Fatalf("trailing slash: want %q, got %q", want, got)
	}
}
