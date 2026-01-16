package shell

import (
	"bytes"
	"os/exec"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

const ModuleName = "stdlib.shell"

// Usage: local output, err = shell.exec("ls -la /tmp")
func luaExec(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Return output even on error (for stderr)
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

// Usage: local code, output, err = shell.run("make build")
func luaRun(L *lua.LState) int {
	cmdStr := L.CheckString(1)

	cmd := exec.Command("sh", "-c", cmdStr)
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			L.Push(lua.LNumber(-1))
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 3
		}
	}

	L.Push(lua.LNumber(exitCode))
	L.Push(lua.LString(string(output)))
	L.Push(lua.LNil)
	return 3
}

// Usage: local output, err = shell.pipe("cat file.txt", "grep pattern", "wc -l")
func luaPipe(L *lua.LState) int {
	nArgs := L.GetTop()
	if nArgs < 1 {
		return util.PushError(L, "pipe requires at least one command")
	}

	commands := make([]string, nArgs)
	for i := 1; i <= nArgs; i++ {
		commands[i-1] = L.CheckString(i)
	}

	// Build pipeline
	var cmds []*exec.Cmd
	for _, cmdStr := range commands {
		cmd := exec.Command("sh", "-c", cmdStr)
		cmds = append(cmds, cmd)
	}

	// Connect pipes
	for i := 0; i < len(cmds)-1; i++ {
		pipe, err := cmds[i].StdoutPipe()
		if err != nil {
			return util.PushError(L, "failed to create pipe: %v", err)
		}
		cmds[i+1].Stdin = pipe
	}

	// Capture output of last command
	var outBuf bytes.Buffer
	cmds[len(cmds)-1].Stdout = &outBuf
	cmds[len(cmds)-1].Stderr = &outBuf

	// Start all commands
	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return util.PushError(L, "failed to start command: %v", err)
		}
	}

	// Wait for all commands
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			// Don't fail on exit errors for intermediate commands
			if _, ok := err.(*exec.ExitError); !ok {
				return util.PushError(L, "command failed: %v", err)
			}
		}
	}

	L.Push(lua.LString(outBuf.String()))
	L.Push(lua.LNil)
	return 2
}

// Usage: local quoted = shell.quote("file with spaces.txt")
func luaQuote(L *lua.LState) int {
	str := L.CheckString(1)

	// Use single quotes and escape any existing single quotes
	// This is the safest way to quote for POSIX shells
	if str == "" {
		L.Push(lua.LString("''"))
		return 1
	}

	// Replace ' with '\'' (end quote, escaped quote, start quote)
	escaped := strings.ReplaceAll(str, "'", "'\\''")
	quoted := "'" + escaped + "'"

	L.Push(lua.LString(quoted))
	return 1
}

// Usage: local path, err = shell.which("git")
func luaWhich(L *lua.LState) int {
	name := L.CheckString(1)

	path, err := exec.LookPath(name)
	if err != nil {
		return util.PushError(L, "command not found: %s", name)
	}

	L.Push(lua.LString(path))
	L.Push(lua.LNil)
	return 2
}

var exports = map[string]lua.LGFunction{
	"exec":  luaExec,
	"run":   luaRun,
	"pipe":  luaPipe,
	"quote": luaQuote,
	"which": luaWhich,
}

func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

func init() {
	modules.Register(ModuleName, Loader)
}
