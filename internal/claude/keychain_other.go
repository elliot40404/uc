//go:build !darwin

package claude

import "github.com/elliot40404/uc/internal/usage"

func keychainSecret(string) ([]byte, error) {
	return nil, usage.ErrNoLogin
}
