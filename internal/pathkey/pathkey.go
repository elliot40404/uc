package pathkey

import (
	"path/filepath"
	"runtime"
	"strings"
)

func Of(path string) string {
	path = filepath.Clean(path)
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}
