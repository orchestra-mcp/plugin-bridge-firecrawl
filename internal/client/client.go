// Package client provides the Firecrawl HTTP client used by all tool handlers.
// It calls the Firecrawl REST API (https://api.firecrawl.dev/v1) using plain
// net/http. Authentication is via Bearer token from FIRECRAWL_API_KEY.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DefaultHost is the default Firecrawl API base URL.
const DefaultHost = "https://api.firecrawl.dev"

// FirecrawlClient holds the configuration needed to call the Firecrawl API.
type FirecrawlClient struct {
	Host   string // Base URL (default: https://api.firecrawl.dev)
	APIKey string // Bearer token for Authorization header
}

// NewFirecrawlClient creates a client from the plugin's environment map.
// It reads FIRECRAWL_API_KEY (required) and FIRECRAWL_BASE_URL (optional).
func NewFirecrawlClient(env map[string]string) *FirecrawlClient {
	host := DefaultHost
	if v, ok := env["FIRECRAWL_BASE_URL"]; ok && v != "" {
		host = v
	}
	return &FirecrawlClient{
		Host:   host,
		APIKey: env["FIRECRAWL_API_KEY"],
	}
}

// Post sends a POST request to the given Firecrawl API path with a JSON body
// and decodes the response into dest. The path should start with "/" (e.g.,
// "/v1/scrape").
func (c *FirecrawlClient) Post(ctx context.Context, path string, body any, dest any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Host+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	return c.doRequest(req, dest)
}

// Get sends a GET request to the given Firecrawl API path and decodes the
// response into dest.
func (c *FirecrawlClient) Get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.Host+path, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	return c.doRequest(req, dest)
}

// setHeaders adds the Authorization and Content-Type headers.
func (c *FirecrawlClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
}

// doRequest executes the HTTP request, checks for errors, and decodes the
// response body into dest.
func (c *FirecrawlClient) doRequest(req *http.Request, dest any) error {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("firecrawl request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firecrawl error %d: %s", resp.StatusCode, string(respBody))
	}

	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
