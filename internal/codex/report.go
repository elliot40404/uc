package codex

import (
	"cmp"
	"context"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func (c Client) Report(ctx context.Context, name, dir string, now time.Time) usage.Report {
	r := usage.Report{Provider: "codex", Account: name, Dir: dir}
	creds, err := LoadCreds(dir)
	if err != nil {
		return r.Fail(err)
	}
	r.Email, r.Plan = creds.Email, creds.Plan
	if creds.Expired(now) {
		r.Status = usage.StatusExpired
		return r
	}
	resp, err := c.Usage(ctx, creds)
	if err != nil {
		return r.Fail(err)
	}
	r.Plan = cmp.Or(resp.Plan, r.Plan)
	r.Windows = resp.windows()
	r.Status = usage.StatusOK
	return r
}
