package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// ListModels lists the currently available models
func (c *Client) ListModels(ctx context.Context) (openai.ModelsList, error) {
	return c.client.ListModels(ctx)
}
