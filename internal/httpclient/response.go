package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response wraps an HTTP response with helper methods
type Response struct {
	*http.Response
	body []byte
}

// NewResponse creates a Response from an http.Response
func NewResponse(resp *http.Response) (*Response, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	resp.Body.Close()

	return &Response{
		Response: resp,
		body:     body,
	}, nil
}

// Body returns the response body as bytes
func (r *Response) Body() []byte {
	return r.body
}

// String returns the response body as a string
func (r *Response) String() string {
	return string(r.body)
}

// JSON unmarshals the response body into the given struct
func (r *Response) JSON(v interface{}) error {
	if err := json.Unmarshal(r.body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// JSONMap returns the response body as a map
func (r *Response) JSONMap() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.JSON(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// IsSuccess returns true if the response status code is 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsRedirect returns true if the response status code is 3xx
func (r *Response) IsRedirect() bool {
	return r.StatusCode >= 300 && r.StatusCode < 400
}

// IsClientError returns true if the response status code is 4xx
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if the response status code is 5xx
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500
}

// IsError returns true if the response status code is 4xx or 5xx
func (r *Response) IsError() bool {
	return r.StatusCode >= 400
}

// =============================================================================
// Error Types
// =============================================================================

// APIError represents an error from an API response
type APIError struct {
	StatusCode int
	Status     string
	Body       string
	Message    string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Status)
}

// CheckStatus returns an error if the response is not successful
func (r *Response) CheckStatus() error {
	if r.IsSuccess() {
		return nil
	}

	return &APIError{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Body:       string(r.body),
	}
}

// CheckStatusWithMessage returns an error with a parsed message if not successful
func (r *Response) CheckStatusWithMessage(messageField string) error {
	if r.IsSuccess() {
		return nil
	}

	apiErr := &APIError{
		StatusCode: r.StatusCode,
		Status:     r.Status,
		Body:       string(r.body),
	}

	// Try to extract error message from JSON
	var data map[string]interface{}
	if err := json.Unmarshal(r.body, &data); err == nil {
		if msg, ok := data[messageField].(string); ok {
			apiErr.Message = msg
		} else if msg, ok := data["error"].(string); ok {
			apiErr.Message = msg
		} else if msg, ok := data["message"].(string); ok {
			apiErr.Message = msg
		}
	}

	return apiErr
}

// =============================================================================
// Common API Response Patterns
// =============================================================================

// SlackResponse represents a Slack API response
type SlackResponse struct {
	OK       bool        `json:"ok"`
	Error    string      `json:"error,omitempty"`
	Warning  string      `json:"warning,omitempty"`
	Metadata interface{} `json:"response_metadata,omitempty"`
}

// CheckSlack checks if a Slack API response is successful
func (r *Response) CheckSlack() (*SlackResponse, error) {
	var resp SlackResponse
	if err := r.JSON(&resp); err != nil {
		return nil, fmt.Errorf("failed to parse Slack response: %w", err)
	}

	if !resp.OK {
		return &resp, &APIError{
			StatusCode: r.StatusCode,
			Message:    resp.Error,
		}
	}

	return &resp, nil
}

// GitHubError represents a GitHub API error response
type GitHubError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url,omitempty"`
}

// CheckGitHub checks if a GitHub API response is successful
func (r *Response) CheckGitHub() error {
	if r.IsSuccess() {
		return nil
	}

	var ghErr GitHubError
	if err := r.JSON(&ghErr); err == nil && ghErr.Message != "" {
		return &APIError{
			StatusCode: r.StatusCode,
			Message:    ghErr.Message,
		}
	}

	return r.CheckStatus()
}

// GitLabError represents a GitLab API error response
type GitLabError struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// CheckGitLab checks if a GitLab API response is successful
func (r *Response) CheckGitLab() error {
	if r.IsSuccess() {
		return nil
	}

	var glErr GitLabError
	if err := r.JSON(&glErr); err == nil {
		msg := glErr.Error
		if msg == "" {
			msg = glErr.Message
		}
		if msg != "" {
			return &APIError{
				StatusCode: r.StatusCode,
				Message:    msg,
			}
		}
	}

	return r.CheckStatus()
}
