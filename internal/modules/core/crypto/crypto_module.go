package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "crypto"

// luaSha256 hashes data using SHA-256
// Usage: local hash = crypto.sha256("data")
func luaSha256(L *lua.LState) int {
	data := L.CheckString(1)

	hash := sha256.Sum256([]byte(data))

	L.Push(lua.LString(hex.EncodeToString(hash[:])))
	return 1
}

// luaSha512 hashes data using SHA-512
// Usage: local hash = crypto.sha512("data")
func luaSha512(L *lua.LState) int {
	data := L.CheckString(1)

	hash := sha512.Sum512([]byte(data))

	L.Push(lua.LString(hex.EncodeToString(hash[:])))
	return 1
}

// luaMd5 hashes data using MD5 (for checksums, not security)
// Usage: local hash = crypto.md5("data")
func luaMd5(L *lua.LState) int {
	data := L.CheckString(1)

	hash := md5.Sum([]byte(data))

	L.Push(lua.LString(hex.EncodeToString(hash[:])))
	return 1
}

// luaHmacSha256 creates an HMAC-SHA256 signature
// Usage: local sig = crypto.hmac_sha256("data", "secret_key")
func luaHmacSha256(L *lua.LState) int {
	data := L.CheckString(1)
	key := L.CheckString(2)

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(data))
	hash := h.Sum(nil)

	L.Push(lua.LString(hex.EncodeToString(hash)))
	return 1
}

// luaBase64Encode encodes data to base64
// Usage: local encoded = crypto.base64_encode("data")
func luaBase64Encode(L *lua.LState) int {
	data := L.CheckString(1)

	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	L.Push(lua.LString(encoded))
	return 1
}

// luaBase64Decode decodes base64 data
// Usage: local decoded, err = crypto.base64_decode(encoded)
func luaBase64Decode(L *lua.LState) int {
	encoded := L.CheckString(1)

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return util.PushError(L, "base64 decode error: %v", err)
	}

	return util.PushSuccess(L, lua.LString(string(decoded)))
}

// luaHexEncode encodes data to hexadecimal
// Usage: local encoded = crypto.hex_encode("data")
func luaHexEncode(L *lua.LState) int {
	data := L.CheckString(1)

	encoded := hex.EncodeToString([]byte(data))

	L.Push(lua.LString(encoded))
	return 1
}

// luaHexDecode decodes hexadecimal data
// Usage: local decoded, err = crypto.hex_decode(encoded)
func luaHexDecode(L *lua.LState) int {
	encoded := L.CheckString(1)

	decoded, err := hex.DecodeString(encoded)
	if err != nil {
		return util.PushError(L, "hex decode error: %v", err)
	}

	return util.PushSuccess(L, lua.LString(string(decoded)))
}

// luaRandomBytes generates cryptographically secure random bytes
// Usage: local bytes = crypto.random_bytes(32)
func luaRandomBytes(L *lua.LState) int {
	length := L.CheckInt(1)

	if length <= 0 {
		return util.PushError(L, "length must be positive")
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return util.PushError(L, "random generation error: %v", err)
	}

	// Return as hex string for easy handling in Lua
	return util.PushSuccess(L, lua.LString(hex.EncodeToString(bytes)))
}

var exports = map[string]lua.LGFunction{
	"sha256":        luaSha256,
	"sha512":        luaSha512,
	"md5":           luaMd5,
	"hmac_sha256":   luaHmacSha256,
	"base64_encode": luaBase64Encode,
	"base64_decode": luaBase64Decode,
	"hex_encode":    luaHexEncode,
	"hex_decode":    luaHexDecode,
	"random_bytes":  luaRandomBytes,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
