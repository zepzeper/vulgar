package process

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const (
	ModuleName         = "stdlib.process"
	luaProcessTypeName = "process_handle"
)

type processHandle struct {
	cmd     *exec.Cmd
	pid     int
	mu      sync.Mutex
	started bool
	done    chan struct{}
}

var processMethods = map[string]lua.LGFunction{
	"wait":   luaProcessWait,
	"kill":   luaProcessKill,
	"pid":    luaProcessPid,
	"signal": luaProcessSignal,
}

func registerProcessType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaProcessTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), processMethods))
}

func checkProcess(L *lua.LState) *processHandle {
	ud := L.CheckUserData(1)
	if h, ok := ud.Value.(*processHandle); ok {
		return h
	}
	L.ArgError(1, "process_handle expected")
	return nil
}

// Usage: local output, err = process.exec("ls", {"-la", "/tmp"})
func luaExec(L *lua.LState) int {
	name := L.CheckString(1)
	argsTable := L.OptTable(2, L.NewTable())

	var args []string
	argsTable.ForEach(func(_, v lua.LValue) {
		if str, ok := v.(lua.LString); ok {
			args = append(args, string(str))
		}
	})

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			L.Push(lua.LString(string(output)))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		return util.PushError(L, "exec failed: %v", err)
	}

	L.Push(lua.LString(string(output)))
	L.Push(lua.LNil)
	return 2
}

// Usage: local proc, err = process.spawn("long-running-cmd", {args...})
func luaSpawn(L *lua.LState) int {
	name := L.CheckString(1)
	argsTable := L.OptTable(2, L.NewTable())

	var args []string
	argsTable.ForEach(func(_, v lua.LValue) {
		if str, ok := v.(lua.LString); ok {
			args = append(args, string(str))
		}
	})

	cmd := exec.Command(name, args...)

	// Start the process
	if err := cmd.Start(); err != nil {
		return util.PushError(L, "spawn failed: %v", err)
	}

	h := &processHandle{
		cmd:     cmd,
		pid:     cmd.Process.Pid,
		started: true,
		done:    make(chan struct{}),
	}

	// Wait in background
	go func() {
		_ = cmd.Wait()
		close(h.done)
	}()

	ud := util.NewUserData(L, h, luaProcessTypeName)
	return util.PushSuccess(L, ud)
}

// Usage: local err = process.kill(pid)
func luaKill(L *lua.LState) int {
	pid := int(L.CheckNumber(1))

	proc, err := os.FindProcess(pid)
	if err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	if err := proc.Kill(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local procs, err = process.list()
func luaList(L *lua.LState) int {
	// Read /proc directory for process list (Linux)
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return util.PushError(L, "failed to read /proc: %v", err)
	}

	procs := L.NewTable()
	idx := 1

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if directory name is a PID (numeric)
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Try to read process info
		cmdline, err := os.ReadFile("/proc/" + entry.Name() + "/cmdline")
		if err != nil {
			continue
		}

		// Replace null bytes with spaces
		cmdStr := strings.ReplaceAll(string(cmdline), "\x00", " ")
		cmdStr = strings.TrimSpace(cmdStr)

		procInfo := L.NewTable()
		procInfo.RawSetString("pid", lua.LNumber(pid))
		procInfo.RawSetString("cmdline", lua.LString(cmdStr))

		procs.RawSetInt(idx, procInfo)
		idx++
	}

	return util.PushSuccess(L, procs)
}

// Usage: local pid = process.pid()
func luaPid(L *lua.LState) int {
	L.Push(lua.LNumber(os.Getpid()))
	return 1
}

// Usage: local value = process.env("PATH") or process.env("MY_VAR", "value")
func luaEnv(L *lua.LState) int {
	name := L.CheckString(1)

	// If second argument provided, set the env var
	if L.GetTop() >= 2 {
		value := L.CheckString(2)
		if err := os.Setenv(name, value); err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(value))
		return 1
	}

	// Otherwise get the env var
	value := os.Getenv(name)
	if value == "" {
		L.Push(lua.LNil)
		return 1
	}

	L.Push(lua.LString(value))
	return 1
}

// Usage: local code, err = proc:wait()
func luaProcessWait(L *lua.LState) int {
	h := checkProcess(L)

	h.mu.Lock()
	cmd := h.cmd
	h.mu.Unlock()

	if cmd == nil {
		return util.PushError(L, "process not started")
	}

	// Wait for done channel
	<-h.done

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	L.Push(lua.LNumber(exitCode))
	L.Push(lua.LNil)
	return 2
}

// Usage: local err = proc:kill()
func luaProcessKill(L *lua.LState) int {
	h := checkProcess(L)

	h.mu.Lock()
	cmd := h.cmd
	h.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return util.PushError(L, "process not started")
	}

	if err := cmd.Process.Kill(); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

// Usage: local pid = proc:pid()
func luaProcessPid(L *lua.LState) int {
	h := checkProcess(L)
	L.Push(lua.LNumber(h.pid))
	return 1
}

// Usage: local err = proc:signal(signum)
func luaProcessSignal(L *lua.LState) int {
	h := checkProcess(L)
	sig := int(L.CheckNumber(2))

	h.mu.Lock()
	cmd := h.cmd
	h.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return util.PushError(L, "process not started")
	}

	if err := cmd.Process.Signal(syscall.Signal(sig)); err != nil {
		L.Push(lua.LString(err.Error()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

var exports = map[string]lua.LGFunction{
	"exec":  luaExec,
	"spawn": luaSpawn,
	"kill":  luaKill,
	"list":  luaList,
	"pid":   luaPid,
	"env":   luaEnv,
}

func Loader(L *lua.LState) int {
	registerProcessType(L)
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
