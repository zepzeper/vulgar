package workflow

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// EditorConfig holds the editor configuration
type EditorConfig struct {
	Command string   // The editor command (e.g., "code", "nvim", "vim")
	Args    []string // Additional arguments
}

// DefaultEditor returns the default editor based on EDITOR env var or common defaults
func DefaultEditor() EditorConfig {
	// Check EDITOR environment variable first
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try to find a common editor
		editors := []string{"code", "nvim", "vim", "nano", "vi"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		editor = "vi" // Ultimate fallback
	}

	return EditorConfig{Command: editor}
}

// OpenFile opens a file in the configured editor at the specified line
func (e EditorConfig) OpenFile(filePath string, line int) error {
	var args []string

	// Add line number argument based on editor type
	switch {
	case strings.Contains(e.Command, "code"):
		// VS Code: --goto file:line
		args = append(args, "--goto", fmt.Sprintf("%s:%d", filePath, line))
	case strings.Contains(e.Command, "nvim") || strings.Contains(e.Command, "vim") || strings.Contains(e.Command, "vi"):
		// Vim/Neovim: +line file
		args = append(args, fmt.Sprintf("+%d", line), filePath)
	case strings.Contains(e.Command, "nano"):
		// Nano: +line file
		args = append(args, fmt.Sprintf("+%d", line), filePath)
	case strings.Contains(e.Command, "emacs"):
		// Emacs: +line file
		args = append(args, fmt.Sprintf("+%d", line), filePath)
	case strings.Contains(e.Command, "subl") || strings.Contains(e.Command, "sublime"):
		// Sublime: file:line
		args = append(args, fmt.Sprintf("%s:%d", filePath, line))
	default:
		// Generic: just open the file
		args = append(args, filePath)
	}

	// Add any additional configured args
	args = append(e.Args, args...)

	cmd := exec.Command(e.Command, args...)

	// For terminal editors, we need to attach to terminal
	if isTerminalEditor(e.Command) {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// For GUI editors, just start and detach
	return cmd.Start()
}

// isTerminalEditor returns true if the editor runs in terminal
func isTerminalEditor(editor string) bool {
	terminalEditors := []string{"vim", "nvim", "vi", "nano", "emacs", "micro", "ne", "joe"}
	for _, te := range terminalEditors {
		if strings.Contains(editor, te) {
			return true
		}
	}
	return false
}

// IsTerminalEditor returns true if the configured editor runs in terminal
func (e EditorConfig) IsTerminalEditor() bool {
	return isTerminalEditor(e.Command)
}
