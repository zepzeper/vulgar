package secrets

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.secrets"

// luaGet retrieves a secret
// Usage: local value, err = secrets.get("database_password")
func luaGet(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaSet stores a secret
// Usage: local err = secrets.set("api_key", "secret_value")
func luaSet(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDelete deletes a secret
// Usage: local err = secrets.delete("old_key")
func luaDelete(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaList lists all secret keys
// Usage: local keys, err = secrets.list()
func luaList(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaFromEnv loads a secret from environment variable
// Usage: local value, err = secrets.from_env("MY_SECRET")
func luaFromEnv(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaFromFile loads a secret from a file
// Usage: local value, err = secrets.from_file("/run/secrets/db_password")
func luaFromFile(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaFromVault loads a secret from HashiCorp Vault
// Usage: local value, err = secrets.from_vault("secret/data/myapp", "password", {address = "...", token = "..."})
func luaFromVault(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaRotate rotates a secret
// Usage: local err = secrets.rotate("api_key", function() return generate_new_key() end)
func luaRotate(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"get":        luaGet,
	"set":        luaSet,
	"delete":     luaDelete,
	"list":       luaList,
	"from_env":   luaFromEnv,
	"from_file":  luaFromFile,
	"from_vault": luaFromVault,
	"rotate":     luaRotate,
}

// Loader is called when the module is required
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
