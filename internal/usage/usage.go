package usage

import (
	"errors"
	"time"
)

type Status string

const (
	StatusOK      Status = "ok"
	StatusExpired Status = "expired"
	StatusNoLogin Status = "not logged in"
	StatusError   Status = "error"
	StatusLimited Status = "rate limited"
)

type Window struct {
	Name     string    `json:"name"`
	UsedPct  float64   `json:"used_pct"`
	ResetsAt time.Time `json:"resets_at,omitzero"`
}

type Report struct {
	Provider  string    `json:"provider"`
	Account   string    `json:"account"`
	Dir       string    `json:"dir"`
	Plan      string    `json:"plan,omitempty"`
	Email     string    `json:"email,omitempty"`
	Status    Status    `json:"status"`
	Error     string    `json:"error,omitempty"`
	Windows   []Window  `json:"windows,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitzero"`
	Stale     bool      `json:"stale,omitempty"`
}

const (
	Session = "5h"
	Week    = "7d"
)

func WithoutEmails(reports []Report) []Report {
	out := make([]Report, len(reports))
	for i, r := range reports {
		r.Email = ""
		out[i] = r
	}
	return out
}

func (r Report) Window(name string) (Window, bool) {
	for _, w := range r.Windows {
		if w.Name == name {
			return w, true
		}
	}
	return Window{}, false
}

func (r Report) Extra() []Window {
	var out []Window
	for _, w := range r.Windows {
		if w.Name != Session && w.Name != Week {
			out = append(out, w)
		}
	}
	return out
}

func (r Report) Left() float64 {
	if r.Status != StatusOK || len(r.Windows) == 0 {
		return -1
	}
	worst := 0.0
	for _, w := range r.Windows {
		worst = max(worst, w.UsedPct)
	}
	return 100 - worst
}

var (
	ErrNoLogin = errors.New("no login")
	ErrExpired = errors.New("expired")
	ErrLimited = errors.New("rate limited")
)

func (r Report) Fail(err error) Report {
	switch {
	case errors.Is(err, ErrNoLogin):
		r.Status = StatusNoLogin
	case errors.Is(err, ErrLimited):
		r.Status = StatusLimited
	case errors.Is(err, ErrExpired):
		r.Status = StatusExpired
	default:
		r.Status = StatusError
		r.Error = err.Error()
	}
	return r
}
