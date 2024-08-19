// pkg/python_module/python_module.go
package python_module

import (
	"path/filepath"
	"strings"
)

// PythonModule is the interface for Python-related operations
type PythonModule interface {
	IsPythonFile(path string) bool
	ShouldIgnorePythonPath(relPath string) bool
	CanAddComment(path string) bool
	GetCommentPrefix() string
}

type pythonModule struct {
	// Add any necessary fields here
	ignoreDirs []string
}

// New creates a new instance of PythonModule
func New() PythonModule {
	return &pythonModule{
		ignoreDirs: []string{"venv", ".venv", "__pycache__"},
	}
}

func (p *pythonModule) IsPythonFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".py"
}

func (p *pythonModule) ShouldIgnorePythonPath(relPath string) bool {
	for _, dir := range p.ignoreDirs {
		if strings.HasPrefix(relPath, dir+string(filepath.Separator)) || relPath == dir {
			return true
		}
	}
	return false
}

func (p *pythonModule) CanAddComment(path string) bool {
	return p.IsPythonFile(path)
}

func (p *pythonModule) GetCommentPrefix() string {
	return "#"
}