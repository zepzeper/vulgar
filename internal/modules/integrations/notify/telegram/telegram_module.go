package telegram

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.telegram"

// luaSend sends a message to a Telegram chat
// Usage: local err = telegram.send(chat_id, message)
func luaSend(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSendPhoto sends a photo to a Telegram chat
// Usage: local err = telegram.send_photo(chat_id, photo_path, caption)
func luaSendPhoto(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSendDocument sends a document to a Telegram chat
// Usage: local err = telegram.send_document(chat_id, document_path, caption)
func luaSendDocument(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaConfigure configures the Telegram bot
// Usage: local bot, err = telegram.configure({token = "123456:ABC-..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"send":          luaSend,
	"send_photo":    luaSendPhoto,
	"send_document": luaSendDocument,
	"configure":     luaConfigure,
}

// Loader is called when the module is required via require("telegram")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
