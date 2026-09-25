package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/discover"
)

func Starter(found []discover.Account, home string) Config {
	c := Config{Defaults: baseDefaults()}
	for _, a := range found {
		c.Accounts = append(c.Accounts, Entry{Provider: a.Provider, Dir: a.Dir, Name: a.Name})
	}
	return c
}

func Write(path string, c Config, home string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}
	return Save(path, c, home)
}

func Save(path string, c Config, home string) error {
	if c.Defaults.Refresh > 0 {
		c.Defaults.Every = FormatEvery(c.Defaults.Refresh)
	}
	c.Accounts = shortDirs(c.Accounts, home)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(path, append(data, '\n'))
}

func FormatEvery(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	return s
}

func shortDirs(entries []Entry, home string) []Entry {
	out := make([]Entry, len(entries))
	for i, e := range entries {
		e.Dir = ShortHome(e.Dir, home)
		out[i] = e
	}
	return out
}

func ShortHome(dir, home string) string {
	rel, err := filepath.Rel(home, dir)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
		return dir
	}
	return "~/" + filepath.ToSlash(rel)
}

func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}
