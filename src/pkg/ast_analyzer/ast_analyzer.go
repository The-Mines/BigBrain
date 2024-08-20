// pkg/ast_analyzer/ast_analyzer.go

package ast_analyzer

import (
    "fmt"
    "path/filepath"
)

// Node represents a generic AST node
type Node interface {
    Type() string
    Children() []Node
}

// Function represents a function in the AST
type Function interface {
    Node
    Name() string
    Parameters() []string
}

// Class represents a class in the AST
type Class interface {
    Node
    Name() string
    Methods() []Function
}

// Variable represents a variable declaration in the AST
type Variable interface {
    Node
    Name() string
    TypeInfo() string
}

// ASTAnalyzer defines the interface for AST analysis operations
type ASTAnalyzer interface {
    ParseFile(filePath string) (Node, error)
    GetFunctions(node Node) []Function
    GetClasses(node Node) []Class
    GetVariables(node Node) []Variable
    TraverseAST(node Node, visitor func(Node) bool)
}

// astAnalyzer implements the ASTAnalyzer interface
type astAnalyzer struct{}

// New creates a new instance of ASTAnalyzer
func New() ASTAnalyzer {
    return &astAnalyzer{}
}

// ParseFile parses a file and returns its AST
func (a *astAnalyzer) ParseFile(filePath string) (Node, error) {
    switch getFileExtension(filePath) {
    case ".py":
        return ParsePythonFile(filePath)
    case ".go":
        return ParseGoFile(filePath)
    // Add more cases for other languages
    default:
        return nil, fmt.Errorf("unsupported file type")
    }
}

// GetFunctions returns all function declarations in the AST
func (a *astAnalyzer) GetFunctions(node Node) []Function {
    var functions []Function
    a.TraverseAST(node, func(n Node) bool {
        if f, ok := n.(Function); ok {
            functions = append(functions, f)
        }
        return true
    })
    return functions
}

// GetClasses returns all class declarations in the AST
func (a *astAnalyzer) GetClasses(node Node) []Class {
    var classes []Class
    a.TraverseAST(node, func(n Node) bool {
        if c, ok := n.(Class); ok {
            classes = append(classes, c)
        }
        return true
    })
    return classes
}

// GetVariables returns all variable declarations in the AST
func (a *astAnalyzer) GetVariables(node Node) []Variable {
    var variables []Variable
    a.TraverseAST(node, func(n Node) bool {
        if v, ok := n.(Variable); ok {
            variables = append(variables, v)
        }
        return true
    })
    return variables
}

// TraverseAST traverses the AST and applies the visitor function to each node
func (a *astAnalyzer) TraverseAST(node Node, visitor func(Node) bool) {
    if !visitor(node) {
        return
    }
    for _, child := range node.Children() {
        a.TraverseAST(child, visitor)
    }
}

// getFileExtension returns the file extension of a given file path
func getFileExtension(filePath string) string {
    return filepath.Ext(filePath)
}
