package slack

import (
	"context"
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/httpclient"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

func sendWebhook(webhookURL string, payload *webhookPayload) error {
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

func luaTableToBlocks(tbl *lua.LTable) []interface{} {
	var blocks []interface{}
	tbl.ForEach(func(_, v lua.LValue) {
		if block := util.LuaToGo(v); block != nil {
			blocks = append(blocks, block)
		}
	})
	return blocks
}
