package s3

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "integrations.s3"

// luaConfigure configures the S3 client
// Usage: local client, err = s3.configure({access_key = "...", secret_key = "...", region = "us-east-1", endpoint = "..."})
func luaConfigure(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaUpload uploads a file to S3
// Usage: local err = s3.upload(client, bucket, key, content, {content_type = "text/plain"})
func luaUpload(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaDownload downloads a file from S3
// Usage: local content, err = s3.download(client, bucket, key)
func luaDownload(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaDelete deletes an object from S3
// Usage: local err = s3.delete(client, bucket, key)
func luaDelete(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

// luaList lists objects in a bucket
// Usage: local objects, err = s3.list(client, bucket, {prefix = "folder/", max_keys = 100})
func luaList(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaExists checks if object exists
// Usage: local exists, err = s3.exists(client, bucket, key)
func luaExists(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LBool(false))
	L.Push(lua.LString("not implemented"))
	return 2
}

// luaPresignedURL generates a presigned URL
// Usage: local url, err = s3.presigned_url(client, bucket, key, {expires = 3600, method = "GET"})
func luaPresignedURL(L *lua.LState) int {
	// TODO: implement
	return util.PushError(L, "not implemented")
}

// luaCopy copies an object
// Usage: local err = s3.copy(client, src_bucket, src_key, dst_bucket, dst_key)
func luaCopy(L *lua.LState) int {
	// TODO: implement
	L.Push(lua.LString("not implemented"))
	return 1
}

var exports = map[string]lua.LGFunction{
	"configure":     luaConfigure,
	"upload":        luaUpload,
	"download":      luaDownload,
	"delete":        luaDelete,
	"list":          luaList,
	"exists":        luaExists,
	"presigned_url": luaPresignedURL,
	"copy":          luaCopy,
}

// Loader is called when the module is required via require("s3")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Open registers the module globally (for backwards compatibility)
func Open(L *lua.LState) {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.SetGlobal(ModuleName, mod)
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
}
