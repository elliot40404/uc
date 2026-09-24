package render

import (
	"strings"
	"testing"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

func TestJSONHasLeftAndNoSecrets(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	reports := []usage.Report{
		{Provider: "codex", Account: "a", Status: usage.StatusOK, Windows: []usage.Window{{Name: usage.Week, UsedPct: 25}}},
		{Provider: "codex", Account: "b", Status: usage.StatusExpired},
	}
	var b strings.Builder
	if err := JSON(&b, reports, now); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{`"left_pct": 75`, `"left_pct": null`, `"status": "expired"`, `"generated_at": "2026-01-01T00:00:00Z"`} {
		if !strings.Contains(out, want) {
			t.Errorf("json missing %s:\n%s", want, out)
		}
	}
	if strings.Contains(out, `"resets_at"`) {
		t.Errorf("zero reset time should be omitted:\n%s", out)
	}
}
