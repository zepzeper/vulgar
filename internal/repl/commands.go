package repl

import (
	"fmt"
	"strings"

	"github.com/zepzeper/vulgar/internal/cli"
	"github.com/zepzeper/vulgar/internal/modules"
)

// CommandHandler handles built-in REPL commands
type CommandHandler struct {
	repl *REPL
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(r *REPL) *CommandHandler {
	return &CommandHandler{repl: r}
}

// Handle processes a REPL command and returns true to continue, false to exit
func (c *CommandHandler) Handle(line string) bool {
	cmd := strings.TrimPrefix(line, ":")
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return true
	}

	switch parts[0] {
	case "quit", "q", "exit":
		return false

	case "help", "h", "?":
		c.printHelp()

	case "clear", "c":
		// Clear screen (ANSI escape)
		fmt.Print("\033[2J\033[H")

	case "modules", "m":
		c.listModules()

	case "reset", "r":
		fmt.Println(cli.Warning("Reset not implemented. Restart REPL for clean state."))

	case "load", "l":
		if len(parts) < 2 {
			fmt.Println(errorStyle.Render("Usage: :load <filename>"))
			return true
		}
		c.loadFile(parts[1])

	default:
		fmt.Println(errorStyle.Render(fmt.Sprintf("Unknown command: %s (type :help for commands)", parts[0])))
	}

	return true
}

// printHelp displays REPL help
func (c *CommandHandler) printHelp() {
	fmt.Println(cli.Title("REPL Commands"))
	fmt.Println()
	fmt.Println("  " + cli.Primary(":help") + ", " + cli.Primary(":h") + ", " + cli.Primary(":?") + "     Show this help")
	fmt.Println("  " + cli.Primary(":quit") + ", " + cli.Primary(":q") + ", " + cli.Primary(":exit") + "  Exit the REPL")
	fmt.Println("  " + cli.Primary(":clear") + ", " + cli.Primary(":c") + "          Clear the screen")
	fmt.Println("  " + cli.Primary(":modules") + ", " + cli.Primary(":m") + "        List available modules")
	fmt.Println("  " + cli.Primary(":load <file>") + ", " + cli.Primary(":l") + "    Load and execute a Lua file")
	fmt.Println()
	fmt.Println(cli.Title("Tips"))
	fmt.Println()
	fmt.Println("  • End a line with \\ to continue on next line")
	fmt.Println("  • Press Enter on empty line to execute multiline code")
	fmt.Println("  • Expressions auto-print results (e.g., '1 + 1' prints '2')")
	fmt.Println("  • Use ↑/↓ arrows to navigate history")
	fmt.Println()
	fmt.Println(cli.Title("Example"))
	fmt.Println()
	fmt.Println(cli.Muted("  lua> ") + "local gsheets = require(\"integrations.gsheets\")")
	fmt.Println(cli.Muted("  lua> ") + "local client = gsheets.configure()")
	fmt.Println(cli.Muted("  lua> ") + "gsheets.get_values(client, \"sheet_id\", \"Sheet1!A1:D10\")")
}

// listModules displays available modules
func (c *CommandHandler) listModules() {
	registry := modules.GetRegistry()

	// Group modules by category
	categories := map[string][]string{
		"core":         {},
		"stdlib":       {},
		"integrations": {},
		"ai":           {},
	}

	for name := range registry {
		if strings.HasPrefix(name, "stdlib.") {
			categories["stdlib"] = append(categories["stdlib"], name)
		} else if strings.HasPrefix(name, "integrations.") {
			categories["integrations"] = append(categories["integrations"], name)
		} else if strings.HasPrefix(name, "ai.") {
			categories["ai"] = append(categories["ai"], name)
		} else {
			categories["core"] = append(categories["core"], name)
		}
	}

	fmt.Printf("%s (%d total)\n", cli.Title("Available Modules"), len(registry))
	fmt.Println(cli.Muted("Use require(\"module.name\") to load a module"))
	fmt.Println()

	categoryOrder := []string{"core", "stdlib", "integrations", "ai"}
	for _, cat := range categoryOrder {
		names := categories[cat]
		if len(names) == 0 {
			continue
		}
		fmt.Printf("  %s (%d):\n", cli.Primary(cat), len(names))
		for _, name := range names {
			fmt.Printf("    require(%q)\n", name)
		}
		fmt.Println()
	}
}

// loadFile loads and executes a Lua file
func (c *CommandHandler) loadFile(filename string) {
	if err := c.repl.Engine().RunWorkflow(filename); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Error loading %s: %v", filename, err)))
	} else {
		fmt.Println(successStyle.Render(fmt.Sprintf("Loaded: %s", filename)))
	}
}
