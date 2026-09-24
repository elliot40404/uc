package render

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func level(pct float64) string {
	switch {
	case pct >= 90:
		return red
	case pct >= 60:
		return yellow
	default:
		return green
	}
}

func bar(pct float64, width int, full, empty string) cell {
	n := int(math.Round(pct / 100 * float64(width)))
	if pct > 0 && n == 0 {
		n = 1
	}
	n = min(max(n, 0), width)
	return cell{
		{level(pct), strings.Repeat(full, n)},
		{dim, strings.Repeat(empty, width-n)},
	}
}

func percent(pct float64) cell {
	return text(level(pct), fmt.Sprintf("%3.0f%%", pct))
}

func status(r usage.Report) cell {
	if r.Stale {
		return text(yellow, "cached "+r.UpdatedAt.Local().Format("15:04")+", rate limited")
	}
	switch r.Status {
	case usage.StatusOK:
		return text(green, "ok")
	case usage.StatusExpired:
		return text(yellow, "expired, run "+r.Provider+" once in this account")
	case usage.StatusError:
		return text(red, "error: "+r.Error)
	case usage.StatusLimited:
		return text(yellow, "rate limited, try again in a few minutes")
	default:
		return text(yellow, string(r.Status))
	}
}

func resets(w usage.Window, now time.Time, prefix string) cell {
	if w.ResetsAt.IsZero() {
		return text(dim, "not started")
	}
	return cell{
		{"", prefix + Clock(w.ResetsAt, now)},
		{dim, " (" + Until(w.ResetsAt, now) + ")"},
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
