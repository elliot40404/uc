package config

import (
	"slices"
	"strings"
)

const MaxNameLen = 32

func (c Config) SetHidden(provider, dir string, hide bool) Config {
	return c.edit(provider, dir, func(e *Entry) { e.Hide = hide })
}

func (c Config) SetName(provider, dir, name string) Config {
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > MaxNameLen {
		name = string(r[:MaxNameLen])
	}
	return c.edit(provider, dir, func(e *Entry) { e.Name = name })
}

func (c Config) edit(provider, dir string, change func(*Entry)) Config {
	c.Accounts = slices.Clone(c.Accounts)
	id := key(provider, dir)
	i := slices.IndexFunc(c.Accounts, func(e Entry) bool { return key(e.Provider, e.Dir) == id })
	if i < 0 {
		c.Accounts = append(c.Accounts, Entry{Provider: provider, Dir: dir})
		i = len(c.Accounts) - 1
	}
	change(&c.Accounts[i])
	return c
}
