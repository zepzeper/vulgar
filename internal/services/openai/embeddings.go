package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// CreateEmbeddings creates embeddings for the given input
func (c *Client) CreateEmbeddings(ctx context.Context, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	if req.Model == "" {
		// Use a safe default for embeddings if not specified, though typically this should be explicit
		req.Model = openai.AdaEmbeddingV2
	}
	return c.client.CreateEmbeddings(ctx, req)
}
