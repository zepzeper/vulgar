package twilio

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.twilio"

// luaConfigure configures the Twilio client
// Usage: local client, err = twilio.configure({account_sid = "AC...", auth_token = "..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSendSMS sends an SMS message
// Usage: local message, err = twilio.send_sms(client, {to = "+1234567890", from = "+0987654321", body = "Hello!"})
func luaSendSMS(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSendWhatsApp sends a WhatsApp message
// Usage: local message, err = twilio.send_whatsapp(client, {to = "whatsapp:+1234567890", from = "whatsapp:+0987654321", body = "Hello!"})
func luaSendWhatsApp(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMakeCall makes a voice call
// Usage: local call, err = twilio.make_call(client, {to = "+1234567890", from = "+0987654321", url = "http://example.com/twiml"})
func luaMakeCall(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetMessage gets message details
// Usage: local message, err = twilio.get_message(client, message_sid)
func luaGetMessage(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListMessages lists messages
// Usage: local messages, err = twilio.list_messages(client, {to = "+1234567890", limit = 20})
func luaListMessages(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaGetCall gets call details
// Usage: local call, err = twilio.get_call(client, call_sid)
func luaGetCall(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaListCalls lists calls
// Usage: local calls, err = twilio.list_calls(client, {to = "+1234567890", limit = 20})
func luaListCalls(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaLookupPhone looks up phone number information
// Usage: local info, err = twilio.lookup_phone(client, "+1234567890", {type = {"carrier", "caller-name"}})
func luaLookupPhone(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaVerifyStart starts phone verification
// Usage: local verification, err = twilio.verify_start(client, service_sid, {to = "+1234567890", channel = "sms"})
func luaVerifyStart(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaVerifyCheck checks verification code
// Usage: local result, err = twilio.verify_check(client, service_sid, {to = "+1234567890", code = "123456"})
func luaVerifyCheck(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"configure":     luaConfigure,
	"send_sms":      luaSendSMS,
	"send_whatsapp": luaSendWhatsApp,
	"make_call":     luaMakeCall,
	"get_message":   luaGetMessage,
	"list_messages": luaListMessages,
	"get_call":      luaGetCall,
	"list_calls":    luaListCalls,
	"lookup_phone":  luaLookupPhone,
	"verify_start":  luaVerifyStart,
	"verify_check":  luaVerifyCheck,
}

// Loader is called when the module is required via require("twilio")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
