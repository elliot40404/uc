package codex

import (
	"cmp"
	"context"
	"time"

	"github.com/elliot40404/uc/internal/httpx"
	"github.com/elliot40404/uc/internal/jsonfile"
	"github.com/elliot40404/uc/internal/renew"
	"github.com/elliot40404/uc/internal/usage"
)

const (
	DefaultTokenURL = "https://auth.openai.com/oauth/token"
	clientID        = "app_EMoamEEZ73f0CkXaXp7hrann"
	scope           = "openid profile email"
)

type tokenResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (c Client) renew(ctx context.Context, dir string, now time.Time) (Creds, error) {
	load := func() (Creds, error) { return LoadCreds(dir) }
	return renew.Once(authPath(dir), now, load, func(cur Creds) (Creds, error) {
		return c.refresh(ctx, dir, cur, now)
	})
}

func (c Client) refresh(ctx context.Context, dir string, cur Creds, now time.Time) (Creds, error) {
	if c.TokenURL == "" || cur.RefreshToken == "" {
		return cur, usage.ErrExpired
	}
	body := map[string]string{
		"client_id":     clientID,
		"grant_type":    "refresh_token",
		"refresh_token": cur.RefreshToken,
		"scope":         scope,
	}
	var r tokenResponse
	if err := httpx.PostJSON(ctx, c.HTTP, c.TokenURL, body, &r); err != nil {
		return cur, err
	}
	next, err := fromTokens(Creds{
		AccessToken:  cmp.Or(r.AccessToken, cur.AccessToken),
		RefreshToken: cmp.Or(r.RefreshToken, cur.RefreshToken),
		IDToken:      cmp.Or(r.IDToken, cur.IDToken),
		AccountID:    cur.AccountID,
	})
	if err != nil {
		return cur, err
	}
	patch := map[string]any{
		"tokens": map[string]any{
			"id_token":      next.IDToken,
			"access_token":  next.AccessToken,
			"refresh_token": next.RefreshToken,
		},
		"last_refresh": now.UTC().Format(time.RFC3339Nano),
	}
	return next, jsonfile.Merge(authPath(dir), patch)
}
