package claude

import (
	"context"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func (c Client) Report(ctx context.Context, name, dir string, now time.Time) usage.Report {
	r := usage.Report{Provider: "claude", Account: name, Dir: dir, Email: Email(dir)}
	creds, err := LoadCreds(dir)
	if err != nil {
		return r.Fail(err)
	}
	r.Plan = creds.Plan
	if creds.Expired(now) {
		r.Status = usage.StatusExpired
		return r
	}
	r.Windows, err = c.Usage(ctx, creds.AccessToken)
	if err != nil {
		return r.Fail(err)
	}
	r.Status = usage.StatusOK
	return r
}
