package log

import (
	"bytes"
	"os"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func setupLuaState() *lua.LState {
	L := lua.NewState()
	L.PreloadModule(ModuleName, Loader)
	return L
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestDebugLogsMessage(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.debug("debug message")
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "[DEBUG]") {
		t.Error("expected DEBUG level in output")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("expected message in output")
	}
}

func TestInfoLogsMessage(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.info("info message")
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "[INFO]") {
		t.Error("expected INFO level in output")
	}
	if !strings.Contains(output, "info message") {
		t.Error("expected message in output")
	}
}

func TestWarnLogsMessage(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.warn("warn message")
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "[WARN]") {
		t.Error("expected WARN level in output")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("expected message in output")
	}
}

func TestErrorLogsMessage(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.error("error message")
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "[ERROR]") {
		t.Error("expected ERROR level in output")
	}
	if !strings.Contains(output, "error message") {
		t.Error("expected message in output")
	}
}

func TestLogIncludesTimestamp(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.info("test")
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	// Timestamp format: [2006-01-02 15:04:05]
	if !strings.Contains(output, "[20") {
		t.Error("expected timestamp in output")
	}
}

func TestLoaderReturnsModule(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`log = require("log")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	mod := L.GetGlobal("log")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected table, got %s", mod.Type())
	}
}

func TestLoaderExportsAllFunctions(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	L.PreloadModule(ModuleName, Loader)
	err := L.DoString(`log = require("log")`)
	if err != nil {
		t.Fatalf("failed to require module: %v", err)
	}

	tbl := L.GetGlobal("log").(*lua.LTable)

	funcs := []string{"debug", "info", "warn", "error"}
	for _, name := range funcs {
		fn := L.GetField(tbl, name)
		if fn.Type() != lua.LTFunction {
			t.Errorf("expected %s to be a function", name)
		}
	}
}

func TestOpenRegistersGlobally(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	mod := L.GetGlobal("log")
	if mod.Type() != lua.LTTable {
		t.Errorf("expected global log table, got %s", mod.Type())
	}
}

func TestOpenMakesLogAvailableWithoutRequire(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	Open(L)

	output := captureOutput(func() {
		err := L.DoString(`log.info("direct access")`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "direct access") {
		t.Error("expected log to work without require")
	}
}

func TestFormatLogOutput(t *testing.T) {
	result := formatLog("INFO", "test message")

	if !strings.Contains(result, "[INFO]") {
		t.Error("expected level in formatted output")
	}
	if !strings.Contains(result, "test message") {
		t.Error("expected message in formatted output")
	}
	if !strings.Contains(result, "[20") {
		t.Error("expected timestamp in formatted output")
	}
}

func TestLogWithSpecialCharacters(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.info("message with 'quotes' and \"double quotes\"")
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "quotes") {
		t.Error("expected message with special characters")
	}
}

func TestLogWithNumbers(t *testing.T) {
	L := setupLuaState()
	defer L.Close()

	output := captureOutput(func() {
		err := L.DoString(`
			local log = require("log")
			log.info("count: " .. 42)
		`)
		if err != nil {
			t.Fatalf("failed to execute: %v", err)
		}
	})

	if !strings.Contains(output, "42") {
		t.Error("expected number in output")
	}
}
