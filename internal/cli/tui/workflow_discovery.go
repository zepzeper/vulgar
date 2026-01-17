package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zepzeper/vulgar/internal/config"
)

type WorkflowInfo struct {
	Name     string
	Path     string
	FullPath string
}

func DiscoverWorkflows() ([]WorkflowInfo, error) {
	workflowsPath := config.GetWorkflowsPath()

	// Resolve path (handle both relative and absolute paths)
	absPath, err := resolveWorkflowsPath(workflowsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workflows path: %w", err)
	}

	if err := validateWorkflowsDirectory(absPath); err != nil {
		if os.IsNotExist(err) {
			// Return empty list if directory doesn't exist (not an error)
			return []WorkflowInfo{}, nil
		}
		return nil, err
	}

	workflows, err := scanForWorkflows(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan workflows directory: %w", err)
	}

	// Sort workflows by path for consistent display
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].Path < workflows[j].Path
	})

	return workflows, nil
}

// resolveWorkflowsPath converts a workflows path (relative or absolute) to an absolute path
func resolveWorkflowsPath(workflowsPath string) (string, error) {
	if filepath.IsAbs(workflowsPath) {
		return workflowsPath, nil
	}

	// Relative to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return filepath.Join(cwd, workflowsPath), nil
}

// validateWorkflowsDirectory checks if the path exists and is a directory
func validateWorkflowsDirectory(absPath string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("workflows path is not a directory: %s", absPath)
	}

	return nil
}

// scanForWorkflows recursively walks the directory tree and collects all .lua files
func scanForWorkflows(rootPath string) ([]WorkflowInfo, error) {
	var workflows []WorkflowInfo

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only process .lua files (skip directories)
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".lua") {
			return nil
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			relPath = filepath.Base(path)
		}

		workflows = append(workflows, WorkflowInfo{
			Name:     filepath.Base(path),
			Path:     relPath,
			FullPath: path,
		})

		return nil
	})

	return workflows, err
}
