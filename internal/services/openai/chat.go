package openai

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

// Chat sends a chat completion request
func (c *Client) Chat(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.Model
	}
	// Sentinel value -1 indicates "use default" to distinguish from explicit 0
	if req.Temperature == -1 {
		req.Temperature = c.Temperature
	}
	return c.client.CreateChatCompletion(ctx, req)
}

// ChatStream sends a streaming chat completion request
func (c *Client) ChatStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error) {
	if req.Model == "" {
		req.Model = c.Model
	}
	// Sentinel value -1 indicates "use default" to distinguish from explicit 0
	if req.Temperature == -1 {
		req.Temperature = c.Temperature
	}
	return c.client.CreateChatCompletionStream(ctx, req)
}
