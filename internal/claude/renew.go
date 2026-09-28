package claude

import (
	"cmp"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/httpx"
	"github.com/elliot40404/uc/internal/jsonfile"
	"github.com/elliot40404/uc/internal/renew"
	"github.com/elliot40404/uc/internal/usage"
)

const (
	DefaultTokenURL = "https://platform.claude.com/v1/oauth/token"
	clientID        = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
)

var errBadToken = errors.New("bad token response")

type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_token_expires_in"`
}

func (c Client) renew(ctx context.Context, dir string, now time.Time) (Creds, error) {
	load := func() (Creds, error) { return LoadCreds(dir) }
	return renew.Once(credsPath(dir), now, load, func(cur Creds) (Creds, error) {
		return c.refresh(ctx, dir, cur, now)
	})
}

func (c Client) refresh(ctx context.Context, dir string, cur Creds, now time.Time) (Creds, error) {
	if c.TokenURL == "" || !cur.Renewable(now) {
		return cur, usage.ErrExpired
	}
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": cur.RefreshToken,
		"client_id":     clientID,
		"scope":         strings.Join(cur.Scopes, " "),
	}
	var r tokenResponse
	if err := httpx.PostJSON(ctx, c.HTTP, c.TokenURL, body, &r); err != nil {
		return cur, err
	}
	if r.AccessToken == "" || r.ExpiresIn <= 0 {
		return cur, errBadToken
	}
	cur.AccessToken = r.AccessToken
	cur.RefreshToken = cmp.Or(r.RefreshToken, cur.RefreshToken)
	cur.ExpiresAt = now.Add(time.Duration(r.ExpiresIn) * time.Second)
	fields := map[string]any{
		"accessToken":  cur.AccessToken,
		"refreshToken": cur.RefreshToken,
		"expiresAt":    cur.ExpiresAt.UnixMilli(),
	}
	if r.RefreshExpiresIn > 0 {
		cur.RefreshExpiresAt = now.Add(time.Duration(r.RefreshExpiresIn) * time.Second)
		fields["refreshTokenExpiresAt"] = cur.RefreshExpiresAt.UnixMilli()
	}
	return cur, jsonfile.Merge(credsPath(dir), map[string]any{"claudeAiOauth": fields})
}
