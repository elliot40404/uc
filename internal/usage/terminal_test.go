package usage

import (
	"strings"
	"testing"
)

func TestTerminalReportSanitizesFieldsWithoutChangingInput(t *testing.T) {
	in := Report{
		Provider: "claude\x1b[31m", Account: "one\nother", Dir: "/a\r/b",
		Plan: "bad\x1b[0m", Email: "mail\x00@test", Error: "error\u202ehello",
		Windows: []Window{{Name: "5h\x1b[2J"}},
	}
	out := TerminalReport(in)
	for _, field := range []string{out.Provider, out.Account, out.Dir, out.Plan, out.Email, out.Error, out.Windows[0].Name} {
		if strings.ContainsAny(field, "\x1b\n\r\x00\u202e") {
			t.Fatalf("unsafe terminal field %q", field)
		}
	}
	if in.Windows[0].Name != "5h\x1b[2J" || in.Account != "one\nother" {
		t.Fatal("input was changed")
	}
}
