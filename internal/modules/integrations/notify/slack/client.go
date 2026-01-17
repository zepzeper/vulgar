package slack

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/config"
	"github.com/zepzeper/vulgar/internal/httpclient"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

var clientMethods = map[string]lua.LGFunction{
	"send":           luaClientSend,
	"send_blocks":    luaClientSendBlocks,
	"upload_file":    luaClientUploadFile,
	"list_channels":  luaClientListChannels,
	"get_user":       luaClientGetUser,
	"list_users":     luaClientListUsers,
	"get_channel":    luaClientGetChannel,
	"react":          luaClientReact,
	"update_message": luaClientUpdateMessage,
	"delete_message": luaClientDeleteMessage,
	"pin_message":    luaClientPinMessage,
	"unpin_message":  luaClientUnpinMessage,
}

func registerSlackClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaSlackClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
	L.SetField(mt, "__gc", L.NewFunction(slackClientGC))
}

func checkSlackClient(L *lua.LState, idx int) *slackClient {
	ud := L.CheckUserData(idx)
	if v, ok := ud.Value.(*slackClient); ok {
		return v
	}
	L.ArgError(idx, "slack_client expected")
	return nil
}

func slackClientGC(L *lua.LState) int {
	ud := L.CheckUserData(1)
	if client, ok := ud.Value.(*slackClient); ok {
		client.close()
	}
	return 0
}

func (c *slackClient) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

func (c *slackClient) isClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

// apiRequest makes an authenticated request to Slack API using httpclient
func (c *slackClient) apiRequest(method, endpoint string, body interface{}) (map[string]interface{}, error) {
	if c.isClosed() {
		return nil, fmt.Errorf("client is closed")
	}

	ctx := context.Background()
	var resp *httpclient.Response
	var err error

	if body != nil {
		resp, err = c.client.NewRequest(method, endpoint).
			Context(ctx).
			BodyJSON(body).
			Do()
	} else {
		resp, err = c.client.Request(ctx, method, endpoint, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Use built-in Slack response checker
	_, err = resp.CheckSlack()
	if err != nil {
		return nil, err
	}

	return resp.JSONMap()
}

// Usage: local client, err = slack.client()  -- Uses token from ~/.config/vulgar/config.toml
// Or: local client, err = slack.client({token = "xoxb-..."})
func luaClient(L *lua.LState) int {
	opts := L.OptTable(1, nil)

	token := ""

	// Try explicit token first
	if opts != nil {
		if v := L.GetField(opts, "token"); v != lua.LNil {
			token = lua.LVAsString(v)
		}
	}

	// Fall back to config
	if token == "" {
		var ok bool
		token, ok = config.GetSlackToken()
		if !ok {
			return util.PushError(L, "Slack token not configured. Run 'vulgar init' and set token in config.toml")
		}
	}

	// Create httpclient with Slack configuration
	httpClient := httpclient.New(
		httpclient.WithBaseURL(slackAPIBase),
		httpclient.WithSlackAuth(token),
		httpclient.WithRetry(2),
		httpclient.WithRateLimit(1), // Slack rate limiting
	)

	client := &slackClient{
		client: httpClient,
		token:  token,
	}

	ud := L.NewUserData()
	ud.Value = client
	L.SetMetatable(ud, L.GetTypeMetatable(luaSlackClientTypeName))

	return util.PushSuccess(L, ud)
}
