package retry

import (
	"math/rand/v2"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.retry"

func parseDuration(_ *lua.LState, val lua.LValue, defaultVal time.Duration) time.Duration {
	if val == lua.LNil {
		return defaultVal
	}

	switch v := val.(type) {
	case lua.LString:
		d, err := time.ParseDuration(string(v))
		if err != nil {
			return defaultVal
		}
		return d
	case lua.LNumber:
		// Treat numbers as milliseconds
		return time.Duration(v) * time.Millisecond
	default:
		return defaultVal
	}
}

type retryOptions struct {
	maxAttempts  int
	delay        time.Duration
	initialDelay time.Duration
	maxDelay     time.Duration
	jitter       float64
}

func parseOptions(L *lua.LState, tbl *lua.LTable) retryOptions {
	opts := retryOptions{
		maxAttempts:  3,                      // default
		delay:        100 * time.Millisecond, // default
		initialDelay: 100 * time.Millisecond,
		maxDelay:     30 * time.Second,
		jitter:       0,
	}

	if tbl == nil {
		return opts
	}

	if v := L.GetField(tbl, "max_attempts"); v != lua.LNil {
		if n, ok := v.(lua.LNumber); ok {
			opts.maxAttempts = int(n)
		}
	}

	opts.delay = parseDuration(L, L.GetField(tbl, "delay"), opts.delay)
	opts.initialDelay = parseDuration(L, L.GetField(tbl, "initial_delay"), opts.initialDelay)
	opts.maxDelay = parseDuration(L, L.GetField(tbl, "max_delay"), opts.maxDelay)

	if v := L.GetField(tbl, "jitter"); v != lua.LNil {
		if n, ok := v.(lua.LNumber); ok {
			opts.jitter = float64(n)
		}
	}

	return opts
}

// luaDo executes a function with retry logic
// Usage: local result, err = retry.do(function() return http.get(url) end, {max_attempts = 3, delay = "1s"})
func luaDo(L *lua.LState) int {
	fn := L.CheckFunction(1)
	opts := parseOptions(L, L.OptTable(2, nil))

	var lastErr error

	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		L.Push(fn)
		err := L.PCall(0, 1, nil) // 0 args, 1 attempt

		if err == nil {
			result := L.Get(-1)
			L.Pop(1)
			return util.PushSuccess(L, result)
		}

		lastErr = err

		if attempt < opts.maxAttempts {
			time.Sleep(opts.delay)
		}
	}

	return util.PushError(L, "retry failed: %v", lastErr)
}

// luaExponential executes with exponential backoff
// Usage: local result, err = retry.exponential(func, {max_attempts = 5, initial_delay = "100ms", max_delay = "30s"})
func luaExponential(L *lua.LState) int {
	fn := L.CheckFunction(1)
	opts := parseOptions(L, L.OptTable(2, nil))

	var lastErr error
	currentDelay := opts.initialDelay

	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		L.Push(fn)
		err := L.PCall(0, 1, nil)

		if err == nil {
			result := L.Get(-1)
			L.Pop(1)
			return util.PushSuccess(L, result)
		}

		lastErr = err

		if attempt < opts.maxAttempts {
			time.Sleep(currentDelay)
			// Double the delay, but cap at maxDelay
			currentDelay *= 2
			if currentDelay > opts.maxDelay {
				currentDelay = opts.maxDelay
			}
		}
	}

	return util.PushError(L, "exponential retry failed: %v", lastErr)
}

// luaLinear executes with linear backoff
// Usage: local result, err = retry.linear(func, {max_attempts = 5, delay = "1s"})
func luaLinear(L *lua.LState) int {
	fn := L.CheckFunction(1)
	opts := parseOptions(L, L.OptTable(2, nil))

	var lastErr error

	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		L.Push(fn)
		err := L.PCall(0, 1, nil)

		if err == nil {
			result := L.Get(-1)
			L.Pop(1)
			return util.PushSuccess(L, result)
		}

		lastErr = err

		if attempt < opts.maxAttempts {
			// Linear: delay * attempt number
			time.Sleep(opts.delay * time.Duration(attempt))
		}
	}

	return util.PushError(L, "linear retry failed: %v", lastErr)
}

// luaForever retries forever until success
// Usage: local result = retry.forever(func, {delay = "5s"})
func luaForever(L *lua.LState) int {
	fn := L.CheckFunction(1)
	opts := parseOptions(L, L.OptTable(2, nil))

	for {
		L.Push(fn)
		err := L.PCall(0, 1, nil)

		if err == nil {
			result := L.Get(-1)
			L.Pop(1)
			L.Push(result)
			return 1
		}

		time.Sleep(opts.delay)
	}
}

// luaWithJitter adds jitter to retry delays
// Usage: local result, err = retry.with_jitter(func, {max_attempts = 3, delay = "1s", jitter = 0.5})
func luaWithJitter(L *lua.LState) int {
	fn := L.CheckFunction(1)
	opts := parseOptions(L, L.OptTable(2, nil))

	var lastErr error

	for attempt := 1; attempt <= opts.maxAttempts; attempt++ {
		L.Push(fn)
		err := L.PCall(0, 1, nil)

		if err == nil {
			result := L.Get(-1)
			L.Pop(1)
			return util.PushSuccess(L, result)
		}

		lastErr = err

		if attempt < opts.maxAttempts {
			// Add jitter: delay ± (delay * jitter * random)
			jitterAmount := float64(opts.delay) * opts.jitter * (rand.Float64()*2 - 1)
			actualDelay := max(time.Duration(float64(opts.delay)+jitterAmount), 0)
			time.Sleep(actualDelay)
		}
	}

	return util.PushError(L, "jitter retry failed: %v", lastErr)
}

var exports = map[string]lua.LGFunction{
	"call":        luaDo,
	"exponential": luaExponential,
	"linear":      luaLinear,
	"forever":     luaForever,
	"with_jitter": luaWithJitter,
}

// Loader is called when the module is required via require("retry")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
