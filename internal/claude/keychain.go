package claude

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

const keychainBase = "Claude Code-credentials"

var readKeychain = keychainSecret

func keychainService(dir string) string {
	if isDefaultDir(dir) {
		return keychainBase
	}
	sum := sha256.Sum256([]byte(filepath.Clean(dir)))
	return keychainBase + "-" + hex.EncodeToString(sum[:])[:8]
}

func isDefaultDir(dir string) bool {
	home, err := os.UserHomeDir()
	return err == nil && filepath.Clean(dir) == filepath.Join(home, ".claude")
}
