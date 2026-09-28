package codex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func accessJWT(expires time.Time) string {
	return jwt(`{"exp":` + strconv.FormatInt(expires.Unix(), 10) +
		`,"https://api.openai.com/auth":{"chatgpt_plan_type":"pro"}}`)
}

func writeExpiredAuth(t *testing.T) string {
	dir := t.TempDir()
	body := "{\n  \"auth_mode\": \"chatgpt\",\n  \"tokens\": {\"id_token\": \"" + jwt(`{"email":"a@b.c"}`) +
		"\", \"access_token\": \"" + accessJWT(now.Add(-time.Hour)) +
		"\", \"refresh_token\": \"r1\", \"account_id\": \"acc\"}\n}\n"
	if err := os.WriteFile(filepath.Join(dir, authName), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func renewServer(t *testing.T, tokenStatus int, fresh string) (Client, *int) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			calls++
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["grant_type"] != "refresh_token" || body["refresh_token"] != "r1" ||
				body["client_id"] != clientID || body["scope"] != scope {
				t.Errorf("bad token request: %v", body)
			}
			w.WriteHeader(tokenStatus)
			json.NewEncoder(w).Encode(map[string]string{"access_token": fresh, "refresh_token": "r2"})
		case "/wham/usage":
			if r.Header.Get("Authorization") != "Bearer "+fresh {
				t.Errorf("usage called with old token")
			}
			w.Write([]byte(usageBody))
		}
	}))
	t.Cleanup(srv.Close)
	return Client{HTTP: srv.Client(), BaseURL: srv.URL, TokenURL: srv.URL + "/token"}, &calls
}

func TestReportRenewsExpiredLogin(t *testing.T) {
	dir := writeExpiredAuth(t)
	fresh := accessJWT(now.Add(time.Hour))
	c, calls := renewServer(t, 200, fresh)
	r := c.Report(context.Background(), "c", dir, now)
	if r.Status != usage.StatusOK || *calls != 1 {
		t.Fatalf("status %s calls %d", r.Status, *calls)
	}
	creds, err := LoadCreds(dir)
	if err != nil || creds.AccessToken != fresh || creds.RefreshToken != "r2" || creds.Email != "a@b.c" {
		t.Fatalf("saved creds: %+v %v", creds, err)
	}
	var f map[string]any
	raw, _ := os.ReadFile(filepath.Join(dir, authName))
	json.Unmarshal(raw, &f)
	if f["auth_mode"] != "chatgpt" || f["last_refresh"] == nil {
		t.Fatalf("file fields: %s", raw)
	}
}

func TestReportRenewFailureIsExpired(t *testing.T) {
	dir := writeExpiredAuth(t)
	before, _ := os.ReadFile(filepath.Join(dir, authName))
	c, _ := renewServer(t, 401, accessJWT(now.Add(time.Hour)))
	r := c.Report(context.Background(), "c", dir, now)
	if r.Status != usage.StatusExpired {
		t.Fatalf("want expired, got %+v", r)
	}
	after, _ := os.ReadFile(filepath.Join(dir, authName))
	if string(before) != string(after) {
		t.Fatal("auth changed after failed renew")
	}
}
