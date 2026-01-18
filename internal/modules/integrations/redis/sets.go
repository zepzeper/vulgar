package redis

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

// luaSAdd adds members to a set
func luaSAdd(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	var members []interface{}
	for i := 3; i <= L.GetTop(); i++ {
		members = append(members, util.LuaToGo(L.Get(i)))
	}

	count, err := w.client.SAdd(context.Background(), key, members...)
	if err != nil {
		return util.PushError(L, "sadd failed: %v", err)
	}

	L.Push(lua.LNumber(count))
	L.Push(lua.LNil)
	return 2
}

// luaSMembers gets all members of a set
func luaSMembers(L *lua.LState) int {
	w := getWrapper(L, 1)
	if w == nil || w.client == nil {
		return util.PushError(L, "invalid redis client")
	}

	key := L.CheckString(2)

	members, err := w.client.SMembers(context.Background(), key)
	if err != nil {
		return util.PushError(L, "smembers failed: %v", err)
	}

	result := L.NewTable()
	for i, m := range members {
		result.RawSetInt(i+1, lua.LString(m))
	}

	L.Push(result)
	L.Push(lua.LNil)
	return 2
}
