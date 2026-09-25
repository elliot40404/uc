package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/elliot40404/uc/internal/cache"
	"github.com/elliot40404/uc/internal/claude"
	"github.com/elliot40404/uc/internal/codex"
	"github.com/elliot40404/uc/internal/collect"
	"github.com/elliot40404/uc/internal/config"
	"github.com/elliot40404/uc/internal/discover"
	"github.com/elliot40404/uc/internal/usage"
)

const timeout = 15 * time.Second

var errCacheSave = errors.New("usage cache was not saved")

type env struct {
	home       string
	configPath string
	cfg        config.Config
}

func newEnv() (env, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return env{}, err
	}
	path := config.Path(home)
	cfg, err := config.Load(path, home)
	return env{home: home, configPath: path, cfg: cfg}, err
}

func (e env) saveConfig(c config.Config) error {
	return config.Save(e.configPath, c, e.home)
}

func (e env) found() []discover.Account {
	return append(
		discover.Scan("claude", e.home, ".claude", os.Getenv("CLAUDE_CONFIG_DIR")),
		discover.Scan("codex", e.home, ".codex", os.Getenv("CODEX_HOME"))...,
	)
}

func (e env) accounts() ([]discover.Account, error) {
	return e.cfg.Apply(e.found()), nil
}

func (e env) reports(parent context.Context) ([]usage.Report, time.Time, error) {
	accounts, err := e.accounts()
	if err != nil {
		return nil, time.Time{}, err
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	httpClient := &http.Client{Timeout: timeout}
	store := openCache()
	reporters := map[string]collect.Reporter{
		"claude": cached(store, "claude", claude.Client{HTTP: httpClient, BaseURL: claude.DefaultBaseURL}, claude.Email),
		"codex":  cached(store, "codex", codex.Client{HTTP: httpClient, BaseURL: codex.DefaultBaseURL}, codexEmail),
	}
	now := time.Now()
	reports := collect.All(ctx, accounts, reporters, now)
	if err := store.Save(); err != nil {
		return reports, now, errCacheSave
	}
	return reports, now, nil
}

func showCacheWarning(err error, at time.Time) error {
	if errors.Is(err, errCacheSave) && !at.IsZero() {
		fmt.Fprintln(os.Stderr, "uc:", errCacheSave)
		return nil
	}
	return err
}

func openCache() *cache.Store {
	path, err := cache.Path()
	if err != nil {
		path = filepath.Join(os.TempDir(), "uc", "usage.json")
	}
	return cache.Open(path)
}

func cached(store *cache.Store, provider string, inner collect.Reporter, email func(string) string) collect.Reporter {
	return cache.Reporter{Inner: inner, Store: store, Provider: provider, Email: email}
}

func codexEmail(dir string) string {
	creds, err := codex.LoadCreds(dir)
	if err != nil {
		return ""
	}
	return creds.Email
}
