package jsonfile

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/elliot40404/uc/internal/usage"
)

func Read(path string, out any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return usage.ErrNoLogin
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
