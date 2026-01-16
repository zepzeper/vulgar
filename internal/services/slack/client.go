package slack

import (
	"context"
	"fmt"

	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/httpclient"
)

const (
	DefaultBaseURL = "https://slack.com/api"
)

type Client struct {
	http           *httpclient.Client
	token          string
	defaultChannel string
}

type ClientOptions struct {
	Token          string
	DefaultChannel string
}

func NewClient(opts ClientOptions) (*Client, error) {
	if opts.Token == "" {
		return nil, fmt.Errorf("slack token is required")
	}

	httpClient := httpclient.New(
		httpclient.WithBaseURL(DefaultBaseURL),
		httpclient.WithSlackAuth(opts.Token),
		httpclient.WithRetry(2),
		httpclient.WithRateLimit(1),
	)

	return &Client{
		http:           httpClient,
		token:          opts.Token,
		defaultChannel: opts.DefaultChannel,
	}, nil
}

func NewClientFromConfig() (*Client, error) {
	token, ok := config.GetSlackToken()
	if !ok {
		return nil, fmt.Errorf("slack token not configured: run 'vulgar init' and set token in %s", config.ConfigPath())
	}

	cfg := config.Get()

	return NewClient(ClientOptions{
		Token:          token,
		DefaultChannel: cfg.Slack.DefaultChannel,
	})
}

func (c *Client) DefaultChannel() string {
	return c.defaultChannel
}

func (c *Client) ListChannels(ctx context.Context, includePrivate, includeArchived bool) ([]Channel, error) {
	types := "public_channel"
	if includePrivate {
		types = "public_channel,private_channel"
	}

	excludeArchived := "true"
	if includeArchived {
		excludeArchived = "false"
	}

	endpoint := fmt.Sprintf("/conversations.list?types=%s&exclude_archived=%s&limit=200", types, excludeArchived)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list channels failed: %w", err)
	}

	_, err = resp.CheckSlack()
	if err != nil {
		return nil, err
	}

	var result struct {
		Channels []Channel `json:"channels"`
	}
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("parse channels failed: %w", err)
	}

	return result.Channels, nil
}

func (c *Client) ListUsers(ctx context.Context, includeBots bool) ([]User, error) {
	resp, err := c.http.Get(ctx, "/users.list?limit=200")
	if err != nil {
		return nil, fmt.Errorf("list users failed: %w", err)
	}

	_, err = resp.CheckSlack()
	if err != nil {
		return nil, err
	}

	var result struct {
		Members []User `json:"members"`
	}
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("parse users failed: %w", err)
	}

	// Filter
	users := make([]User, 0)
	for _, u := range result.Members {
		if u.Deleted {
			continue
		}
		if !includeBots && u.IsBot {
			continue
		}
		// Copy email from profile
		if u.Email == "" && u.Profile.Email != "" {
			u.Email = u.Profile.Email
		}
		users = append(users, u)
	}

	return users, nil
}

func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	endpoint := fmt.Sprintf("/users.info?user=%s", userID)

	resp, err := c.http.Get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("get user failed: %w", err)
	}

	_, err = resp.CheckSlack()
	if err != nil {
		return nil, err
	}

	var result struct {
		User User `json:"user"`
	}
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("parse user failed: %w", err)
	}

	// Copy email from profile
	if result.User.Email == "" && result.User.Profile.Email != "" {
		result.User.Email = result.User.Profile.Email
	}

	return &result.User, nil
}

func (c *Client) AuthTest(ctx context.Context) (map[string]interface{}, error) {
	resp, err := c.http.Get(ctx, "/auth.test")
	if err != nil {
		return nil, fmt.Errorf("auth test failed: %w", err)
	}

	_, err = resp.CheckSlack()
	if err != nil {
		return nil, err
	}

	return resp.JSONMap()
}

func (c *Client) SendMessage(ctx context.Context, req SendMessageRequest) (*Message, error) {
	resp, err := c.http.NewRequest("POST", "/chat.postMessage").
		Context(ctx).
		BodyJSON(req).
		Do()
	if err != nil {
		return nil, fmt.Errorf("send message failed: %w", err)
	}

	_, err = resp.CheckSlack()
	if err != nil {
		return nil, err
	}

	var result struct {
		Message Message `json:"message"`
		Channel string  `json:"channel"`
		TS      string  `json:"ts"`
	}
	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	result.Message.Channel = result.Channel
	result.Message.Timestamp = result.TS

	return &result.Message, nil
}

func SendWebhook(webhookURL string, payload WebhookPayload) error {
	client := httpclient.New(
		httpclient.WithHeader("Content-Type", "application/json"),
	)

	resp, err := client.NewRequest("POST", webhookURL).
		Context(context.Background()).
		BodyJSON(payload).
		Do()

	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}

	if err := resp.CheckStatus(); err != nil {
		return fmt.Errorf("webhook error: %w (body: %s)", err, resp.String())
	}

	return nil
}

func (c *Client) AddReaction(ctx context.Context, channel, timestamp, emoji string) error {
	req := map[string]string{
		"channel":   channel,
		"timestamp": timestamp,
		"name":      emoji,
	}

	resp, err := c.http.NewRequest("POST", "/reactions.add").
		Context(ctx).
		BodyJSON(req).
		Do()
	if err != nil {
		return fmt.Errorf("add reaction failed: %w", err)
	}

	_, err = resp.CheckSlack()
	return err
}
