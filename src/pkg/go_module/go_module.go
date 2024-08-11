// pkg/go_module/go_module.go
package go_module

import (
	"path/filepath"
	"strings"
)

// GoModule is the interface for Go-related operations
type GoModule interface {
	IsGoFile(path string) bool
	ShouldIgnoreGoPath(relPath string) bool
	CanAddComment(path string) bool
}

type goModule struct {
	// Add any necessary fields here
}

// New creates a new instance of GoModule
func New() GoModule {
	return &goModule{}
}

// IsGoFile checks if the file is a Go-related file
func (g *goModule) IsGoFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".go" || filepath.Base(path) == "go.mod" || filepath.Base(path) == "go.sum"
}

// ShouldIgnoreGoPath checks if the path should be ignored for Go projects
func (g *goModule) ShouldIgnoreGoPath(relPath string) bool {
	// Add any Go-specific directories or files that should be ignored
	return strings.HasPrefix(relPath, "vendor/") || strings.HasPrefix(relPath, ".git/")
}

// CanAddComment checks if we can safely add a comment to the file
func (g *goModule) CanAddComment(path string) bool {
	filename := filepath.Base(path)
	return filepath.Ext(path) == ".go" && filename != "go.mod" && filename != "go.sum"
}