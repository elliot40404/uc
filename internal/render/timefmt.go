package render

import (
	"fmt"
	"time"
)

func Until(t, now time.Time) string {
	d := t.Sub(now).Round(time.Minute)
	if d <= 0 {
		return "now"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	switch {
	case days > 0 && hours > 0:
		return fmt.Sprintf("in %dd %dh", days, hours)
	case days > 0:
		return fmt.Sprintf("in %dd", days)
	case hours > 0 && mins > 0:
		return fmt.Sprintf("in %dh %dm", hours, mins)
	case hours > 0:
		return fmt.Sprintf("in %dh", hours)
	default:
		return fmt.Sprintf("in %dm", mins)
	}
}

func Clock(t, now time.Time) string {
	t = t.Local().Round(time.Minute)
	now = now.Local()
	switch {
	case sameDay(t, now):
		return t.Format("15:04")
	case t.Sub(now) < 6*24*time.Hour:
		return t.Format("Mon 15:04")
	default:
		return t.Format("Mon 2 Jan 15:04")
	}
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
