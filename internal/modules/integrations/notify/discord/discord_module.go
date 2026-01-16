package discord

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.discord"

// luaSend sends a message to a Discord channel
// Usage: local err = discord.send(channel_id, message)
func luaSend(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSendWebhook sends a message via Discord webhook
// Usage: local err = discord.send_webhook(webhook_url, message)
func luaSendWebhook(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaSendEmbed sends an embed message
// Usage: local err = discord.send_embed(channel_id, {title = "Hello", description = "World", color = 0x00ff00})
func luaSendEmbed(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaConfigure configures the Discord client
// Usage: local client, err = discord.configure({token = "Bot ..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"send":         luaSend,
	"send_webhook": luaSendWebhook,
	"send_embed":   luaSendEmbed,
	"configure":    luaConfigure,
}

// Loader is called when the module is required via require("discord")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
