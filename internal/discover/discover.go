package discover

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/elliot40404/uc/internal/pathkey"
)

type Account struct {
	Provider string
	Name     string
	Dir      string
}

func Scan(provider, home, prefix, envDir string) []Account {
	var out []Account
	matches, _ := filepath.Glob(filepath.Join(home, prefix+"*"))
	if envDir != "" {
		matches = append(matches, envDir)
	}
	seen := map[string]bool{}
	for _, dir := range matches {
		key := pathkey.Of(dir)
		if seen[key] || !isDir(dir) {
			continue
		}
		seen[key] = true
		out = append(out, Account{Provider: provider, Name: name(dir, prefix), Dir: filepath.Clean(dir)})
	}
	slices.SortFunc(out, func(a, b Account) int { return strings.Compare(a.Name, b.Name) })
	return out
}

func name(dir, prefix string) string {
	base := filepath.Base(dir)
	trimmed := strings.Trim(strings.TrimPrefix(base, prefix), " -_.")
	switch {
	case !strings.HasPrefix(base, prefix):
		return base
	case trimmed == "":
		return "default"
	default:
		return trimmed
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
