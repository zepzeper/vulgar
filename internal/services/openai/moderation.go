package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Moderation checks text for policy violations
func (c *Client) Moderation(ctx context.Context, input string) (openai.ModerationResponse, error) {
	req := openai.ModerationRequest{
		Input: input,
		Model: openai.ModerationTextLatest,
	}
	return c.client.Moderations(ctx, req)
}
