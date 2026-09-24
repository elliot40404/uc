package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/discover"
)

var Providers = []string{"claude", "codex"}

type Entry struct {
	Provider string `json:"provider"`
	Dir      string `json:"dir"`
	Name     string `json:"name,omitempty"`
	Hide     bool   `json:"hide,omitempty"`
}

type Defaults struct {
	Mini       bool          `json:"mini,omitempty"`
	Live       bool          `json:"live,omitempty"`
	AltScreen  bool          `json:"alt_screen,omitempty"`
	Compact    bool          `json:"compact,omitempty"`
	ShowEmails bool          `json:"show_emails,omitempty"`
	Every      string        `json:"every,omitempty"`
	Refresh    time.Duration `json:"-"`
}

type Config struct {
	Defaults Defaults `json:"defaults,omitzero"`
	Accounts []Entry  `json:"accounts"`
}

const DefaultRefresh = 5 * time.Minute

func Path(home string) string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "uc", "config.json")
}

func Load(path, home string) (Config, error) {
	c := Config{Defaults: Defaults{Live: true, AltScreen: true, Refresh: DefaultRefresh}}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return c, fmt.Errorf("%s: %w", path, err)
	}
	return c, c.normalize(path, home)
}

func (c *Config) normalize(path, home string) error {
	c.Defaults.Refresh = DefaultRefresh
	if c.Defaults.Every != "" {
		d, err := time.ParseDuration(c.Defaults.Every)
		if err != nil || d <= 0 {
			return fmt.Errorf("%s: defaults.every must be a duration like 5m", path)
		}
		c.Defaults.Refresh = d
	}
	seen := make(map[string]bool, len(c.Accounts))
	for i := range c.Accounts {
		e := &c.Accounts[i]
		if !slices.Contains(Providers, e.Provider) {
			return fmt.Errorf("%s: account %d: provider must be one of %v", path, i+1, Providers)
		}
		if e.Dir == "" {
			return fmt.Errorf("%s: account %d: dir is required", path, i+1)
		}
		e.Dir = ExpandHome(e.Dir, home)
		id := key(e.Provider, e.Dir)
		if seen[id] {
			return fmt.Errorf("%s: account %d: duplicate provider and dir", path, i+1)
		}
		seen[id] = true
	}
	return nil
}

func ExpandHome(dir, home string) string {
	if dir == "~" || strings.HasPrefix(dir, "~/") || strings.HasPrefix(dir, `~\`) {
		dir = filepath.Join(home, dir[1:])
	}
	return filepath.Clean(dir)
}

func Starter(found []discover.Account, home string) Config {
	var c Config
	for _, a := range found {
		c.Accounts = append(c.Accounts, Entry{Provider: a.Provider, Dir: shortHome(a.Dir, home), Name: a.Name})
	}
	return c
}

func Write(path string, c Config) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func shortHome(dir, home string) string {
	rel, err := filepath.Rel(home, dir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return dir
	}
	return "~/" + filepath.ToSlash(rel)
}
