package renew

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type login struct{ exp time.Time }

func (l login) Expired(now time.Time) bool { return !l.exp.After(now) }

func TestOnceSkipsFreshLogin(t *testing.T) {
	now := time.Now()
	path := filepath.Join(t.TempDir(), "creds.json")
	load := func() (login, error) { return login{now.Add(time.Hour)}, nil }
	got, err := Once(path, now, load, func(login) (login, error) {
		t.Fatal("renew called for fresh login")
		return login{}, nil
	})
	if err != nil || !got.exp.Equal(now.Add(time.Hour)) {
		t.Fatalf("got %v %v", got, err)
	}
}

func TestOnceRenewsOnlyOnceAcrossCallers(t *testing.T) {
	now := time.Now()
	path := filepath.Join(t.TempDir(), "creds.json")
	var mu sync.Mutex
	cur := login{now.Add(-time.Hour)}
	var calls atomic.Int32
	load := func() (login, error) { mu.Lock(); defer mu.Unlock(); return cur, nil }
	renew := func(login) (login, error) {
		calls.Add(1)
		mu.Lock()
		defer mu.Unlock()
		cur = login{now.Add(time.Hour)}
		return cur, nil
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			if _, err := Once(path, now, load, renew); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("renew calls = %d", calls.Load())
	}
	if _, err := os.Stat(path + ".uc-lock"); !os.IsNotExist(err) {
		t.Fatal("lock file left behind")
	}
}

func TestStaleLockIsTakenOver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.uc-lock")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Minute)
	os.Chtimes(path, old, old)
	unlock, err := lock(path)
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}
