package usage

import (
	"errors"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)

func acct(provider, name string, weekUsed float64, weekIn time.Duration, sessionUsed float64) Report {
	r := Report{Provider: provider, Account: name, Status: StatusOK, Windows: []Window{
		{Name: Session, UsedPct: sessionUsed, ResetsAt: now.Add(2 * time.Hour)},
		{Name: Week, UsedPct: weekUsed},
	}}
	if weekIn > 0 {
		r.Windows[1].ResetsAt = now.Add(weekIn)
	}
	return r
}

const day = 24 * time.Hour

func TestBestPrefersQuotaAboutToExpire(t *testing.T) {
	reports := []Report{
		acct("claude", "a", 30, 4*day, 0),
		acct("claude", "b", 40, 4*day, 0),
		acct("claude", "c", 10, 36*time.Hour, 0),
		acct("codex", "d", 0, 6*day, 0),
	}
	p, ok := Best(reports, "claude", now)
	if !ok || p.Report.Account != "c" || p.WeekLeft != 90 || p.Score != 60 {
		t.Errorf("got %+v", p)
	}
	if p, _ := Best(reports, "codex", now); p.Report.Account != "d" {
		t.Errorf("codex best = %q", p.Report.Account)
	}
}

func TestBestKeepsBlockedSessionWithWait(t *testing.T) {
	p, ok := Best([]Report{acct("claude", "a", 10, day, 100), acct("claude", "b", 10, 5*day, 0)}, "", now)
	if !ok || p.Report.Account != "a" || !p.UsableAt.Equal(now.Add(2*time.Hour)) {
		t.Errorf("got %+v", p)
	}
}

func TestBestSkipsUnusable(t *testing.T) {
	reports := []Report{
		{Provider: "claude", Account: "x", Status: StatusExpired},
		acct("claude", "full", 100, day, 0),
		{Provider: "claude", Account: "noweek", Status: StatusOK, Windows: []Window{{Name: Session}}},
	}
	if p, found := Best(reports, "", now); found {
		t.Errorf("want none, got %+v", p)
	}
}

func TestBestHorizonFloorAndNotStarted(t *testing.T) {
	soon := acct("claude", "soon", 50, time.Minute, 0)
	fresh := acct("claude", "fresh", 0, 0, 0)
	p, _ := Best([]Report{soon, fresh}, "", now)
	if p.Report.Account != "soon" || p.Score != 50/(5.0/24) {
		t.Errorf("got %+v", p)
	}
	if p, _ := Best([]Report{fresh}, "", now); p.Score != 100.0/7 {
		t.Errorf("not started week should use 7 days, got %v", p.Score)
	}
}

func TestLeft(t *testing.T) {
	if got := acct("c", "a", 30, day, 80).Left(); got != 20 {
		t.Errorf("Left = %v, want 20", got)
	}
}

func TestFail(t *testing.T) {
	cases := map[error]Status{ErrNoLogin: StatusNoLogin, ErrExpired: StatusExpired, errors.New("x"): StatusError}
	for err, want := range cases {
		if got := (Report{}).Fail(err).Status; got != want {
			t.Errorf("Fail(%v) = %q, want %q", err, got, want)
		}
	}
}
