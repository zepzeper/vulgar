package slack

import (
	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/services/slack"
)

var cachedClient *slack.Client

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slack",
		Short: "Discover Slack resources",
		Long: `Discover Slack channels, users, and other resources.

Requires Slack token to be configured. Set up with:
  vulgar init
  
Then configure token in ~/.config/vulgar/config.toml:
  [slack]
  token = "xoxb-your-bot-token"

Required OAuth Scopes (add at https://api.slack.com/apps):
  - channels:read     List public channels
  - groups:read       List private channels  
  - users:read        List users
  - chat:write        Send messages (for workflows)
  - reactions:write   Add reactions (for workflows)
  - files:write       Upload files (for workflows)`,
	}

	cmd.AddCommand(checkCmd())
	cmd.AddCommand(channelsCmd())
	cmd.AddCommand(usersCmd())
	cmd.AddCommand(findUserCmd())
	cmd.AddCommand(findChannelCmd())
	cmd.AddCommand(infoCmd())

	return cmd
}

func getClient() (*slack.Client, error) {
	if cachedClient != nil {
		return cachedClient, nil
	}

	client, err := slack.NewClientFromConfig()
	if err != nil {
		return nil, err
	}

	cachedClient = client
	return client, nil
}
