package main

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zepzeper/vulgar/cmd/vulgar/discover"
	"github.com/zepzeper/vulgar/cmd/vulgar/ui"
	"github.com/zepzeper/vulgar/internal/engine"
	"github.com/zepzeper/vulgar/internal/modules"
)

// Version information - set via ldflags at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// Flag variables organized by category
var (
	// Logging flags
	flagLogLevel  string
	flagLogFormat string
	flagVerbose   bool

	// Execution flags
	flagEval    string
	flagTimeout string
	flagDryRun  bool

	// Inspection flags
	flagCheck       bool
	flagListModules bool

	// Debug/profiling flags
	flagProfile bool
	flagTrace   bool
)

var rootCmd = &cobra.Command{
	Use:     "vulgar [script] [args...]",
	Short:   "The Go+Lua Workflow Automation Engine",
	Long:    `A powerful, modular engine for running automated workflows defined in Lua scripts.`,
	Version: Version,
	Args:    cobra.ArbitraryArgs,
	Run:     runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&flagLogLevel, "log-level", "l", "INFO", "Log level (DEBUG, INFO, WARN, ERROR)")
	rootCmd.Flags().StringVar(&flagLogFormat, "log-format", "text", "Log format (text, json)")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Enable verbose logging (sets log-level to DEBUG)")

	rootCmd.Flags().StringVarP(&flagEval, "eval", "e", "", "Execute Lua code directly instead of a file")
	rootCmd.Flags().StringVarP(&flagTimeout, "timeout", "t", "", "Execution timeout (e.g., 30s, 5m, 1h)")
	rootCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Parse and validate without executing side effects")

	rootCmd.Flags().BoolVarP(&flagCheck, "check", "c", false, "Check syntax only, do not execute")
	rootCmd.Flags().BoolVar(&flagListModules, "list-modules", false, "List all available modules and exit")

	rootCmd.Flags().BoolVar(&flagProfile, "profile", false, "Enable CPU profiling (writes to vulgar.prof)")
	rootCmd.Flags().BoolVar(&flagTrace, "trace", false, "Enable execution tracing (writes to vulgar.trace)")

	rootCmd.SetVersionTemplate(fmt.Sprintf("vulgar %s (built %s, commit %s)\n", Version, BuildTime, GitCommit))

	// Register discovery commands (init, gdrive, gsheets, etc.)
	discover.RegisterCommands(rootCmd)
	ui.RegisterCommands(rootCmd)
}

func runRoot(cmd *cobra.Command, args []string) {
	if flagListModules {
		listModules()
		return
	}

	if len(args) == 0 && flagEval == "" {
		cmd.Help()
		return
	}

	// Start CPU profiling if requested
	if flagProfile {
		f, err := os.Create("vulgar.prof")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create profile file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			pprof.StopCPUProfile()
			fmt.Fprintf(os.Stderr, "CPU profile written to vulgar.prof\n")
			fmt.Fprintf(os.Stderr, "Analyze with: go tool pprof vulgar.prof\n")
		}()
	}

	// Start execution tracing if requested
	if flagTrace {
		f, err := os.Create("vulgar.trace")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not create trace file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := trace.Start(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not start trace: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			trace.Stop()
			fmt.Fprintf(os.Stderr, "Execution trace written to vulgar.trace\n")
			fmt.Fprintf(os.Stderr, "Analyze with: go tool trace vulgar.trace\n")
		}()
	}

	logLevel := flagLogLevel
	if flagVerbose {
		logLevel = "DEBUG"
	}

	cfg := engine.Config{
		LogLevel:  logLevel,
		LogFormat: flagLogFormat,
		DryRun:    flagDryRun,
		Profile:   flagProfile,
		Trace:     flagTrace,
	}

	eng := engine.NewEngine(cfg)
	defer eng.Close()

	// Create context with optional timeout
	ctx := context.Background()
	if flagTimeout != "" {
		duration, err := time.ParseDuration(flagTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid timeout format %q: %v\n", flagTimeout, err)
			os.Exit(1)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, duration)
		defer cancel()
	}
	eng.SetContext(ctx)

	// Execute based on mode
	if flagEval != "" {
		runEval(eng, flagEval)
	} else {
		runScript(eng, args[0], args[1:])
	}
}

func runEval(eng *engine.Engine, code string) {
	if err := eng.Eval(code); err != nil {
		fmt.Fprintf(os.Stderr, "Error: eval failed: %v\n", err)
		os.Exit(1)
	}
}

func runScript(eng *engine.Engine, scriptPath string, scriptArgs []string) {
	// Syntax check mode
	if flagCheck {
		if err := eng.Compile(scriptPath); err != nil {
			fmt.Fprintf(os.Stderr, "Syntax error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Syntax OK: %s\n", scriptPath)
		return
	}

	// Dry run mode
	if flagDryRun {
		fmt.Printf("Dry run: %s\n", scriptPath)
		if err := eng.Compile(scriptPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Validation passed (no side effects executed)")
		return
	}

	// Normal execution
	if flagVerbose {
		fmt.Printf("Running workflow: %s\n", scriptPath)
	}

	if err := eng.RunWorkflow(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: workflow failed: %v\n", err)
		os.Exit(1)
	}

	if flagVerbose {
		fmt.Println("Workflow completed successfully")
	}
}

func listModules() {
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

	// Sort each category
	for cat := range categories {
		names := categories[cat]
		for i := 0; i < len(names)-1; i++ {
			for j := i + 1; j < len(names); j++ {
				if names[i] > names[j] {
					names[i], names[j] = names[j], names[i]
				}
			}
		}
	}

	total := len(registry)
	fmt.Printf("Available modules (%d):\n", total)

	// Print in order: core, stdlib, integrations, ai
	categoryOrder := []string{"core", "stdlib", "integrations", "ai"}
	for _, cat := range categoryOrder {
		names := categories[cat]
		if len(names) == 0 {
			continue
		}
		fmt.Printf("\n  %s (%d):\n", cat, len(names))
		for _, name := range names {
			fmt.Printf("    require(\"%s\")\n", name)
		}
	}
}
