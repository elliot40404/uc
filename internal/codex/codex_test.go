package codex

import (
	"context"
	"encoding/base64"
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
	"plan_type": "pro",
	"rate_limit": {
		"primary_window": {"used_percent": 42, "reset_at": 1790000000, "limit_window_seconds": 18000},
		"secondary_window": {"used_percent": 7, "reset_at": 1790500000, "limit_window_seconds": 604800}
	}
}`

func jwt(payload string) string {
	return "h." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".s"
}

func writeAuth(t *testing.T, expires time.Time) string {
	dir := t.TempDir()
	access := jwt(`{"exp":` + strconv.FormatInt(expires.Unix(), 10) +
		`,"https://api.openai.com/auth":{"chatgpt_plan_type":"plus"}}`)
	id := jwt(`{"email":"a@b.c"}`)
	body := `{"tokens":{"id_token":"` + id + `","access_token":"` + access + `","account_id":"acc"}}`
	if err := os.WriteFile(filepath.Join(dir, "auth.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func server(t *testing.T, status int) Client {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wham/usage" || r.Header.Get("ChatGPT-Account-Id") != "acc" ||
			len(r.Header.Get("Authorization")) < 10 {
			t.Errorf("bad request: %s %v", r.URL.Path, r.Header)
		}
		w.WriteHeader(status)
		w.Write([]byte(usageBody))
	}))
	t.Cleanup(srv.Close)
	return Client{HTTP: srv.Client(), BaseURL: srv.URL}
}

func offline() Client {
	return Client{HTTP: &http.Client{}, BaseURL: "http://127.0.0.1:1"}
}

func TestReportOK(t *testing.T) {
	r := server(t, 200).Report(context.Background(), "x", writeAuth(t, now.Add(time.Hour)), now)
	if r.Status != usage.StatusOK || r.Plan != "pro" || r.Email != "a@b.c" {
		t.Fatalf("bad report: %+v", r)
	}
	s, ok := r.Window(usage.Session)
	w, ok2 := r.Window(usage.Week)
	if !ok || !ok2 || s.UsedPct != 42 || w.UsedPct != 7 || s.ResetsAt.Unix() != 1790000000 {
		t.Errorf("bad windows: %+v", r.Windows)
	}
}

func TestReportExpiredSkipsNetwork(t *testing.T) {
	r := offline().Report(context.Background(), "x", writeAuth(t, now.Add(-time.Hour)), now)
	if r.Status != usage.StatusExpired || r.Plan != "plus" || r.Email != "a@b.c" {
		t.Fatalf("want expired with plan and email, got %+v", r)
	}
}

func TestReport401IsExpired(t *testing.T) {
	r := server(t, 401).Report(context.Background(), "x", writeAuth(t, now.Add(time.Hour)), now)
	if r.Status != usage.StatusExpired {
		t.Fatalf("want expired, got %+v", r)
	}
}

func TestReportNoLogin(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"OPENAI_API_KEY":"k","tokens":null}`), 0o600)
	for _, d := range []string{dir, t.TempDir()} {
		if r := offline().Report(context.Background(), "x", d, now); r.Status != usage.StatusNoLogin {
			t.Errorf("want no login, got %+v", r)
		}
	}
}

func TestBadTokenIsError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{"tokens":{"access_token":"junk"}}`), 0o600)
	r := offline().Report(context.Background(), "x", dir, now)
	if r.Status != usage.StatusError || r.Error != "bad token format" {
		t.Fatalf("want error, got %+v", r)
	}
}

func TestWindowName(t *testing.T) {
	cases := map[int64]string{18000: "5h", 604800: "7d", 86400: "1d", 3600: "1h"}
	for s, want := range cases {
		if got := windowName(s); got != want {
			t.Errorf("windowName(%d) = %q, want %q", s, got, want)
		}
	}
}
