package claude

import (
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/elliot40404/uc/internal/usage"
)

const keychainTimeout = 5 * time.Second

func keychainSecret(service string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), keychainTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/usr/bin/security", "find-generic-password", "-s", service, "-w").Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return nil, usage.ErrNoLogin
	}
	return out, err
}
