package claude

import (
	"encoding/json"
	"errors"
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
	err := jsonfile.Read(filepath.Join(dir, ".credentials.json"), &f)
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
		AccessToken: f.OAuth.AccessToken,
		ExpiresAt:   time.UnixMilli(f.OAuth.ExpiresAt),
		Plan:        f.OAuth.SubscriptionType,
	}, nil
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
