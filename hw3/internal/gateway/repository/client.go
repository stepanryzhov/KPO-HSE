package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

type client struct {
	baseURL     string
	httpClient  *http.Client
	unavailable error
}

func newClient(baseURL string, httpClient *http.Client, unavailable error) client {
	return client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient, unavailable: unavailable}
}

func (c client) sendJSON(ctx context.Context, method, path string, payload, result any) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("%w: encode %s request: %w", c.unavailable, path, err)
	}
	return c.request(ctx, method, path, "application/json", bytes.NewReader(body), result, http.StatusOK)
}

func (c client) request(ctx context.Context, method, path, contentType string, body io.Reader, result any, expectedStatuses ...int) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return 0, fmt.Errorf("%w: create %s request: %w", c.unavailable, path, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %s %s: %w", c.unavailable, method, path, err)
	}
	defer resp.Body.Close()

	if !slices.Contains(expectedStatuses, resp.StatusCode) {
		return resp.StatusCode, fmt.Errorf("%w: %s %s: HTTP %d", c.unavailable, method, path, resp.StatusCode)
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(result); err != nil {
		return resp.StatusCode, fmt.Errorf("%w: decode %s response: %w", c.unavailable, path, err)
	}
	if err := decoder.Decode(new(json.RawMessage)); err != io.EOF {
		if err != nil {
			return resp.StatusCode, fmt.Errorf("%w: read %s response: %w", c.unavailable, path, err)
		}
		return resp.StatusCode, fmt.Errorf("%w: expected one JSON value in %s response", c.unavailable, path)
	}
	return resp.StatusCode, nil
}
