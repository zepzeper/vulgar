package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	headers     map[string]string
	middlewares []Middleware
	mu          sync.RWMutex

	maxRetries    int
	retryWaitMin  time.Duration
	retryWaitMax  time.Duration
	retryOnStatus []int

	rateLimiter *rateLimiter

	debug bool
}

type Middleware func(next RoundTripperFunc) RoundTripperFunc

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers:       make(map[string]string),
		middlewares:   make([]Middleware, 0),
		maxRetries:    0,
		retryWaitMin:  1 * time.Second,
		retryWaitMax:  30 * time.Second,
		retryOnStatus: []int{429, 500, 502, 503, 504},
	}

	// Set default User-Agent
	c.headers["User-Agent"] = "vulgar/1.0"

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Clone creates a copy of the client with the same configuration
func (c *Client) Clone() *Client {
	c.mu.RLock()
	defer c.mu.RUnlock()

	headers := make(map[string]string)
	for k, v := range c.headers {
		headers[k] = v
	}

	middlewares := make([]Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)

	return &Client{
		httpClient:    c.httpClient,
		baseURL:       c.baseURL,
		headers:       headers,
		middlewares:   middlewares,
		maxRetries:    c.maxRetries,
		retryWaitMin:  c.retryWaitMin,
		retryWaitMax:  c.retryWaitMax,
		retryOnStatus: c.retryOnStatus,
		rateLimiter:   c.rateLimiter,
		debug:         c.debug,
	}
}

// With returns a new client with additional options applied
func (c *Client) With(opts ...Option) *Client {
	clone := c.Clone()
	for _, opt := range opts {
		opt(clone)
	}
	return clone
}

// Do executes an HTTP request with all middleware applied
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Apply default headers
	for k, v := range c.headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}

	// Build the middleware chain
	var handler RoundTripperFunc = c.httpClient.Do

	// Apply middlewares in reverse order (first added = outermost)
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i](handler)
	}

	// Apply retry middleware if configured
	if c.maxRetries > 0 {
		handler = c.retryMiddleware(handler)
	}

	// Apply rate limiting if configured
	if c.rateLimiter != nil {
		handler = c.rateLimitMiddleware(handler)
	}

	return handler(req)
}

// Request creates and executes an HTTP request
func (c *Client) Request(ctx context.Context, method, urlPath string, body io.Reader) (*Response, error) {
	fullURL := c.buildURL(urlPath)

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return NewResponse(resp)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, urlPath string) (*Response, error) {
	return c.Request(ctx, http.MethodGet, urlPath, nil)
}

// Post performs a POST request with a body
func (c *Client) Post(ctx context.Context, urlPath string, body io.Reader) (*Response, error) {
	return c.Request(ctx, http.MethodPost, urlPath, body)
}

// Put performs a PUT request with a body
func (c *Client) Put(ctx context.Context, urlPath string, body io.Reader) (*Response, error) {
	return c.Request(ctx, http.MethodPut, urlPath, body)
}

// Patch performs a PATCH request with a body
func (c *Client) Patch(ctx context.Context, urlPath string, body io.Reader) (*Response, error) {
	return c.Request(ctx, http.MethodPatch, urlPath, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, urlPath string) (*Response, error) {
	return c.Request(ctx, http.MethodDelete, urlPath, nil)
}

// GetJSON performs a GET request and unmarshals the JSON response
func (c *Client) GetJSON(ctx context.Context, urlPath string, result interface{}) error {
	resp, err := c.Get(ctx, urlPath)
	if err != nil {
		return err
	}
	return resp.JSON(result)
}

// PostJSON performs a POST request with a JSON body and unmarshals the response
func (c *Client) PostJSON(ctx context.Context, urlPath string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(urlPath), bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	response, err := NewResponse(resp)
	if err != nil {
		return err
	}

	if result != nil {
		return response.JSON(result)
	}
	return nil
}

// PutJSON performs a PUT request with a JSON body and unmarshals the response
func (c *Client) PutJSON(ctx context.Context, urlPath string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL(urlPath), bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	response, err := NewResponse(resp)
	if err != nil {
		return err
	}

	if result != nil {
		return response.JSON(result)
	}
	return nil
}

// PatchJSON performs a PATCH request with a JSON body and unmarshals the response
func (c *Client) PatchJSON(ctx context.Context, urlPath string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.buildURL(urlPath), bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	response, err := NewResponse(resp)
	if err != nil {
		return err
	}

	if result != nil {
		return response.JSON(result)
	}
	return nil
}

// PostForm performs a POST request with form data
func (c *Client) PostForm(ctx context.Context, urlPath string, data url.Values) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(urlPath), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return NewResponse(resp)
}

// buildURL combines the base URL with a path
func (c *Client) buildURL(path string) string {
	if c.baseURL == "" {
		return path
	}

	// If path is already a full URL, return it
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	// Combine base URL and path
	base := strings.TrimSuffix(c.baseURL, "/")
	path = strings.TrimPrefix(path, "/")

	return base + "/" + path
}

