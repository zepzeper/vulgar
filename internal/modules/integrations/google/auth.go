package google

import (
	"context"
	"net/http"

	googleauth "github.com/zepzeper/vulgar/internal/services/google"
	"google.golang.org/api/option"
)

// NewHTTPClient creates an authenticated HTTP client for Google APIs using OAuth
func NewHTTPClient(ctx context.Context) (*http.Client, error) {
	return googleauth.GetClient(ctx)
}

// ClientOption returns a Google API client option for authentication
func ClientOption(ctx context.Context) (option.ClientOption, error) {
	client, err := googleauth.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	return option.WithHTTPClient(client), nil
}
