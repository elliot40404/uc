package claude

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"

	"github.com/elliot40404/uc/internal/jsonfile"
	"github.com/elliot40404/uc/internal/usage"
)

const credsName = ".credentials.json"

type Creds struct {
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time
	RefreshExpiresAt time.Time
	Scopes           []string
	Plan             string
	InFile           bool
}

type credsFile struct {
	OAuth *struct {
		AccessToken      string   `json:"accessToken"`
		RefreshToken     string   `json:"refreshToken"`
		ExpiresAt        int64    `json:"expiresAt"`
		RefreshExpiresAt int64    `json:"refreshTokenExpiresAt"`
		Scopes           []string `json:"scopes"`
		SubscriptionType string   `json:"subscriptionType"`
	} `json:"claudeAiOauth"`
}

func LoadCreds(dir string) (Creds, error) {
	var f credsFile
	err := jsonfile.Read(credsPath(dir), &f)
	inFile := err == nil
	if errors.Is(err, usage.ErrNoLogin) {
		err = keychainCreds(dir, &f)
	}
	if err != nil {
		return Creds{}, err
	}
	if f.OAuth == nil || f.OAuth.AccessToken == "" {
		return Creds{}, usage.ErrNoLogin
	}
	return Creds{
		AccessToken:      f.OAuth.AccessToken,
		RefreshToken:     f.OAuth.RefreshToken,
		ExpiresAt:        time.UnixMilli(f.OAuth.ExpiresAt),
		RefreshExpiresAt: unixMilliOrZero(f.OAuth.RefreshExpiresAt),
		Scopes:           f.OAuth.Scopes,
		Plan:             f.OAuth.SubscriptionType,
		InFile:           inFile,
	}, nil
}

func credsPath(dir string) string {
	return filepath.Join(dir, credsName)
}

func unixMilliOrZero(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func keychainCreds(dir string, f *credsFile) error {
	data, err := readKeychain(keychainService(dir))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, f)
}

func (c Creds) Expired(now time.Time) bool {
	return !c.ExpiresAt.After(now)
}

func (c Creds) Renewable(now time.Time) bool {
	return c.InFile && c.RefreshToken != "" && (c.RefreshExpiresAt.IsZero() || c.RefreshExpiresAt.After(now))
}

type accountFile struct {
	OAuthAccount struct {
		Email string `json:"emailAddress"`
	} `json:"oauthAccount"`
}

func Email(dir string) string {
	var f accountFile
	if jsonfile.Read(accountPath(dir), &f) != nil {
		return ""
	}
	return f.OAuthAccount.Email
}

func accountPath(dir string) string {
	if isDefaultDir(dir) {
		return filepath.Join(filepath.Dir(filepath.Clean(dir)), ".claude.json")
	}
	return filepath.Join(dir, ".claude.json")
}
