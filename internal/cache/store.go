package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/elliot40404/uc/internal/pathkey"
	"github.com/elliot40404/uc/internal/usage"
)

const (
	firstBackoff = 2 * time.Minute
	maxBackoff   = 30 * time.Minute
)

type block struct {
	Until   time.Time `json:"until,omitzero"`
	Strikes int       `json:"strikes,omitempty"`
}

type data struct {
	Version   int                     `json:"version"`
	Accounts  map[string]usage.Report `json:"accounts"`
	Providers map[string]block        `json:"providers"`
}

type Store struct {
	mu   sync.Mutex
	path string
	data data
}

func Path() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "uc", "usage.json"), nil
}

func Open(path string) *Store {
	s := &Store{path: path, data: data{Version: 2, Accounts: map[string]usage.Report{}, Providers: map[string]block{}}}
	raw, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	var d data
	if json.Unmarshal(raw, &d) == nil {
		for k, r := range d.Accounts {
			if strings.Contains(k, "|") {
				if r.Provider == "" || r.Dir == "" {
					continue
				}
				k = key(r.Provider, r.Dir)
			} else if d.Version < 2 && runtime.GOOS != "windows" {
				continue
			} else if len(k) != 64 {
				continue
			} else if _, err := hex.DecodeString(k); err != nil {
				continue
			}
			s.data.Accounts[k] = withoutIdentity(r)
		}
		if d.Providers != nil {
			s.data.Providers = d.Providers
		}
	}
	return s
}

func (s *Store) Save() error {
	s.mu.Lock()
	raw, err := json.Marshal(s.data)
	s.mu.Unlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), "usage-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}

func key(provider, dir string) string {
	sum := sha256.Sum256([]byte(provider + "|" + pathkey.Of(dir)))
	return hex.EncodeToString(sum[:])
}

func withoutIdentity(r usage.Report) usage.Report {
	r.Account = ""
	r.Dir = ""
	r.Email = ""
	r.Error = ""
	r.Stale = false
	return r
}

func (s *Store) get(provider, dir string) (usage.Report, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.data.Accounts[key(provider, dir)]
	return r, ok
}

func (s *Store) put(r usage.Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Accounts[key(r.Provider, r.Dir)] = withoutIdentity(r)
	delete(s.data.Providers, r.Provider)
}

func (s *Store) blocked(provider string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return now.Before(s.data.Providers[provider].Until)
}

func (s *Store) strike(provider string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b := s.data.Providers[provider]
	if now.Before(b.Until) {
		return
	}
	b.Strikes++
	b.Until = now.Add(backoff(b.Strikes))
	s.data.Providers[provider] = b
}

func backoff(strikes int) time.Duration {
	d := firstBackoff
	for i := 1; i < strikes && d < maxBackoff; i++ {
		d *= 2
	}
	return min(d, maxBackoff)
}
