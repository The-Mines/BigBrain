// pkg/node_module/node_module.go
package node_module

import (
	"path/filepath"
	"strings"
)

// NodeModule is the interface for Node.js related operations
type NodeModule interface {
	IsNodeFile(path string) bool
	ShouldIgnoreNodePath(relPath string) bool
}

type nodeModule struct {
	// Add any necessary fields here
}

// New creates a new instance of NodeModule
func New() NodeModule {
	return &nodeModule{}
}

func (n *nodeModule) IsNodeFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".ts" || ext == ".js" || ext == ".jsx" || ext == ".mjs" || ext == ".cjs"
}

func (n *nodeModule) ShouldIgnoreNodePath(relPath string) bool {
	return relPath == "public" || strings.HasPrefix(relPath, "public/") ||
		relPath == ".next" || strings.HasPrefix(relPath, ".next/")
}