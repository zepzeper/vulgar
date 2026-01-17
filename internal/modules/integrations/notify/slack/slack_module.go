package slack

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

var exports = map[string]lua.LGFunction{
	"client":         luaClient,
	"send":           luaSend,
	"send_webhook":   luaSendWebhook,
	"send_blocks":    luaSendBlocks,
	"upload_file":    luaUploadFile,
	"list_channels":  luaListChannels,
	"get_user":       luaGetUser,
	"list_users":     luaListUsers,
	"get_channel":    luaGetChannel,
	"react":          luaReact,
	"update_message": luaUpdateMessage,
	"delete_message": luaDeleteMessage,
	"pin_message":    luaPinMessage,
	"unpin_message":  luaUnpinMessage,
}

func Loader(L *lua.LState) int {
	registerSlackClientType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
