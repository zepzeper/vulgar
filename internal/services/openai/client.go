package openai

import (
	"fmt"
	"strconv"

	"github.com/sashabaranov/go-openai"
	"github.com/zepzeper/vulgar/internal/config"
)

type Client struct {
	client      *openai.Client
	Model       string
	Temperature float32
}

type ClientOptions struct {
	APIkey string
	Model  string
	Temp   string
}

func NewClient(opts ClientOptions) (*Client, error) {
	if opts.APIkey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	config := openai.DefaultConfig(opts.APIkey)

	client := openai.NewClientWithConfig(config)

	model := opts.Model
	if model == "" {
		model = openai.GPT4 // Set default
	}

	var temp float32 = 1.0
	if opts.Temp != "" {
		if t, err := strconv.ParseFloat(opts.Temp, 32); err == nil {
			temp = float32(t)
		}
	}

	return &Client{
		client:      client,
		Model:       model,
		Temperature: temp,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	apiKey, model, ok := config.GetOpenAIKey()

	if !ok {
		return nil, fmt.Errorf("openAI not configured")
	}

	return NewClient(ClientOptions{
		APIkey: apiKey,
		Model:  model,
	})
}

func (c *Client) SDK() *openai.Client {
	return c.client
}
