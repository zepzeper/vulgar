package repl

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/engine"
)

// Config holds REPL configuration options
type Config struct {
	// Preload common modules (gsheets, gdrive, http, json, log) into global scope
	Preload bool
	// HistoryFile is the path to the command history file
	HistoryFile string
	// MaxHistory is the maximum number of history entries to keep
	MaxHistory int
}

// DefaultConfig returns the default REPL configuration
func DefaultConfig() Config {
	homeDir, _ := os.UserHomeDir()
	return Config{
		Preload:     false,
		HistoryFile: filepath.Join(homeDir, ".vulgar_history"),
		MaxHistory:  1000,
	}
}

// REPL is an interactive Read-Eval-Print Loop for Lua code
type REPL struct {
	engine   *engine.Engine
	config   Config
	readline *readline.Instance
	commands *CommandHandler
}

// New creates a new REPL instance
func New(eng *engine.Engine, cfg Config) (*REPL, error) {
	r := &REPL{
		engine: eng,
		config: cfg,
	}
	r.commands = NewCommandHandler(r)

	// Configure readline
	rlConfig := &readline.Config{
		Prompt:            promptStyle.Render("lua> "),
		HistoryFile:       cfg.HistoryFile,
		HistoryLimit:      cfg.MaxHistory,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	}

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize readline: %w", err)
	}
	r.readline = rl

	return r, nil
}

// Run starts the REPL loop
func (r *REPL) Run() error {
	defer r.readline.Close()

	// Print welcome message
	fmt.Println(cli.Title("Vulgar Lua REPL"))
	fmt.Println(cli.Muted("Type Lua code to execute. Use :help for commands, :quit to exit."))

	// Preload modules if requested
	if r.config.Preload {
		if err := r.preloadModules(); err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("Warning: %v", err)))
		}
	}
	fmt.Println()

	// Main REPL loop
	var multilineBuffer strings.Builder
	inMultiline := false

	for {
		// Set prompt based on mode
		if inMultiline {
			r.readline.SetPrompt(continueStyle.Render("... "))
		} else {
			r.readline.SetPrompt(promptStyle.Render("lua> "))
		}

		// Read input
		line, err := r.readline.Readline()
		if err == readline.ErrInterrupt {
			if inMultiline {
				// Cancel multiline input
				multilineBuffer.Reset()
				inMultiline = false
				fmt.Println(cli.Muted("(input cancelled)"))
				continue
			}
			continue
		} else if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("readline error: %w", err)
		}

		line = strings.TrimRight(line, "\r\n")

		// Handle empty input
		if line == "" {
			if inMultiline {
				// Execute buffered code
				code := multilineBuffer.String()
				multilineBuffer.Reset()
				inMultiline = false
				r.execute(code)
			}
			continue
		}

		// Handle REPL commands (only when not in multiline mode)
		if !inMultiline && strings.HasPrefix(line, ":") {
			if !r.commands.Handle(line) {
				break // :quit was called
			}
			continue
		}

		// Check for explicit multiline continuation (ends with \)
		if strings.HasSuffix(line, "\\") {
			inMultiline = true
			multilineBuffer.WriteString(strings.TrimSuffix(line, "\\"))
			multilineBuffer.WriteString("\n")
			continue
		}

		// If in multiline mode, accumulate
		if inMultiline {
			multilineBuffer.WriteString(line)
			multilineBuffer.WriteString("\n")

			// Check if we should auto-execute (balanced blocks)
			if isCodeComplete(multilineBuffer.String()) {
				code := multilineBuffer.String()
				multilineBuffer.Reset()
				inMultiline = false
				r.execute(code)
			}
			continue
		}

		// Check if this line starts a block that needs more input
		if needsMoreInput(line) {
			inMultiline = true
			multilineBuffer.WriteString(line)
			multilineBuffer.WriteString("\n")
			continue
		}

		// Execute single line
		r.execute(line)
	}

	fmt.Println(cli.Muted("Goodbye!"))
	return nil
}

func (r *REPL) execute(code string) {
	code = strings.TrimSpace(code)
	if code == "" {
		return
	}

	results, err := Evaluate(r.engine, code)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
		return
	}

	// Print results if any
	if len(results) > 0 {
		PrintValues(results)
	}
}

func (r *REPL) preloadModules() error {
	preloadScript := `
gsheets = require("integrations.gsheets")
gdrive = require("integrations.gdrive")
http = require("http")
json = require("json")
log = require("log")
`
	if err := r.engine.Eval(preloadScript); err != nil {
		return fmt.Errorf("failed to preload modules: %w", err)
	}

	fmt.Println()
	fmt.Println(cli.Success("Preloaded modules:") + cli.Muted(" gsheets, gdrive, http, json, log"))
	fmt.Println(cli.Muted("Example: ") + cli.Code("client = gsheets.configure()"))
	return nil
}

func (r *REPL) Engine() *engine.Engine {
	return r.engine
}

func isCodeComplete(code string) bool {
	openers := 0
	closers := 0

	blockOpeners := []string{"function", "if", "for", "while", "repeat", "do"}
	for _, opener := range blockOpeners {
		openers += strings.Count(code, opener+" ") + strings.Count(code, opener+"(")
	}

	closers += strings.Count(code, "end")
	closers += strings.Count(code, "until")

	return closers >= openers
}

func needsMoreInput(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Check for block starters that need 'end'
	blockStarters := []string{"function ", "if ", "for ", "while ", "do", "repeat"}
	for _, starter := range blockStarters {
		if strings.HasPrefix(trimmed, starter) {
			// Check if it's a one-liner (contains 'end' or 'then ... end')
			if !strings.Contains(trimmed, "end") {
				return true
			}
		}
	}

	return false
}
