package jwt

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.jwt"

// luaSign creates a signed JWT
// Usage: local token, err = jwt.sign({sub = "1234", name = "John"}, secret, {alg = "HS256", exp = 3600})
func luaSign(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaVerify verifies and decodes a JWT
// Usage: local claims, err = jwt.verify(token, secret)
func luaVerify(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDecode decodes a JWT without verification (unsafe, for inspection)
// Usage: local header, payload, err = jwt.decode(token)
func luaDecode(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LNil)
	L.Push(lua.LNil)
	L.Push(lua.LString("not implemented"))
	return 3
}

// luaGetClaims extracts claims from a verified token
// Usage: local claims, err = jwt.get_claims(token, secret)
func luaGetClaims(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaIsExpired checks if a token is expired
// Usage: local expired, err = jwt.is_expired(token)
func luaIsExpired(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(true))
	L.Push(lua.LString("not implemented"))
	return 2 // This returns 2 values (bool, error) not (result, error) pattern
}

// luaSignRS256 creates a JWT signed with RSA private key
// Usage: local token, err = jwt.sign_rs256(claims, private_key_pem)
func luaSignRS256(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaVerifyRS256 verifies a JWT with RSA public key
// Usage: local claims, err = jwt.verify_rs256(token, public_key_pem)
func luaVerifyRS256(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

var exports = map[string]lua.LGFunction{
	"sign":         luaSign,
	"verify":       luaVerify,
	"decode":       luaDecode,
	"get_claims":   luaGetClaims,
	"is_expired":   luaIsExpired,
	"sign_rs256":   luaSignRS256,
	"verify_rs256": luaVerifyRS256,
}

// Loader is called when the module is required via require("jwt")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
