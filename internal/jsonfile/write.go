package jsonfile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	renameTries = 5
	renameWait  = 50 * time.Millisecond
)

func Merge(path string, patch map[string]any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return err
	}
	for k, v := range patch {
		if fields, ok := v.(map[string]any); ok {
			if v, err = mergeObject(top[k], fields); err != nil {
				return err
			}
		}
		if top[k], err = json.Marshal(v); err != nil {
			return err
		}
	}
	out, err := encode(top, bytes.ContainsRune(raw, '\n'))
	if err != nil {
		return err
	}
	return WriteAtomic(path, out)
}

func mergeObject(raw json.RawMessage, fields map[string]any) (map[string]json.RawMessage, error) {
	obj := map[string]json.RawMessage{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, err
		}
	}
	for k, v := range fields {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		obj[k] = b
	}
	return obj, nil
}

func encode(v any, indent bool) ([]byte, error) {
	if !indent {
		return json.Marshal(v)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	return append(out, '\n'), err
}

func WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return rename(f.Name(), path)
}

func rename(from, to string) error {
	var err error
	for range renameTries {
		if err = os.Rename(from, to); err == nil {
			return nil
		}
		time.Sleep(renameWait)
	}
	return err
}
