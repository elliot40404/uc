package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/config"
)

func TestReportsKeepUsageWhenCacheSaveFails(t *testing.T) {
	home := t.TempDir()
	blocked := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blocked, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LOCALAPPDATA", blocked)
	t.Setenv("XDG_CACHE_HOME", blocked)
	e := env{home: home}
	reports, at, err := e.reports(t.Context())
	if !errors.Is(err, errCacheSave) || at.IsZero() || len(reports) != 0 {
		t.Fatalf("reports=%+v time=%v err=%v", reports, at, err)
	}
}

func TestParseFlagsAnywhere(t *testing.T) {
	opts, pos, err := parse([]string{"best", "--json", "claude", "--every", "2m"}, "", config.Defaults{Refresh: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.json || opts.every != 2*time.Minute || len(pos) != 2 || pos[0] != "best" || pos[1] != "claude" {
		t.Errorf("got %+v %v", opts, pos)
	}
}

func TestParseUsesConfigDefaults(t *testing.T) {
	d := config.Defaults{Mini: true, Live: true, AltScreen: true, Refresh: 3 * time.Minute}
	opts, _, _ := parse(nil, "", d)
	if !opts.mini || !opts.live || !opts.altScreen || opts.every != 3*time.Minute {
		t.Errorf("defaults not applied: %+v", opts)
	}
	opts, _, _ = parse([]string{"--live=false", "--alt-screen=false"}, "", d)
	if !opts.mini || opts.live || opts.altScreen {
		t.Errorf("flag should override config: %+v", opts)
	}
	if opts.emails {
		t.Error("emails should be hidden by default")
	}
	opts, _, _ = parse(nil, "", config.Defaults{ShowEmails: true})
	if !opts.emails {
		t.Error("show_emails should show emails")
	}
	opts, _, _ = parse([]string{"--emails"}, "", d)
	if !opts.emails {
		t.Error("--emails should show emails")
	}
	opts, _, _ = parse([]string{"--full"}, "", d)
	if opts.mini || !opts.live || !opts.altScreen {
		t.Errorf("--full should turn off mini only: %+v", opts)
	}
}
