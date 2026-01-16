package httpclient

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client
type Option func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBaseURL sets the base URL for all requests
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHeader adds a default header to all requests
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithHeaders adds multiple default headers to all requests
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithUserAgent sets the User-Agent header
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.headers["User-Agent"] = userAgent
	}
}

// WithMiddleware adds a middleware to the client
func WithMiddleware(m Middleware) Option {
	return func(c *Client) {
		c.middlewares = append(c.middlewares, m)
	}
}

// WithRetry configures retry behavior
func WithRetry(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryConfig configures detailed retry behavior
func WithRetryConfig(maxRetries int, waitMin, waitMax time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryWaitMin = waitMin
		c.retryWaitMax = waitMax
	}
}

// WithRetryOnStatus configures which status codes trigger a retry
func WithRetryOnStatus(codes ...int) Option {
	return func(c *Client) {
		c.retryOnStatus = codes
	}
}

// WithRateLimit configures rate limiting
func WithRateLimit(requestsPerSecond float64) Option {
	return func(c *Client) {
		c.rateLimiter = newRateLimiter(requestsPerSecond)
	}
}

// WithRateLimitBurst configures rate limiting with burst
func WithRateLimitBurst(requestsPerSecond float64, burst int) Option {
	return func(c *Client) {
		c.rateLimiter = newRateLimiterWithBurst(requestsPerSecond, burst)
	}
}

// WithDebug enables debug logging
func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

// WithHTTPClient sets a custom underlying http.Client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTransport sets a custom http.RoundTripper
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		c.httpClient.Transport = transport
	}
}

// WithNoRedirects disables following redirects
func WithNoRedirects() Option {
	return func(c *Client) {
		c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

// WithMaxRedirects sets the maximum number of redirects to follow
func WithMaxRedirects(max int) Option {
	return func(c *Client) {
		c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= max {
				return http.ErrUseLastResponse
			}
			return nil
		}
	}
}

