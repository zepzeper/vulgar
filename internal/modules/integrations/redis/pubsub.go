package redis

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaPublish publishes a message to a channel
func luaPublish(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	channel := L.CheckString(2)
	message := L.CheckString(3)

	count, err := w.client.Publish(context.Background(), channel, message)
	if err != nil {
		return util.PushError(L, "publish failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}
