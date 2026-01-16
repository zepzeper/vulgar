package httpclient

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// =============================================================================
// Authentication Options
// =============================================================================

// WithBearerToken adds Bearer token authentication
func WithBearerToken(token string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "Bearer " + token
	}
}

// WithAPIKey adds an API key header
func WithAPIKey(headerName, apiKey string) Option {
	return func(c *Client) {
		c.headers[headerName] = apiKey
	}
}

// WithBasicAuth adds HTTP Basic authentication
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		c.headers["Authorization"] = "Basic " + auth
	}
}

// WithAPIKeyQuery adds an API key as a query parameter (less secure, use when required by API)
func WithAPIKeyQuery(paramName, apiKey string) Option {
	return func(c *Client) {
		c.middlewares = append(c.middlewares, func(next RoundTripperFunc) RoundTripperFunc {
			return func(req *http.Request) (*http.Response, error) {
				q := req.URL.Query()
				q.Set(paramName, apiKey)
				req.URL.RawQuery = q.Encode()
				return next(req)
			}
		})
	}
}

// =============================================================================
// OAuth2 Token Management
// =============================================================================

// TokenSource is an interface for providing OAuth2 tokens
type TokenSource interface {
	Token() (*Token, error)
}

// Token represents an OAuth2 token
type Token struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	if t.Expiry.IsZero() {
		return false
	}
	// Consider expired 10 seconds before actual expiry
	return time.Now().After(t.Expiry.Add(-10 * time.Second))
}

// StaticTokenSource returns a TokenSource that always returns the same token
func StaticTokenSource(token *Token) TokenSource {
	return &staticTokenSource{token: token}
}

type staticTokenSource struct {
	token *Token
}

func (s *staticTokenSource) Token() (*Token, error) {
	return s.token, nil
}

// RefreshableTokenSource is a TokenSource that can refresh tokens
type RefreshableTokenSource struct {
	mu           sync.Mutex
	token        *Token
	refreshFunc  func(refreshToken string) (*Token, error)
	onRefresh    func(*Token)
}

// NewRefreshableTokenSource creates a new refreshable token source
func NewRefreshableTokenSource(
	initialToken *Token,
	refreshFunc func(refreshToken string) (*Token, error),
) *RefreshableTokenSource {
	return &RefreshableTokenSource{
		token:       initialToken,
		refreshFunc: refreshFunc,
	}
}

// OnRefresh sets a callback for when the token is refreshed
func (r *RefreshableTokenSource) OnRefresh(fn func(*Token)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRefresh = fn
}

// Token returns a valid token, refreshing if necessary
func (r *RefreshableTokenSource) Token() (*Token, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.token != nil && !r.token.IsExpired() {
		return r.token, nil
	}

	if r.token == nil || r.token.RefreshToken == "" {
		return nil, fmt.Errorf("no token available and cannot refresh")
	}

	newToken, err := r.refreshFunc(r.token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	r.token = newToken

	if r.onRefresh != nil {
		r.onRefresh(newToken)
	}

	return r.token, nil
}

// WithTokenSource adds dynamic token authentication using a TokenSource
func WithTokenSource(ts TokenSource) Option {
	return func(c *Client) {
		c.middlewares = append(c.middlewares, func(next RoundTripperFunc) RoundTripperFunc {
			return func(req *http.Request) (*http.Response, error) {
				token, err := ts.Token()
				if err != nil {
					return nil, fmt.Errorf("failed to get token: %w", err)
				}

				tokenType := token.TokenType
				if tokenType == "" {
					tokenType = "Bearer"
				}

				req.Header.Set("Authorization", tokenType+" "+token.AccessToken)
				return next(req)
			}
		})
	}
}

// =============================================================================
// Common API Authentication Patterns
// =============================================================================

// WithGitHubAuth configures GitHub API authentication
func WithGitHubAuth(token string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "token " + token
		c.headers["Accept"] = "application/vnd.github.v3+json"
	}
}

// WithGitLabAuth configures GitLab API authentication
func WithGitLabAuth(token string) Option {
	return func(c *Client) {
		c.headers["PRIVATE-TOKEN"] = token
	}
}

// WithSlackAuth configures Slack API authentication
func WithSlackAuth(token string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "Bearer " + token
		c.headers["Content-Type"] = "application/json; charset=utf-8"
	}
}

// WithOpenAIAuth configures OpenAI API authentication
func WithOpenAIAuth(apiKey string) Option {
	return func(c *Client) {
		c.headers["Authorization"] = "Bearer " + apiKey
		c.headers["Content-Type"] = "application/json"
	}
}

// WithAnthropicAuth configures Anthropic API authentication
func WithAnthropicAuth(apiKey string) Option {
	return func(c *Client) {
		c.headers["x-api-key"] = apiKey
		c.headers["Content-Type"] = "application/json"
		c.headers["anthropic-version"] = "2023-06-01"
	}
}

// WithStripeAuth configures Stripe API authentication
func WithStripeAuth(secretKey string) Option {
	return func(c *Client) {
		auth := base64.StdEncoding.EncodeToString([]byte(secretKey + ":"))
		c.headers["Authorization"] = "Basic " + auth
	}
}

// WithTwilioAuth configures Twilio API authentication
func WithTwilioAuth(accountSID, authToken string) Option {
	return WithBasicAuth(accountSID, authToken)
}

