package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client := New()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}
}

func TestWithTimeout(t *testing.T) {
	client := New(WithTimeout(5 * time.Second))
	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout of 5s, got %v", client.httpClient.Timeout)
	}
}

func TestWithBaseURL(t *testing.T) {
	client := New(WithBaseURL("https://api.example.com"))
	if client.baseURL != "https://api.example.com" {
		t.Errorf("expected base URL https://api.example.com, got %s", client.baseURL)
	}
}

func TestWithHeaders(t *testing.T) {
	client := New(
		WithHeader("X-Custom", "value1"),
		WithHeaders(map[string]string{
			"X-Another": "value2",
		}),
	)

	if client.headers["X-Custom"] != "value1" {
		t.Errorf("expected X-Custom header, got %s", client.headers["X-Custom"])
	}
	if client.headers["X-Another"] != "value2" {
		t.Errorf("expected X-Another header, got %s", client.headers["X-Another"])
	}
}

func TestWithBearerToken(t *testing.T) {
	client := New(WithBearerToken("test-token"))
	if client.headers["Authorization"] != "Bearer test-token" {
		t.Errorf("expected Bearer token header, got %s", client.headers["Authorization"])
	}
}

func TestWithBasicAuth(t *testing.T) {
	client := New(WithBasicAuth("user", "pass"))
	if client.headers["Authorization"] == "" {
		t.Error("expected Authorization header to be set")
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
	}{
		{
			name:     "no base URL",
			baseURL:  "",
			path:     "https://api.example.com/test",
			expected: "https://api.example.com/test",
		},
		{
			name:     "with base URL and relative path",
			baseURL:  "https://api.example.com",
			path:     "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "with base URL trailing slash",
			baseURL:  "https://api.example.com/",
			path:     "users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "absolute URL ignores base",
			baseURL:  "https://api.example.com",
			path:     "https://other.com/test",
			expected: "https://other.com/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(WithBaseURL(tt.baseURL))
			result := client.buildURL(tt.path)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := New()
	resp, err := client.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.IsSuccess() {
		t.Errorf("expected success response, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := resp.JSON(&result); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %s", result["status"])
	}
}

func TestPostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"received": body["message"],
		})
	}))
	defer server.Close()

	client := New()
	var result map[string]string
	err := client.PostJSON(context.Background(), server.URL, map[string]string{
		"message": "hello",
	}, &result)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["received"] != "hello" {
		t.Errorf("expected received=hello, got %s", result["received"])
	}
}

func TestRequestBuilder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query params
		if r.URL.Query().Get("page") != "1" {
			t.Errorf("expected page=1 query param")
		}
		// Check headers
		if r.Header.Get("X-Request-ID") != "test-123" {
			t.Errorf("expected X-Request-ID header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New()
	resp, err := client.NewRequest("GET", server.URL).
		Query("page", "1").
		Header("X-Request-ID", "test-123").
		Do()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		statusCode      int
		isSuccess       bool
		isClientError   bool
		isServerError   bool
		isError         bool
	}{
		{200, true, false, false, false},
		{201, true, false, false, false},
		{301, false, false, false, false},
		{400, false, true, false, true},
		{404, false, true, false, true},
		{500, false, false, true, true},
		{503, false, false, true, true},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.statusCode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := New()
			resp, _ := client.Get(context.Background(), server.URL)

			if resp.IsSuccess() != tt.isSuccess {
				t.Errorf("IsSuccess: expected %v, got %v", tt.isSuccess, resp.IsSuccess())
			}
			if resp.IsClientError() != tt.isClientError {
				t.Errorf("IsClientError: expected %v, got %v", tt.isClientError, resp.IsClientError())
			}
			if resp.IsServerError() != tt.isServerError {
				t.Errorf("IsServerError: expected %v, got %v", tt.isServerError, resp.IsServerError())
			}
			if resp.IsError() != tt.isError {
				t.Errorf("IsError: expected %v, got %v", tt.isError, resp.IsError())
			}
		})
	}
}

func TestClone(t *testing.T) {
	original := New(
		WithBaseURL("https://api.example.com"),
		WithHeader("X-Original", "value"),
		WithRetry(3),
	)

	clone := original.Clone()

	// Verify clone has same values
	if clone.baseURL != original.baseURL {
		t.Error("clone should have same base URL")
	}
	if clone.headers["X-Original"] != "value" {
		t.Error("clone should have same headers")
	}
	if clone.maxRetries != 3 {
		t.Error("clone should have same retry config")
	}

	// Modify clone and verify original is unchanged
	clone.headers["X-Clone"] = "clone-value"
	if original.headers["X-Clone"] == "clone-value" {
		t.Error("modifying clone should not affect original")
	}
}

func TestWith(t *testing.T) {
	original := New(WithBaseURL("https://api.example.com"))

	modified := original.With(
		WithHeader("X-New", "value"),
		WithRetry(5),
	)

	// Original should be unchanged
	if original.headers["X-New"] == "value" {
		t.Error("With should not modify original")
	}
	if original.maxRetries == 5 {
		t.Error("With should not modify original")
	}

	// Modified should have new values
	if modified.headers["X-New"] != "value" {
		t.Error("With should add new header to modified")
	}
	if modified.maxRetries != 5 {
		t.Error("With should set retry on modified")
	}
}

func TestSlackResponseCheck(t *testing.T) {
	tests := []struct {
		name        string
		response    string
		expectError bool
	}{
		{
			name:        "success",
			response:    `{"ok": true, "channel": "C123"}`,
			expectError: false,
		},
		{
			name:        "error",
			response:    `{"ok": false, "error": "channel_not_found"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client := New()
			resp, err := client.Get(context.Background(), server.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, err = resp.CheckSlack()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

