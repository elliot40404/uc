package claude

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

var now = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)

const usageBody = `{
	"five_hour": {"utilization": 42.0, "resets_at": "2026-09-25T01:09:59.694807+00:00"},
	"seven_day": {"utilization": 7.0, "resets_at": "2026-09-26T11:59:59+00:00"},
	"seven_day_opus": null,
	"seven_day_sonnet": {"utilization": 12.0, "resets_at": null}
}`

func writeCreds(t *testing.T, expires time.Time) string {
	dir := t.TempDir()
	body := `{"claudeAiOauth":{"accessToken":"tok","expiresAt":` +
		strconv.FormatInt(expires.UnixMilli(), 10) + `,"subscriptionType":"pro"}}`
	write(t, filepath.Join(dir, ".credentials.json"), body)
	write(t, filepath.Join(dir, ".claude.json"), `{"oauthAccount":{"emailAddress":"a@b.c"}}`)
	return dir
}

func write(t *testing.T, path, body string) {
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func server(t *testing.T, status int) Client {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/oauth/usage" || r.Header.Get("Authorization") != "Bearer tok" ||
			r.Header.Get("anthropic-beta") != "oauth-2025-04-20" {
			t.Errorf("bad request: %s %v", r.URL.Path, r.Header)
		}
		w.WriteHeader(status)
		w.Write([]byte(usageBody))
	}))
	t.Cleanup(srv.Close)
	return Client{HTTP: srv.Client(), BaseURL: srv.URL}
}

func TestReportOK(t *testing.T) {
	r := server(t, 200).Report(context.Background(), "p", writeCreds(t, now.Add(time.Hour)), now)
	if r.Status != usage.StatusOK || r.Plan != "pro" || r.Email != "a@b.c" {
		t.Fatalf("bad report: %+v", r)
	}
	if len(r.Windows) != 3 {
		t.Fatalf("want 3 windows, got %+v", r.Windows)
	}
	s, _ := r.Window(usage.Session)
	if s.UsedPct != 42 || s.ResetsAt.IsZero() {
		t.Errorf("bad session window: %+v", s)
	}
	if r.Windows[2].Name != "sonnet" || !r.Windows[2].ResetsAt.IsZero() {
		t.Errorf("bad sonnet window: %+v", r.Windows[2])
	}
}

func TestReportExpiredSkipsNetwork(t *testing.T) {
	c := Client{HTTP: &http.Client{}, BaseURL: "http://127.0.0.1:1"}
	r := c.Report(context.Background(), "p", writeCreds(t, now.Add(-time.Hour)), now)
	if r.Status != usage.StatusExpired {
		t.Fatalf("want expired, got %+v", r)
	}
}

func TestReport401IsExpired(t *testing.T) {
	r := server(t, 401).Report(context.Background(), "p", writeCreds(t, now.Add(time.Hour)), now)
	if r.Status != usage.StatusExpired {
		t.Fatalf("want expired, got %+v", r)
	}
}

func TestReport500IsError(t *testing.T) {
	r := server(t, 500).Report(context.Background(), "p", writeCreds(t, now.Add(time.Hour)), now)
	if r.Status != usage.StatusError || r.Error != "http 500" {
		t.Fatalf("want error, got %+v", r)
	}
}

func TestReport429IsLimited(t *testing.T) {
	r := server(t, 429).Report(context.Background(), "p", writeCreds(t, now.Add(time.Hour)), now)
	if r.Status != usage.StatusLimited {
		t.Fatalf("want rate limited, got %+v", r)
	}
}

func TestReportNoLogin(t *testing.T) {
	stubKeychain(t, "", usage.ErrNoLogin)
	c := Client{HTTP: &http.Client{}, BaseURL: "http://127.0.0.1:1"}
	r := c.Report(context.Background(), "w", t.TempDir(), now)
	if r.Status != usage.StatusNoLogin {
		t.Fatalf("want no login, got %+v", r)
	}
}

func stubKeychain(t *testing.T, body string, err error) *string {
	var asked string
	old := readKeychain
	readKeychain = func(service string) ([]byte, error) {
		asked = service
		return []byte(body), err
	}
	t.Cleanup(func() { readKeychain = old })
	return &asked
}

func TestLoadCredsFromKeychain(t *testing.T) {
	body := `{"claudeAiOauth":{"accessToken":"kc","expiresAt":1,"subscriptionType":"max"}}`
	asked := stubKeychain(t, body, nil)
	dir := t.TempDir()
	c, err := LoadCreds(dir)
	if err != nil || c.AccessToken != "kc" || c.Plan != "max" {
		t.Fatalf("bad creds: %+v %v", c, err)
	}
	if *asked != keychainService(dir) {
		t.Errorf("asked %q", *asked)
	}
}

func TestLoadCredsFilePreferredOverKeychain(t *testing.T) {
	asked := stubKeychain(t, "", nil)
	c, err := LoadCreds(writeCreds(t, now.Add(time.Hour)))
	if err != nil || c.AccessToken != "tok" || *asked != "" {
		t.Fatalf("bad creds: %+v %v asked %q", c, err, *asked)
	}
}

func TestLoadCredsKeychainMissing(t *testing.T) {
	stubKeychain(t, "", usage.ErrNoLogin)
	if _, err := LoadCreds(t.TempDir()); err != usage.ErrNoLogin {
		t.Fatalf("want no login, got %v", err)
	}
}
