package dns

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.dns"

// luaLookup performs a DNS lookup for a hostname
// Usage: local ips, err = dns.lookup("example.com")
func luaLookup(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaReverse performs a reverse DNS lookup
// Usage: local names, err = dns.reverse("93.184.216.34")
func luaReverse(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaMX looks up MX records
// Usage: local records, err = dns.mx("example.com")
func luaMX(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaTXT looks up TXT records
// Usage: local records, err = dns.txt("example.com")
func luaTXT(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCNAME looks up CNAME records
// Usage: local cname, err = dns.cname("www.example.com")
func luaCNAME(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaNS looks up NS records
// Usage: local records, err = dns.ns("example.com")
func luaNS(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSRV looks up SRV records
// Usage: local records, err = dns.srv("_sip._tcp.example.com")
func luaSRV(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"lookup":  luaLookup,
	"reverse": luaReverse,
	"mx":      luaMX,
	"txt":     luaTXT,
	"cname":   luaCNAME,
	"ns":      luaNS,
	"srv":     luaSRV,
}

// Loader is called when the module is required via require("dns")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
