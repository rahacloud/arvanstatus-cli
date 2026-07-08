// Package cachet is a thin client for the Cachet v3 status-page JSON:API,
// scoped to the endpoints arvanstatus-cli needs. It targets ArvanCloud's live
// status page at https://www.arvancloudstatus.ir. All exported methods are safe
// to call from multiple goroutines.
package cachet

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// defaultBaseURL is ArvanCloud's live Cachet status page. The old
	// statuspal.io/arvancloud page is stale and no longer matches this site.
	defaultBaseURL = "https://www.arvancloudstatus.ir"
	userAgent      = "arvanstatus-cli/0.1"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }
func WithBaseURL(u string) Option          { return func(c *Client) { c.baseURL = u } }

func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    defaultBaseURL,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// getJSON issues GET {baseURL}/api/{path}?{query} and decodes the JSON body
// into out.
func (c *Client) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	u := fmt.Sprintf("%s/api/%s", c.baseURL, path)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("cachet: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cachet: GET %s: %w", u, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("cachet: GET %s: status %d: %s", u, resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("cachet: decode %s: %w", u, err)
	}
	return nil
}
