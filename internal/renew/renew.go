package renew

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

const (
	Early     = time.Minute
	lockWait  = 5 * time.Second
	lockPoll  = 100 * time.Millisecond
	lockStale = 30 * time.Second
)

var ErrBusy = errors.New("renew lock busy")

type Login interface {
	Expired(now time.Time) bool
}

func Due(l Login, now time.Time) bool {
	return l.Expired(now.Add(Early))
}

func Once[T Login](path string, now time.Time, load func() (T, error), renew func(T) (T, error)) (T, error) {
	var zero T
	unlock, err := lock(path + ".uc-lock")
	if err != nil {
		return zero, err
	}
	defer unlock()
	cur, err := load()
	if err != nil || !Due(cur, now) {
		return cur, err
	}
	return renew(cur)
}

func lock(path string) (func(), error) {
	deadline := time.Now().Add(lockWait)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			f.Close()
			return func() { os.Remove(path) }, nil
		}
		if !errors.Is(err, fs.ErrExist) && !errors.Is(err, fs.ErrPermission) {
			return nil, err
		}
		if stale(path) {
			os.Remove(path)
			continue
		}
		if time.Now().After(deadline) {
			return nil, ErrBusy
		}
		time.Sleep(lockPoll)
	}
}

func stale(path string) bool {
	info, err := os.Stat(path)
	return err == nil && time.Since(info.ModTime()) > lockStale
}
