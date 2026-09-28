package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elliot40404/uc/internal/usage"
)

const maxBody = 1 << 20

func GetJSON(ctx context.Context, c *http.Client, url string, headers map[string]string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return do(c, req, out)
}

func PostJSON(ctx context.Context, c *http.Client, url string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return do(c, req, out)
}

func do(c *http.Client, req *http.Request, out any) error {
	req.Header.Set("Accept", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body := io.LimitReader(resp.Body, maxBody)
	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return usage.ErrExpired
	case resp.StatusCode == http.StatusTooManyRequests:
		return usage.ErrLimited
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("http %d", resp.StatusCode)
	}
	return json.NewDecoder(body).Decode(out)
}
