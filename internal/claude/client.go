package claude

import (
	"context"
	"net/http"
	"time"

	"github.com/elliot40404/uc/internal/httpx"
	"github.com/elliot40404/uc/internal/usage"
)

const DefaultBaseURL = "https://api.anthropic.com"

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

type limit struct {
	Utilization *float64 `json:"utilization"`
	ResetsAt    string   `json:"resets_at"`
}

type usageResponse struct {
	FiveHour     *limit `json:"five_hour"`
	SevenDay     *limit `json:"seven_day"`
	SevenDayOpus *limit `json:"seven_day_opus"`
	SevenDaySonn *limit `json:"seven_day_sonnet"`
}

func (c Client) Usage(ctx context.Context, token string) ([]usage.Window, error) {
	var r usageResponse
	headers := map[string]string{
		"Authorization":  "Bearer " + token,
		"anthropic-beta": "oauth-2025-04-20",
	}
	if err := httpx.GetJSON(ctx, c.HTTP, c.BaseURL+"/api/oauth/usage", headers, &r); err != nil {
		return nil, err
	}
	return r.windows(), nil
}

func (r usageResponse) windows() []usage.Window {
	named := []struct {
		name string
		l    *limit
	}{
		{usage.Session, r.FiveHour},
		{usage.Week, r.SevenDay},
		{"opus", r.SevenDayOpus},
		{"sonnet", r.SevenDaySonn},
	}
	var out []usage.Window
	for _, n := range named {
		if n.l == nil || n.l.Utilization == nil {
			continue
		}
		out = append(out, usage.Window{Name: n.name, UsedPct: *n.l.Utilization, ResetsAt: parseTime(n.l.ResetsAt)})
	}
	return out
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}
