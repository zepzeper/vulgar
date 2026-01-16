package gdrive

import (
	"context"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"google.golang.org/api/drive/v3"
)

const ModuleName = "integrations.gdrive"

type clientWrapper struct {
	service *drive.Service
	ctx     context.Context
}

var (
	clientMutex sync.Mutex
	clients     = make(map[*lua.LUserData]*clientWrapper)
)

// registerClient stores a client wrapper and returns userdata
func registerClient(L *lua.LState, wrapper *clientWrapper) *lua.LUserData {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	ud := L.NewUserData()
	ud.Value = wrapper
	clients[ud] = wrapper

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
