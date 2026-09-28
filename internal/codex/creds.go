package codex

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/elliot40404/uc/internal/jsonfile"
	"github.com/elliot40404/uc/internal/usage"
)

const authName = "auth.json"

type Creds struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	AccountID    string
	ExpiresAt    time.Time
	Email        string
	Plan         string
}

type authFile struct {
	Tokens *struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		AccountID    string `json:"account_id"`
	} `json:"tokens"`
}

type claims struct {
	Exp   int64  `json:"exp"`
	Email string `json:"email"`
	Auth  struct {
		Plan string `json:"chatgpt_plan_type"`
	} `json:"https://api.openai.com/auth"`
}

func LoadCreds(dir string) (Creds, error) {
	var f authFile
	if err := jsonfile.Read(authPath(dir), &f); err != nil {
		return Creds{}, err
	}
	if f.Tokens == nil || f.Tokens.AccessToken == "" {
		return Creds{}, usage.ErrNoLogin
	}
	return fromTokens(Creds{
		AccessToken:  f.Tokens.AccessToken,
		RefreshToken: f.Tokens.RefreshToken,
		IDToken:      f.Tokens.IDToken,
		AccountID:    f.Tokens.AccountID,
	})
}

func fromTokens(c Creds) (Creds, error) {
	access, err := parseJWT(c.AccessToken)
	if err != nil {
		return Creds{}, err
	}
	id, _ := parseJWT(c.IDToken)
	c.ExpiresAt = time.Unix(access.Exp, 0)
	c.Email = id.Email
	c.Plan = access.Auth.Plan
	return c, nil
}

func authPath(dir string) string {
	return filepath.Join(dir, authName)
}

func (c Creds) Expired(now time.Time) bool {
	return !c.ExpiresAt.After(now)
}

func parseJWT(token string) (claims, error) {
	var c claims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, errors.New("bad token format")
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return c, errors.New("bad token payload")
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, errors.New("bad token claims")
	}
	return c, nil
}
