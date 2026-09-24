package config

import (
	"cmp"
	"path/filepath"

	"github.com/elliot40404/uc/internal/discover"
	"github.com/elliot40404/uc/internal/pathkey"
)

func (c Config) Apply(found []discover.Account) []discover.Account {
	entries := map[string]Entry{}
	for _, e := range c.Accounts {
		entries[key(e.Provider, e.Dir)] = e
	}
	var out []discover.Account
	for _, a := range found {
		e, ok := entries[key(a.Provider, a.Dir)]
		delete(entries, key(a.Provider, a.Dir))
		if ok && e.Hide {
			continue
		}
		a.Name = cmp.Or(e.Name, a.Name)
		out = append(out, a)
	}
	for _, e := range c.Accounts {
		if _, extra := entries[key(e.Provider, e.Dir)]; extra && !e.Hide {
			out = append(out, discover.Account{Provider: e.Provider, Dir: e.Dir, Name: cmp.Or(e.Name, filepath.Base(e.Dir))})
		}
	}
	return out
}

func key(provider, dir string) string {
	return provider + "|" + pathkey.Of(dir)
}
