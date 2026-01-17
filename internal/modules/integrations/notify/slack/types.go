package slack

import (
	"sync"

	"github.com/zepzeper/vulgar/internal/httpclient"
)

const ModuleName = "integrations.slack"
const luaSlackClientTypeName = "slack_client"
const slackAPIBase = "https://slack.com/api"

// slackClient wraps httpclient.Client with Slack-specific functionality
type slackClient struct {
	client *httpclient.Client
	token  string
	mu     sync.Mutex
	closed bool
}

type webhookPayload struct {
	Text        string        `json:"text,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	IconURL     string        `json:"icon_url,omitempty"`
	Attachments []interface{} `json:"attachments,omitempty"`
	Blocks      []interface{} `json:"blocks,omitempty"`
}

type chatPostMessageRequest struct {
	Channel     string        `json:"channel"`
	Text        string        `json:"text,omitempty"`
	Blocks      []interface{} `json:"blocks,omitempty"`
	Attachments []interface{} `json:"attachments,omitempty"`
	ThreadTS    string        `json:"thread_ts,omitempty"`
	Mrkdwn      bool          `json:"mrkdwn,omitempty"`
}
