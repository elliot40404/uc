package flyout

import (
	"time"

	"github.com/elliot40404/uc/internal/render"
	"github.com/elliot40404/uc/internal/tray"
	"github.com/elliot40404/uc/internal/usage"
)

type View struct {
	Status   string
	Sections []Section
}

type Section struct {
	Provider string
	Accounts []Account
}

type Account struct {
	Name    string
	Plan    string
	Pick    bool
	Note    string
	Problem bool
	Bars    []Bar
}

type Bar struct {
	Name   string
	Used   float64
	Resets string
}

func Build(reports []usage.Report, now time.Time) View {
	picks := tray.Picks(reports, now)
	var v View
	for _, p := range tray.Providers(reports) {
		s := Section{Provider: p}
		for _, r := range reports {
			if r.Provider == p {
				s.Accounts = append(s.Accounts, account(usage.TerminalReport(r), tray.IsPick(r, picks), now))
			}
		}
		v.Sections = append(v.Sections, s)
	}
	return v
}

func account(r usage.Report, pick bool, now time.Time) Account {
	a := Account{Name: r.Account, Plan: r.Plan, Pick: pick}
	if r.Status != usage.StatusOK {
		a.Note, a.Problem = problem(r), true
		return a
	}
	if r.Stale {
		a.Note = "cached " + r.UpdatedAt.Local().Format("15:04")
	}
	for _, w := range r.Windows {
		b := Bar{Name: w.Name, Used: w.UsedPct}
		if !w.ResetsAt.IsZero() {
			b.Resets = render.Until(w.ResetsAt, now)
		}
		a.Bars = append(a.Bars, b)
	}
	return a
}

func problem(r usage.Report) string {
	switch r.Status {
	case usage.StatusExpired:
		return "Login expired, open " + r.Provider + " once to renew"
	case usage.StatusNoLogin:
		return "Not logged in"
	case usage.StatusLimited:
		return "Rate limited, try again soon"
	default:
		return "Error: " + r.Error
	}
}
