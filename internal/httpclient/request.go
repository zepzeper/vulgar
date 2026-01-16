package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// RequestBuilder provides a fluent interface for building HTTP requests
type RequestBuilder struct {
	client      *Client
	method      string
	url         string
	headers     map[string]string
	queryParams url.Values
	body        io.Reader
	ctx         context.Context
}

// NewRequest creates a new request builder
func (c *Client) NewRequest(method, urlPath string) *RequestBuilder {
	return &RequestBuilder{
		client:      c,
		method:      method,
		url:         urlPath,
		headers:     make(map[string]string),
		queryParams: make(url.Values),
		ctx:         context.Background(),
	}
}

// Context sets the request context
func (rb *RequestBuilder) Context(ctx context.Context) *RequestBuilder {
	rb.ctx = ctx
	return rb
}

// Header adds a header to the request
func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// Headers adds multiple headers to the request
func (rb *RequestBuilder) Headers(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

// Query adds a query parameter
func (rb *RequestBuilder) Query(key, value string) *RequestBuilder {
	rb.queryParams.Add(key, value)
	return rb
}

// QueryParams adds multiple query parameters
func (rb *RequestBuilder) QueryParams(params map[string]string) *RequestBuilder {
	for k, v := range params {
		rb.queryParams.Add(k, v)
	}
	return rb
}

// Body sets the request body
func (rb *RequestBuilder) Body(body io.Reader) *RequestBuilder {
	rb.body = body
	return rb
}

// BodyString sets the request body from a string
func (rb *RequestBuilder) BodyString(body string) *RequestBuilder {
	rb.body = strings.NewReader(body)
	return rb
}

// BodyBytes sets the request body from bytes
func (rb *RequestBuilder) BodyBytes(body []byte) *RequestBuilder {
	rb.body = bytes.NewReader(body)
	return rb
}

// BodyJSON sets the request body from a JSON-serializable object
func (rb *RequestBuilder) BodyJSON(body interface{}) *RequestBuilder {
	data, err := json.Marshal(body)
	if err != nil {
		// Store error to be returned on Do()
		rb.body = &errorReader{err: err}
		return rb
	}
	rb.body = bytes.NewReader(data)
	rb.headers["Content-Type"] = "application/json"
	return rb
}

// BodyForm sets the request body from form values
func (rb *RequestBuilder) BodyForm(values url.Values) *RequestBuilder {
	rb.body = strings.NewReader(values.Encode())
	rb.headers["Content-Type"] = "application/x-www-form-urlencoded"
	return rb
}

// Do executes the request
func (rb *RequestBuilder) Do() (*Response, error) {
	// Check for body encoding errors
	if er, ok := rb.body.(*errorReader); ok {
		return nil, er.err
	}

	// Build URL with query params
	reqURL := rb.client.buildURL(rb.url)
	if len(rb.queryParams) > 0 {
		if strings.Contains(reqURL, "?") {
			reqURL += "&" + rb.queryParams.Encode()
		} else {
			reqURL += "?" + rb.queryParams.Encode()
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(rb.ctx, rb.method, reqURL, rb.body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Apply headers
	for k, v := range rb.headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := rb.client.Do(req)
	if err != nil {
		return nil, err
	}

	return NewResponse(resp)
}

// JSON executes the request and unmarshals the response
func (rb *RequestBuilder) JSON(result interface{}) error {
	resp, err := rb.Do()
	if err != nil {
		return err
	}
	return resp.JSON(result)
}

// errorReader is a reader that always returns an error
type errorReader struct {
	err error
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

// =============================================================================
// Multipart Form Data
// =============================================================================

// MultipartBuilder helps build multipart form data requests
type MultipartBuilder struct {
	client  *Client
	method  string
	url     string
	headers map[string]string
	ctx     context.Context
	body    *bytes.Buffer
	writer  *multipart.Writer
	err     error
}

// NewMultipartRequest creates a new multipart request builder
func (c *Client) NewMultipartRequest(method, urlPath string) *MultipartBuilder {
	body := &bytes.Buffer{}
	return &MultipartBuilder{
		client:  c,
		method:  method,
		url:     urlPath,
		headers: make(map[string]string),
		ctx:     context.Background(),
		body:    body,
		writer:  multipart.NewWriter(body),
	}
}

// Context sets the request context
func (mb *MultipartBuilder) Context(ctx context.Context) *MultipartBuilder {
	mb.ctx = ctx
	return mb
}

// Header adds a header to the request
func (mb *MultipartBuilder) Header(key, value string) *MultipartBuilder {
	mb.headers[key] = value
	return mb
}

// Field adds a form field
func (mb *MultipartBuilder) Field(name, value string) *MultipartBuilder {
	if mb.err != nil {
		return mb
	}
	if err := mb.writer.WriteField(name, value); err != nil {
		mb.err = err
	}
	return mb
}

// File adds a file field from a file path
func (mb *MultipartBuilder) File(fieldName, filePath string) *MultipartBuilder {
	if mb.err != nil {
		return mb
	}

	file, err := os.Open(filePath)
	if err != nil {
		mb.err = fmt.Errorf("failed to open file: %w", err)
		return mb
	}
	defer file.Close()

	part, err := mb.writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		mb.err = fmt.Errorf("failed to create form file: %w", err)
		return mb
	}

	if _, err := io.Copy(part, file); err != nil {
		mb.err = fmt.Errorf("failed to copy file content: %w", err)
		return mb
	}

	return mb
}

// FileFromReader adds a file field from a reader
func (mb *MultipartBuilder) FileFromReader(fieldName, fileName string, reader io.Reader) *MultipartBuilder {
	if mb.err != nil {
		return mb
	}

	part, err := mb.writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		mb.err = fmt.Errorf("failed to create form file: %w", err)
		return mb
	}

	if _, err := io.Copy(part, reader); err != nil {
		mb.err = fmt.Errorf("failed to copy content: %w", err)
		return mb
	}

	return mb
}

// FileFromBytes adds a file field from bytes
func (mb *MultipartBuilder) FileFromBytes(fieldName, fileName string, content []byte) *MultipartBuilder {
	return mb.FileFromReader(fieldName, fileName, bytes.NewReader(content))
}

// Do executes the multipart request
func (mb *MultipartBuilder) Do() (*Response, error) {
	if mb.err != nil {
		return nil, mb.err
	}

	// Close the writer to finalize the multipart form
	if err := mb.writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	reqURL := mb.client.buildURL(mb.url)
	req, err := http.NewRequestWithContext(mb.ctx, mb.method, reqURL, mb.body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type with boundary
	req.Header.Set("Content-Type", mb.writer.FormDataContentType())

	// Apply additional headers
	for k, v := range mb.headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := mb.client.Do(req)
	if err != nil {
		return nil, err
	}

	return NewResponse(resp)
}

