package gzip

import (
	"bytes"
	"compress/gzip"
	"io"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.gzip"

// Usage: local compressed, err = gzip.compress(data)
func luaCompress(L *lua.LState) int {
	data := L.CheckString(1)

	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write([]byte(data))
	if err != nil {
		writer.Close()
		return util.PushError(L, "compression failed: %v", err)
	}

	if err := writer.Close(); err != nil {
		return util.PushError(L, "compression failed: %v", err)
	}

	L.Push(lua.LString(buf.String()))
	L.Push(lua.LNil)
	return 2
}

// Usage: local data, err = gzip.decompress(compressed)
func luaDecompress(L *lua.LState) int {
	data := L.CheckString(1)

	if len(data) == 0 {
		return util.PushError(L, "empty input")
	}

	reader, err := gzip.NewReader(bytes.NewReader([]byte(data)))
	if err != nil {
		return util.PushError(L, "invalid gzip data: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return util.PushError(L, "decompression failed: %v", err)
	}

	L.Push(lua.LString(string(decompressed)))
	L.Push(lua.LNil)
	return 2
}

var exports = map[string]lua.LGFunction{
	"compress":   luaCompress,
	"decompress": luaDecompress,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
