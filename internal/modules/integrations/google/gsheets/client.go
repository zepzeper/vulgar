package gsheets

import (
	"context"

	lua "github.com/yuin/gopher-lua"
	gauth "github.com/zepzeper/vulgar/internal/modules/integrations/google"
	"github.com/zepzeper/vulgar/internal/modules/util"
	"google.golang.org/api/sheets/v4"
)

// registerClient stores a client wrapper and returns userdata
func registerClient(L *lua.LState, wrapper *clientWrapper) *lua.LUserData {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	ud := L.NewUserData()
	ud.Value = wrapper
	clients[ud] = wrapper

	// Set metatable for type checking
	mt := L.NewTable()
	L.SetField(mt, "__type", lua.LString(clientType))
	L.SetMetatable(ud, mt)

	return ud
}

// getClient retrieves the client wrapper from userdata
func getClient(L *lua.LState, idx int) *clientWrapper {
	ud := L.CheckUserData(idx)
	if ud == nil {
		return nil
	}

	wrapper, ok := ud.Value.(*clientWrapper)
	if !ok {
		return nil
	}

	return wrapper
}

// luaConfigure creates a new Google Sheets client using OAuth authentication
// Usage: local client, err = gsheets.configure()
// Note: Requires prior authentication via 'vulgar gsheets login'
func luaConfigure(L *lua.LState) int {
	ctx := context.Background()

	clientOpt, err := gauth.ClientOption(ctx)
	if err != nil {
		return util.PushError(L, "failed to get OAuth credentials: %v (run 'vulgar gsheets login' first)", err)
	}

	// Create Sheets service
	service, err := sheets.NewService(ctx, clientOpt)
	if err != nil {
		return util.PushError(L, "failed to create sheets service: %v", err)
	}

	// Wrap and return client
	wrapper := &clientWrapper{
		service: service,
		ctx:     ctx,
	}

	L.Push(registerClient(L, wrapper))
	L.Push(lua.LNil)
	return 2
}

