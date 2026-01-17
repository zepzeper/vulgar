package gsheets

import (
	"context"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"google.golang.org/api/sheets/v4"
)

const ModuleName = "integrations.gsheets"

// clientWrapper wraps the Sheets service for Lua userdata
type clientWrapper struct {
	service *sheets.Service
	ctx     context.Context
}

var (
	clientType  = "gsheets.client"
	clientMutex sync.Mutex
	clients     = make(map[*lua.LUserData]*clientWrapper)
)
