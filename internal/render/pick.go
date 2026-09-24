package render

import (
	"fmt"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func WeekLeft(p usage.Pick) string {
	return fmt.Sprintf("%.0f%% of 7d left", p.WeekLeft)
}

func WeekResets(p usage.Pick, now time.Time) string {
	if p.WeekResets.IsZero() {
		return "7d not started"
	}
	return "resets " + Until(p.WeekResets, now)
}

func UsableIn(p usage.Pick, now time.Time) string {
	if p.UsableAt.IsZero() || !p.UsableAt.After(now) {
		return ""
	}
	return "5h full, usable " + Until(p.UsableAt, now)
}

func PickSummary(p usage.Pick, now time.Time) string {
	s := WeekLeft(p) + ", " + WeekResets(p, now)
	if u := UsableIn(p, now); u != "" {
		s += ", " + u
	}
	return s
}
