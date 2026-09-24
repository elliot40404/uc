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

type Creds struct {
	AccessToken string
	AccountID   string
	ExpiresAt   time.Time
	Email       string
	Plan        string
}

type authFile struct {
	Tokens *struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
		AccountID   string `json:"account_id"`
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
	if err := jsonfile.Read(filepath.Join(dir, "auth.json"), &f); err != nil {
		return Creds{}, err
	}
	if f.Tokens == nil || f.Tokens.AccessToken == "" {
		return Creds{}, usage.ErrNoLogin
	}
	access, err := parseJWT(f.Tokens.AccessToken)
	if err != nil {
		return Creds{}, err
	}
	id, _ := parseJWT(f.Tokens.IDToken)
	return Creds{
		AccessToken: f.Tokens.AccessToken,
		AccountID:   f.Tokens.AccountID,
		ExpiresAt:   time.Unix(access.Exp, 0),
		Email:       id.Email,
		Plan:        access.Auth.Plan,
	}, nil
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
