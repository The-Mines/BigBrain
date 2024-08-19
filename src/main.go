// main.go
package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/The-Mines/BigBrain/pkg/file_processor"
	"github.com/The-Mines/BigBrain/pkg/node_module"
	"github.com/The-Mines/BigBrain/pkg/go_module"
	"github.com/The-Mines/BigBrain/pkg/python_module"
)

var (
	rootPath    string
	dryRun      bool
	verbose     bool
	runMode     bool
	ignoreRules []string
	nodeModule  node_module.NodeModule
	goModule    go_module.GoModule
	pythonModule python_module.PythonModule
	nodeOnly    bool
	goOnly      bool
	pythonOnly  bool
)

func loadGitignore(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ignoreRules = append(ignoreRules, line)
		}
	}

	return scanner.Err()
}

func shouldIgnore(path string) bool {
	relPath, err := filepath.Rel(rootPath, path)
	if err != nil {
		log.Printf("Error getting relative path for %s: %v\n", path, err)
		return false
	}

	// Convert Windows path separators to forward slashes
	relPath = filepath.ToSlash(relPath)

	// Don't ignore the root directory
	if relPath == "." {
		return false
	}

	// Use the Node module to check for Node.js specific paths
	if nodeModule != nil && nodeModule.ShouldIgnoreNodePath(relPath) {
		if verbose {
			log.Printf("Ignoring Node.js specific directory: %s\n", relPath)
		}
		return true
	}

	// Use the Python module to check for Python specific paths
	if pythonModule != nil && pythonModule.ShouldIgnorePythonPath(relPath) {
		if verbose {
			log.Printf("Ignoring Python specific directory: %s\n", relPath)
		}
		return true
	}

	// Ignore hidden files and directories (starting with a dot), except the root
	if strings.HasPrefix(filepath.Base(relPath), ".") && relPath != "." {
		if verbose {
			log.Printf("Ignoring hidden file/directory: %s\n", relPath)
		}
		return true
	}

	for _, rule := range ignoreRules {
		if rule[0] == '/' {
			// Rule starts with /, anchor to root
			if matched, _ := filepath.Match(rule[1:], relPath); matched {
				if verbose {
					log.Printf("Ignoring due to root-anchored rule %s: %s\n", rule, relPath)
				}
				return true
			}
		} else {
			// Rule applies to any depth
			if matched, _ := filepath.Match("**/"+rule, relPath); matched {
				if verbose {
					log.Printf("Ignoring due to rule %s: %s\n", rule, relPath)
				}
				return true
			}
		}
	}

	return false
}

var rootCmd = &cobra.Command{
	Use:   "BigBrain [path]",
	Short: "Recursively search for and update file paths",
	Long: `BigBrain is a CLI tool that recursively searches for files and updates their paths.
It can be configured to work specifically with Node.js, Go, or Python files.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			rootPath = args[0]
		} else {
			rootPath, _ = os.Getwd()
		}

		gitignorePath := filepath.Join(rootPath, ".gitignore")
		if err := loadGitignore(gitignorePath); err != nil {
			log.Printf("Warning: Could not load .gitignore: %v\n", err)
		}

		processor := fileprocessor.New(rootPath, verbose, nodeModule, goModule, pythonModule, nodeOnly, goOnly, pythonOnly)

		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if shouldIgnore(path) {
				if verbose {
					log.Printf("Ignoring: %s\n", path)
				}
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if info.IsDir() {
				return nil
			}

			if nodeOnly && !nodeModule.IsNodeFile(path) {
				if verbose {
					log.Printf("Skipping non-Node.js file: %s\n", path)
				}
				return nil
			}

			if goOnly && !goModule.IsGoFile(path) {
				if verbose {
					log.Printf("Skipping non-Go file: %s\n", path)
				}
				return nil
			}

			if pythonOnly && !pythonModule.IsPythonFile(path) {
				if verbose {
					log.Printf("Skipping non-Python file: %s\n", path)
				}
				return nil
			}

			if runMode {
				return processor.ProcessFileRun(path)
			}
			return processor.ProcessFile(path, dryRun)
		})

		if err != nil {
			log.Fatalf("Error during processing: %v\n", err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show which files would be modified without actually changing them")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().BoolVarP(&runMode, "run", "r", false, "Copy files with path comments to .bb folder")
	rootCmd.Flags().BoolVarP(&nodeOnly, "node", "n", false, "Process only Node.js files (.js, .ts, .jsx, .mjs, .cjs)")
	rootCmd.Flags().BoolVarP(&goOnly, "go", "g", false, "Process only Go files (.go, go.mod, go.sum)")
	rootCmd.Flags().BoolVarP(&pythonOnly, "python", "p", false, "Process only Python files (.py)")
}

func main() {
	// Initialize the Node module
	nodeModule = node_module.New()
	// Initialize the Go module
	goModule = go_module.New()
	// Initialize the Python module
	pythonModule = python_module.New()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v\n", err)
	}
}