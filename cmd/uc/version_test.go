package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
)

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{name: "unavailable", want: "uc dev"},
		{name: "missing revision", info: &debug.BuildInfo{}, ok: true, want: "uc dev"},
		{name: "clean", info: &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789abcdef"}, {Key: "vcs.modified", Value: "false"}}}, ok: true, want: "uc 0123456789ab"},
		{name: "modified", info: &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.modified", Value: "true"}, {Key: "vcs.revision", Value: "abcdef"}}}, ok: true, want: "uc abcdef-dirty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionFromBuildInfo(tt.info, tt.ok); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersionSkipsInvalidConfig(t *testing.T) {
	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)
	path := filepath.Join(root, "uc", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("invalid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(t.TempDir(), "version-output")
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = stdout }()
	for _, flag := range []string{"--version", "-version"} {
		if err := run([]string{flag}); err != nil {
			t.Errorf("%s: %v", flag, err)
		}
	}
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Repeat(version()+"\n", 2); string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
