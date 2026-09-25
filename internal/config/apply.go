package config

import (
	"cmp"
	"path/filepath"

	"github.com/elliot40404/uc/internal/discover"
	"github.com/elliot40404/uc/internal/pathkey"
)

type Item struct {
	discover.Account
	Hide bool
}

func (c Config) Apply(found []discover.Account) []discover.Account {
	var out []discover.Account
	for _, it := range c.Merge(found) {
		if !it.Hide {
			out = append(out, it.Account)
		}
	}
	return out
}

func (c Config) Merge(found []discover.Account) []Item {
	entries := map[string]Entry{}
	for _, e := range c.Accounts {
		entries[key(e.Provider, e.Dir)] = e
	}
	var out []Item
	for _, a := range found {
		e := entries[key(a.Provider, a.Dir)]
		delete(entries, key(a.Provider, a.Dir))
		a.Name = cmp.Or(e.Name, a.Name)
		out = append(out, Item{Account: a, Hide: e.Hide})
	}
	for _, e := range c.Accounts {
		if _, extra := entries[key(e.Provider, e.Dir)]; extra {
			a := discover.Account{Provider: e.Provider, Dir: e.Dir, Name: cmp.Or(e.Name, filepath.Base(e.Dir))}
			out = append(out, Item{Account: a, Hide: e.Hide})
		}
	}
	return out
}

func key(provider, dir string) string {
	return provider + "|" + pathkey.Of(dir)
}
