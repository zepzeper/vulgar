package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/zepzeper/vulgar/internal/modules"
	_ "github.com/zepzeper/vulgar/internal/modules/all"
	log "github.com/zepzeper/vulgar/internal/modules/core/log"
	"github.com/zepzeper/vulgar/internal/modules/util"
)

type Engine struct {
	L          *lua.LState
	EventQueue *util.EventQueue
}

type Config struct {
	LogLevel  string
	LogFormat string
	DryRun    bool
	Profile   bool
	Trace     bool
}

func NewEngine(cfg Config) *Engine {
	L := lua.NewState()

	log.SetLevel(cfg.LogLevel)
	log.SetFormat(cfg.LogFormat)

	// Initialize EventQueue
	queue := util.NewEventQueue(L, 100)

	// Store in registry for modules to access
	ud := L.NewUserData()
	ud.Value = queue
	L.SetField(L.Get(lua.RegistryIndex), util.EventQueueRegistryKey, ud)

	e := &Engine{
		L:          L,
		EventQueue: queue,
	}

	e.setupModuleLoader()
	e.preloadCriticalModules()
	return e
}

func (e *Engine) setupModuleLoader() {
	preload := e.L.GetField(e.L.GetGlobal("package"), "preload")

	// Register each module from the auto-registry
	for name, loader := range modules.GetRegistry() {
		e.L.SetField(preload, name, e.L.NewFunction(loader))
	}
}

// preloadCriticalModules makes certain modules globally available without require()
// Currently only log is preloaded for error reporting before other modules load
func (e *Engine) preloadCriticalModules() {
	for name, opener := range modules.GetPreloadRegistry() {
		_ = name // name is available if we need to log which modules are preloaded
		opener(e.L)
	}
}

func (e *Engine) Eval(code string) error {
	if err := e.L.DoString(code); err != nil {
		return formatLuaError(err, "<eval>")
	}
	return nil
}

func (e *Engine) RunWorkflow(path string) error {
	// Execute the script
	if err := e.L.DoFile(path); err != nil {
		return formatLuaError(err, path)
	}

	// Check if optional RunWorkflow function exists and run it
	fn := e.L.GetGlobal("RunWorkflow")
	if fn != lua.LNil {
		if err := e.L.CallByParam(lua.P{
			Fn:      fn,
			NRet:    0,
			Protect: true,
		}); err != nil {
			return formatLuaError(err, path)
		}
	}

	// Main Event Loop
	// Continue running as long as there are active async sources (timers, watchers, etc.)
	// or pending events in the queue.
	for e.EventQueue.HasActiveSources() {
		if !e.EventQueue.WaitForEvents() {
			break
		}
	}

	return nil
}

// formatLuaError formats Lua errors with helpful suggestions
func formatLuaError(err error, scriptPath string) error {
	errStr := err.Error()

	// Check for module not found errors and provide helpful suggestions
	if strings.Contains(errStr, "module") && strings.Contains(errStr, "not found") {
		return formatModuleNotFoundError(errStr, scriptPath)
	}

	// Check for syntax errors
	if strings.Contains(errStr, "syntax error") {
		return formatSyntaxError(errStr, scriptPath)
	}

	// Default formatting
	return fmt.Errorf("%s: %w", filepath.Base(scriptPath), err)
}

// formatModuleNotFoundError provides suggestions for module naming
func formatModuleNotFoundError(errStr string, scriptPath string) error {
	// Extract the module name from error like "module xyz not found"
	var moduleName string
	if idx := strings.Index(errStr, "module "); idx != -1 {
		rest := errStr[idx+7:]
		if endIdx := strings.Index(rest, " "); endIdx != -1 {
			moduleName = rest[:endIdx]
		}
	}

	// Build suggestion based on module name
	var suggestion string
	if moduleName != "" {
		// Check if it's a known module that needs a prefix
		stdlibModules := []string{"timer", "cron", "event", "shell", "process", "filewatch",
			"gzip", "tar", "zip", "yaml", "xml", "csv", "regex", "strings", "cache",
			"parallel", "mathx", "jwt", "html", "url", "validator", "template",
			"queue", "workflow", "retry", "health", "metrics", "osinfo", "secrets", "trace"}

		integrationModules := []string{"redis", "postgres", "mongodb", "sqlite", "kafka",
			"rabbitmq", "nats", "s3", "smtp", "slack", "discord", "telegram", "github",
			"docker", "k8s", "webhook", "websocket", "graphql", "ftp", "dns", "stripe",
			"twilio", "notion", "airtable", "gsheets", "ssh"}

		aiModules := []string{"openai", "anthropic", "ollama", "huggingface", "localai"}

		for _, m := range stdlibModules {
			if moduleName == m {
				suggestion = fmt.Sprintf("\n  Hint: Use require(\"stdlib.%s\") instead of require(\"%s\")", m, m)
				break
			}
		}
		if suggestion == "" {
			for _, m := range integrationModules {
				if moduleName == m {
					suggestion = fmt.Sprintf("\n  Hint: Use require(\"integrations.%s\") instead of require(\"%s\")", m, m)
					break
				}
			}
		}
		if suggestion == "" {
			for _, m := range aiModules {
				if moduleName == m {
					suggestion = fmt.Sprintf("\n  Hint: Use require(\"ai.%s\") instead of require(\"%s\")", m, m)
					break
				}
			}
		}
		if suggestion == "" {
			suggestion = "\n  Run 'vulgar --list-modules' to see available modules"
		}
	}

	return fmt.Errorf("module not found: %s%s", moduleName, suggestion)
}

// formatSyntaxError formats syntax errors with line information
func formatSyntaxError(errStr string, scriptPath string) error {
	// Lua syntax errors usually include line numbers
	return fmt.Errorf("syntax error in %s:\n  %s", filepath.Base(scriptPath), errStr)
}

func (e *Engine) Close() {
	e.EventQueue.Close()
	e.L.Close()
}

func (e *Engine) Compile(path string) error {
	_, err := e.L.LoadFile(path)
	if err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}

	return nil
}

func (e *Engine) SetContext(ctx context.Context) {
	e.L.SetContext(ctx)
}
