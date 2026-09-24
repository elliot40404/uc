package claude

import (
	"os"
	"path/filepath"
	"time"

	"github.com/elliot40404/uc/internal/jsonfile"
	"github.com/elliot40404/uc/internal/usage"
)

type Creds struct {
	AccessToken string
	ExpiresAt   time.Time
	Plan        string
}

type credsFile struct {
	OAuth *struct {
		AccessToken      string `json:"accessToken"`
		ExpiresAt        int64  `json:"expiresAt"`
		SubscriptionType string `json:"subscriptionType"`
	} `json:"claudeAiOauth"`
}

func LoadCreds(dir string) (Creds, error) {
	var f credsFile
	if err := jsonfile.Read(filepath.Join(dir, ".credentials.json"), &f); err != nil {
		return Creds{}, err
	}
	if f.OAuth == nil || f.OAuth.AccessToken == "" {
		return Creds{}, usage.ErrNoLogin
	}
	return Creds{
		AccessToken: f.OAuth.AccessToken,
		ExpiresAt:   time.UnixMilli(f.OAuth.ExpiresAt),
		Plan:        f.OAuth.SubscriptionType,
	}, nil
}

func (c Creds) Expired(now time.Time) bool {
	return !c.ExpiresAt.After(now)
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
	home, err := os.UserHomeDir()
	if err == nil && filepath.Clean(dir) == filepath.Join(home, ".claude") {
		return filepath.Join(home, ".claude.json")
	}
	return filepath.Join(dir, ".claude.json")
}
