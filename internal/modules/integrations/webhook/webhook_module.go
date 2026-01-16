package webhook

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.webhook"

// luaSend sends a webhook request
// Usage: local response, err = webhook.send(url, payload, {method = "POST", headers = {}})
func luaSend(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSendJSON sends a JSON webhook request
// Usage: local response, err = webhook.send_json(url, data)
func luaSendJSON(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListen starts a webhook listener server
// Usage: local server, err = webhook.listen(port, function(req) return {status = 200, body = "ok"} end)
func luaListen(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaVerify verifies a webhook signature
// Usage: local valid, err = webhook.verify(payload, signature, secret, {algorithm = "sha256"})
func luaVerify(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(false))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaSign signs a webhook payload
// Usage: local signature, err = webhook.sign(payload, secret, {algorithm = "sha256"})
func luaSign(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaStop stops a webhook listener
// Usage: local err = webhook.stop(server)
func luaStop(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"send":      luaSend,
	"send_json": luaSendJSON,
	"listen":    luaListen,
	"verify":    luaVerify,
	"sign":      luaSign,
	"stop":      luaStop,
}

// Loader is called when the module is required via require("webhook")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
