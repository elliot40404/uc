package render

import (
	"encoding/json"
	"io"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

type jsonReport struct {
	usage.Report
	LeftPct *float64 `json:"left_pct"`
}

type jsonOutput struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Accounts    []jsonReport `json:"accounts"`
}

func JSON(w io.Writer, reports []usage.Report, now time.Time) error {
	out := jsonOutput{GeneratedAt: now.UTC(), Accounts: []jsonReport{}}
	for _, r := range reports {
		out.Accounts = append(out.Accounts, toJSON(r))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func toJSON(r usage.Report) jsonReport {
	out := jsonReport{Report: r}
	if left := r.Left(); left >= 0 {
		out.LeftPct = &left
	}
	return out
}
