package jsonfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMergeKeepsUnknownFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	in := `{"auth":{"token":"old","scopes":["a","b"],"plan":"max"},"other":{"x":1}}`
	if err := os.WriteFile(path, []byte(in), 0o600); err != nil {
		t.Fatal(err)
	}
	err := Merge(path, map[string]any{"auth": map[string]any{"token": "new", "expires": 42}, "stamp": "now"})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	want := `{"auth":{"expires":42,"plan":"max","scopes":["a","b"],"token":"new"},"other":{"x":1},"stamp":"now"}`
	if string(got) != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
}

func TestMergeKeepsIndent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(path, []byte("{\n  \"a\": 1\n}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Merge(path, map[string]any{"b": 2}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "{\n  \"a\": 1,\n  \"b\": 2\n}\n" {
		t.Fatalf("got %q", got)
	}
}

func TestWriteAtomicLeavesNoTemp(t *testing.T) {
	dir := t.TempDir()
	if err := WriteAtomic(filepath.Join(dir, "f.json"), []byte("{}")); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left: %s", e.Name())
		}
	}
}
