package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// GenerateImage generates images based on the prompt
func (c *Client) GenerateImage(ctx context.Context, req openai.ImageRequest) (openai.ImageResponse, error) {
	if req.Model == "" {
		req.Model = openai.CreateImageModelDallE3
	}
	// DALL-E 3 requires 1 image, so default to that if not set
	if req.N == 0 {
		req.N = 1
	}
	if req.Size == "" {
		req.Size = openai.CreateImageSize1024x1024
	}
	return c.client.CreateImage(ctx, req)
}
