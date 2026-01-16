package log

import (
	"encoding/json"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
)

const ModuleName = "log"

const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

var currentLevel = LevelInfo
var currentFormat = FormatText

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

var levels = map[string]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
}

func SetLevel(level string) {
	currentLevel = level
}

func SetFormat(format string) {
	if format == FormatJSON {
		currentFormat = FormatJSON
	} else {
		currentFormat = FormatText
	}
}

func shouldLog(level string) bool {
	currentVal, ok1 := levels[currentLevel]
	msgVal, ok2 := levels[level]

	if !ok1 || !ok2 {
		return true
	}

	return msgVal >= currentVal
}

func formatLog(level, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if currentFormat == FormatJSON {
		entry := LogEntry{
			Timestamp: timestamp,
			Level:     level,
			Message:   message,
		}
		b, _ := json.Marshal(entry)
		return string(b)
	}

	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
}

// luaDebug logs a debug message
func luaDebug(L *lua.LState) int {
	if !shouldLog(LevelDebug) {
		return 0
	}
	message := L.CheckString(1)
	fmt.Println(formatLog(LevelDebug, message))
	return 0
}

// luaInfo logs an info message
func luaInfo(L *lua.LState) int {
	if !shouldLog(LevelInfo) {
		return 0
	}
	message := L.CheckString(1)
	fmt.Println(formatLog(LevelInfo, message))
	return 0
}

// luaWarn logs a warning message
func luaWarn(L *lua.LState) int {
	if !shouldLog(LevelWarn) {
		return 0
	}
	message := L.CheckString(1)
	fmt.Println(formatLog(LevelWarn, message))
	return 0
}

// luaError logs an error message
func luaError(L *lua.LState) int {
	if !shouldLog(LevelError) {
		return 0
	}
	message := L.CheckString(1)
	fmt.Println(formatLog(LevelError, message))
	return 0
}

// exports defines all functions exposed to Lua
var exports = map[string]lua.LGFunction{
	"debug": luaDebug,
	"info":  luaInfo,
	"warn":  luaWarn,
	"error": luaError,
}

// Loader is called when the module is required via require("log")
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)
	L.Push(mod)
	return 1
}

// Open registers the module globally
func Open(L *lua.LState) {
	L.SetGlobal(ModuleName, L.SetFuncs(L.NewTable(), exports))
}

// Auto-register with the module registry
func init() {
	modules.Register(ModuleName, Loader)
	modules.RegisterPreload(ModuleName, Open)
}
