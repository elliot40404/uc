package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

var now = time.Date(2026, 9, 25, 3, 0, 0, 0, time.UTC)

type fake struct {
	calls  *int
	status usage.Status
}

func (f fake) Report(_ context.Context, name, dir string, _ time.Time) usage.Report {
	*f.calls++
	return usage.Report{Provider: "claude", Account: name, Dir: dir, Status: f.status,
		Windows: []usage.Window{{Name: usage.Week, UsedPct: 40}}}
}

func setup(t *testing.T, status usage.Status) (Reporter, *int, string) {
	path := filepath.Join(t.TempDir(), "uc", "usage.json")
	calls := 0
	return Reporter{Inner: fake{&calls, status}, Store: Open(path), Provider: "claude"}, &calls, path
}

func TestFreshCacheSkipsNetwork(t *testing.T) {
	c, calls, _ := setup(t, usage.StatusOK)
	c.Report(context.Background(), "a", "/d", now)
	dir := "/d"
	if runtime.GOOS == "windows" {
		dir = "/D"
	}
	r := c.Report(context.Background(), "renamed", dir, now.Add(30*time.Second))
	if *calls != 1 || r.Account != "renamed" || r.Stale {
		t.Errorf("calls=%d report=%+v", *calls, r)
	}
	c.Report(context.Background(), "a", "/d", now.Add(FreshFor))
	if *calls != 2 {
		t.Errorf("old cache should refetch, calls=%d", *calls)
	}
}

func TestRateLimitShowsStaleAndBacksOff(t *testing.T) {
	c, calls, _ := setup(t, usage.StatusOK)
	c.Report(context.Background(), "a", "/d", now)
	c.Inner = fake{calls, usage.StatusLimited}
	later := now.Add(5 * time.Minute)
	r := c.Report(context.Background(), "a", "/d", later)
	if !r.Stale || r.Status != usage.StatusOK || r.Windows[0].UsedPct != 40 {
		t.Errorf("want stale cached report, got %+v", r)
	}
	c.Report(context.Background(), "a", "/d", later.Add(time.Minute))
	if *calls != 2 {
		t.Errorf("blocked provider must not call network, calls=%d", *calls)
	}
	c.Report(context.Background(), "a", "/d", later.Add(firstBackoff))
	if *calls != 3 {
		t.Errorf("should retry after backoff, calls=%d", *calls)
	}
}

func TestRateLimitWithoutCache(t *testing.T) {
	c, _, _ := setup(t, usage.StatusLimited)
	if r := c.Report(context.Background(), "a", "/d", now); r.Status != usage.StatusLimited || r.Stale {
		t.Errorf("got %+v", r)
	}
}

func TestStaleTooOldIsDropped(t *testing.T) {
	c, calls, _ := setup(t, usage.StatusOK)
	c.Report(context.Background(), "a", "/d", now)
	c.Inner = fake{calls, usage.StatusLimited}
	if r := c.Report(context.Background(), "a", "/d", now.Add(staleMax)); r.Status != usage.StatusLimited {
		t.Errorf("got %+v", r)
	}
}

func TestSaveAndReopen(t *testing.T) {
	c, calls, path := setup(t, usage.StatusOK)
	dir := filepath.Join(t.TempDir(), "private-account")
	c.Email = func(string) string { return "new@example.test" }
	c.Report(context.Background(), "private-user", dir, now)
	if err := c.Store.Save(); err != nil {
		t.Fatal(err)
	}
	assertNoIdentity(t, path)
	c.Store = Open(path)
	r := c.Report(context.Background(), "renamed", dir, now.Add(time.Second))
	if *calls != 1 {
		t.Errorf("reopened cache should be used, calls=%d", *calls)
	}
	if r.Account != "renamed" || r.Dir != dir || r.Email != "new@example.test" {
		t.Errorf("cached identity should come from local inputs, got %+v", r)
	}
}

func TestLegacyCacheIsSanitized(t *testing.T) {
	c, calls, path := setup(t, usage.StatusOK)
	dir := filepath.Join(t.TempDir(), "private-account")
	legacy := data{Accounts: map[string]usage.Report{
		"claude|" + strings.ToLower(filepath.Clean(dir)): {
			Provider: "claude", Account: "private-user", Dir: dir, Email: "old@example.test",
			Status: usage.StatusOK, UpdatedAt: now,
		},
	}}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	c.Store = Open(path)
	c.Email = func(string) string { return "new@example.test" }
	r := c.Report(context.Background(), "renamed", dir, now.Add(time.Second))
	if *calls != 0 || r.Account != "renamed" || r.Dir != dir || r.Email != "new@example.test" {
		t.Errorf("legacy cache not hydrated: calls=%d report=%+v", *calls, r)
	}
	if err := c.Store.Save(); err != nil {
		t.Fatal(err)
	}
	assertNoIdentity(t, path)
}

func TestOldHashedCacheIsNotSharedOnCaseSensitiveSystems(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows paths are case-insensitive")
	}
	c, calls, path := setup(t, usage.StatusOK)
	dir := filepath.Join(t.TempDir(), "Account")
	oldKey := key("claude", strings.ToLower(dir))
	raw, err := json.Marshal(data{Accounts: map[string]usage.Report{oldKey: {
		Provider: "claude", Status: usage.StatusOK, UpdatedAt: now,
	}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	c.Store = Open(path)
	c.Report(context.Background(), "a", strings.ToLower(dir), now)
	if *calls != 1 {
		t.Fatal("old ambiguous cache key was reused")
	}
}

func assertNoIdentity(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"private-account", "private-user", "old@example.test", "new@example.test"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("cache contains %q", secret)
		}
	}
}

func TestBackoffGrowsAndCaps(t *testing.T) {
	cases := map[int]time.Duration{1: 2 * time.Minute, 2: 4 * time.Minute, 3: 8 * time.Minute, 10: maxBackoff}
	for strikes, want := range cases {
		if got := backoff(strikes); got != want {
			t.Errorf("backoff(%d) = %v, want %v", strikes, got, want)
		}
	}
}
