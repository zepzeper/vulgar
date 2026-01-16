package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/internal/engine"
	"github.com/zepzeper/vulgar/internal/repl"
)

var (
	flagReplPreload     bool
	flagReplHistoryFile string
)

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start an interactive Lua REPL",
	Long: `Start an interactive Read-Eval-Print Loop for executing Lua code.

This allows you to interactively test Lua commands against the Vulgar engine,
including all available modules like gsheets, http, etc.

Features:
  • Command history with persistence (~/.vulgar_history)
  • Arrow key navigation (↑/↓ for history, ←/→ for editing)
  • Multiline input support
  • Expression auto-printing
  • Pretty-printed table output

Example usage:
  vulgar repl --preload

Then in the REPL (with --preload):
  > client = gsheets.configure()
  > gsheets.get_values(client, "spreadsheet_id", "Sheet1!A1:D10")

Without --preload, manually require modules:
  > local gsheets = require("integrations.gsheets")
  > local client = gsheets.configure()`,
	Run: runRepl,
}

func init() {
	replCmd.Flags().BoolVarP(&flagReplPreload, "preload", "p", false, "Preload common modules (gsheets, gdrive, http, json, log) into global scope")
	replCmd.Flags().StringVar(&flagReplHistoryFile, "history-file", "", "Path to history file (default: ~/.vulgar_history)")
	rootCmd.AddCommand(replCmd)
}

func runRepl(cmd *cobra.Command, args []string) {
	// Create engine
	cfg := engine.Config{
		LogLevel:  "INFO",
		LogFormat: "text",
	}
	eng := engine.NewEngine(cfg)
	defer eng.Close()

	// Create REPL config
	replCfg := repl.DefaultConfig()
	replCfg.Preload = flagReplPreload
	if flagReplHistoryFile != "" {
		replCfg.HistoryFile = flagReplHistoryFile
	}

	// Create and run REPL
	r, err := repl.New(eng, replCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to start REPL: %v\n", err)
		os.Exit(1)
	}

	if err := r.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
