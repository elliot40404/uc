package cache

import (
	"context"
	"time"

	"github.com/elliot40404/uc/internal/collect"
	"github.com/elliot40404/uc/internal/usage"
)

const (
	FreshFor = time.Minute
	staleMax = 6 * time.Hour
)

type Reporter struct {
	Inner    collect.Reporter
	Store    *Store
	Provider string
	Email    func(string) string
}

func (c Reporter) Report(ctx context.Context, name, dir string, now time.Time) usage.Report {
	cached, ok := c.Store.get(c.Provider, dir)
	cached.Account = name
	cached.Dir = dir
	if ok && c.Email != nil {
		cached.Email = c.Email(dir)
	}
	switch {
	case ok && now.Sub(cached.UpdatedAt) < FreshFor:
		return cached
	case c.Store.blocked(c.Provider, now):
		return c.fallback(cached, ok, name, dir, now)
	}
	r := c.Inner.Report(ctx, name, dir, now)
	switch r.Status {
	case usage.StatusOK:
		r.UpdatedAt = now
		c.Store.put(r)
	case usage.StatusLimited:
		c.Store.strike(c.Provider, now)
		return c.fallback(cached, ok, name, dir, now)
	}
	return r
}

func (c Reporter) fallback(cached usage.Report, ok bool, name, dir string, now time.Time) usage.Report {
	if ok && now.Sub(cached.UpdatedAt) < staleMax {
		cached.Stale = true
		return cached
	}
	return usage.Report{Provider: c.Provider, Account: name, Dir: dir, Status: usage.StatusLimited}
}
