package claude

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func writeExpiredCreds(t *testing.T) string {
	dir := t.TempDir()
	body := `{"claudeAiOauth":{"accessToken":"old","refreshToken":"r1","expiresAt":` +
		strconv.FormatInt(now.Add(-time.Hour).UnixMilli(), 10) +
		`,"scopes":["user:inference","user:profile"],"subscriptionType":"max"},"mcpOAuth":{"k":"v"}}`
	write(t, filepath.Join(dir, credsName), body)
	return dir
}

func renewServer(t *testing.T, tokenStatus int) (Client, *int) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			calls++
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["grant_type"] != "refresh_token" || body["refresh_token"] != "r1" ||
				body["client_id"] != clientID || body["scope"] != "user:inference user:profile" {
				t.Errorf("bad token request: %v", body)
			}
			w.WriteHeader(tokenStatus)
			w.Write([]byte(`{"access_token":"tok","refresh_token":"r2","expires_in":3600}`))
		case "/api/oauth/usage":
			if r.Header.Get("Authorization") != "Bearer tok" {
				t.Errorf("usage called with %q", r.Header.Get("Authorization"))
			}
			w.Write([]byte(usageBody))
		}
	}))
	t.Cleanup(srv.Close)
	return Client{HTTP: srv.Client(), BaseURL: srv.URL, TokenURL: srv.URL + "/token"}, &calls
}

func TestReportRenewsExpiredLogin(t *testing.T) {
	dir := writeExpiredCreds(t)
	c, calls := renewServer(t, 200)
	r := c.Report(context.Background(), "p", dir, now)
	if r.Status != usage.StatusOK || *calls != 1 {
		t.Fatalf("status %s calls %d", r.Status, *calls)
	}
	creds, err := LoadCreds(dir)
	if err != nil || creds.AccessToken != "tok" || creds.RefreshToken != "r2" || creds.Plan != "max" {
		t.Fatalf("saved creds: %+v %v", creds, err)
	}
	if !creds.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("expiry %v", creds.ExpiresAt)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, credsName))
	if !strings.Contains(string(raw), `"mcpOAuth":{"k":"v"}`) {
		t.Fatalf("other fields lost: %s", raw)
	}
}

func TestReportRenewFailureIsExpired(t *testing.T) {
	dir := writeExpiredCreds(t)
	before, _ := os.ReadFile(filepath.Join(dir, credsName))
	c, _ := renewServer(t, 400)
	r := c.Report(context.Background(), "p", dir, now)
	if r.Status != usage.StatusExpired {
		t.Fatalf("want expired, got %+v", r)
	}
	after, _ := os.ReadFile(filepath.Join(dir, credsName))
	if string(before) != string(after) {
		t.Fatal("creds changed after failed renew")
	}
}

func TestReportSkipsRenewWithoutRefreshToken(t *testing.T) {
	c, calls := renewServer(t, 200)
	r := c.Report(context.Background(), "p", writeCreds(t, now.Add(-time.Hour)), now)
	if r.Status != usage.StatusExpired || *calls != 0 {
		t.Fatalf("status %s calls %d", r.Status, *calls)
	}
}

func TestKeychainLoginIsNotRenewable(t *testing.T) {
	stubKeychain(t, `{"claudeAiOauth":{"accessToken":"kc","refreshToken":"r","expiresAt":1}}`, nil)
	creds, err := LoadCreds(t.TempDir())
	if err != nil || creds.Renewable(now) {
		t.Fatalf("keychain creds renewable: %+v %v", creds, err)
	}
}
