package codex

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/elliot40404/uc/internal/httpx"
	"github.com/elliot40404/uc/internal/usage"
)

const DefaultBaseURL = "https://chatgpt.com/backend-api"

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

type window struct {
	UsedPct float64 `json:"used_percent"`
	ResetAt int64   `json:"reset_at"`
	Seconds int64   `json:"limit_window_seconds"`
}

type usageResponse struct {
	Plan      string `json:"plan_type"`
	RateLimit *struct {
		Primary   *window `json:"primary_window"`
		Secondary *window `json:"secondary_window"`
	} `json:"rate_limit"`
}

func (c Client) Usage(ctx context.Context, creds Creds) (usageResponse, error) {
	var r usageResponse
	headers := map[string]string{
		"Authorization":      "Bearer " + creds.AccessToken,
		"ChatGPT-Account-Id": creds.AccountID,
	}
	err := httpx.GetJSON(ctx, c.HTTP, c.BaseURL+"/wham/usage", headers, &r)
	return r, err
}

func (r usageResponse) windows() []usage.Window {
	if r.RateLimit == nil {
		return nil
	}
	var out []usage.Window
	for _, w := range []*window{r.RateLimit.Primary, r.RateLimit.Secondary} {
		if w != nil {
			out = append(out, w.toUsage())
		}
	}
	return out
}

func (w window) toUsage() usage.Window {
	out := usage.Window{Name: windowName(w.Seconds), UsedPct: w.UsedPct}
	if w.ResetAt > 0 {
		out.ResetsAt = time.Unix(w.ResetAt, 0)
	}
	return out
}

func windowName(seconds int64) string {
	d := time.Duration(seconds) * time.Second
	switch {
	case d == 5*time.Hour:
		return usage.Session
	case d == 7*24*time.Hour:
		return usage.Week
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
}
