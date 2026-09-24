package app

import (
	"context"
	"errors"
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

var ErrCacheSave = errors.New("usage cache was not saved")

type Env struct {
	Home       string
	ConfigPath string
	Cfg        config.Config
}

func New() (Env, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Env{}, err
	}
	path := config.Path(home)
	cfg, err := config.Load(path, home)
	return Env{Home: home, ConfigPath: path, Cfg: cfg}, err
}

func (e Env) Found() []discover.Account {
	return append(
		discover.Scan("claude", e.Home, ".claude", os.Getenv("CLAUDE_CONFIG_DIR")),
		discover.Scan("codex", e.Home, ".codex", os.Getenv("CODEX_HOME"))...,
	)
}

func (e Env) Reports(parent context.Context) ([]usage.Report, time.Time, error) {
	accounts := e.Cfg.Apply(e.Found())
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
		return reports, now, ErrCacheSave
	}
	return reports, now, nil
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
