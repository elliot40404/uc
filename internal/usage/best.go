package usage

import "time"

const (
	fullWeek   = 7 * 24 * time.Hour
	minHorizon = 5 * time.Hour
)

type Pick struct {
	Report     Report
	WeekLeft   float64
	WeekResets time.Time
	UsableAt   time.Time
	Score      float64
}

func Best(reports []Report, provider string, now time.Time) (Pick, bool) {
	var best Pick
	found := false
	for _, r := range reports {
		if provider != "" && r.Provider != provider {
			continue
		}
		p, ok := score(r, now)
		if ok && (!found || better(p, best)) {
			best, found = p, true
		}
	}
	return best, found
}

func score(r Report, now time.Time) (Pick, bool) {
	week, ok := r.Window(Week)
	if r.Status != StatusOK || !ok || week.UsedPct >= 100 {
		return Pick{}, false
	}
	horizon := fullWeek
	if !week.ResetsAt.IsZero() {
		horizon = max(week.ResetsAt.Sub(now), minHorizon)
	}
	p := Pick{Report: r, WeekLeft: 100 - week.UsedPct, WeekResets: week.ResetsAt}
	p.Score = p.WeekLeft / (horizon.Hours() / 24)
	if s, ok := r.Window(Session); ok && s.UsedPct >= 100 {
		p.UsableAt = s.ResetsAt
	}
	return p, true
}

func better(a, b Pick) bool {
	if a.Score != b.Score {
		return a.Score > b.Score
	}
	return a.WeekLeft > b.WeekLeft
}
