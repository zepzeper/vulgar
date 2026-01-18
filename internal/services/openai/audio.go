package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// CreateTranscription transcribes audio
func (c *Client) CreateTranscription(ctx context.Context, req openai.AudioRequest) (openai.AudioResponse, error) {
	if req.Model == "" {
		req.Model = openai.Whisper1
	}
	return c.client.CreateTranscription(ctx, req)
}

// CreateTranslation translates audio
func (c *Client) CreateTranslation(ctx context.Context, req openai.AudioRequest) (openai.AudioResponse, error) {
	if req.Model == "" {
		req.Model = openai.Whisper1
	}
	return c.client.CreateTranslation(ctx, req)
}
