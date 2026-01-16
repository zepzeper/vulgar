package cron

import (
	"sync"

	"github.com/robfig/cron/v3"
	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const (
	ModuleName     = "stdlib.cron"
	luaJobTypeName = "cron_job"
)

// scheduler holds the global cron scheduler and job registry
type scheduler struct {
	c       *cron.Cron
	jobs    map[cron.EntryID]*jobHandle
	mu      sync.Mutex
	started bool
}

// jobHandle wraps a cron job with its callback
type jobHandle struct {
	id       cron.EntryID
	callback *lua.LFunction
	L        *lua.LState
	expr     string
	mu       sync.Mutex
	stopped  bool
	queue    *util.EventQueue
}
