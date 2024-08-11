// main.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/The-Mines/BigBrain/pkg/node_module"
	"github.com/The-Mines/BigBrain/pkg/go_module"
)

var (
	rootPath    string
	dryRun      bool
	verbose     bool
	runMode     bool
	ignoreRules []string
	nodeModule  node_module.NodeModule
	goModule    go_module.GoModule
	nodeOnly    bool
	goOnly      bool
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
It can be configured to work specifically with Node.js or Go files.`,
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

		if runMode {
			bbDir := filepath.Join(rootPath, ".bb")
			err := os.MkdirAll(bbDir, os.ModePerm)
			if err != nil {
				log.Fatalf("Error creating .bb directory: %v\n", err)
			}
			log.Printf("Created .bb directory: %s\n", bbDir)
		}

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

			if runMode {
				return processFileRun(path)
			}
			return processFile(path, dryRun)
		})

		if err != nil {
			log.Fatalf("Error during processing: %v\n", err)
		}
	},
}


func processFile(path string, dryRun bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	hasFirstLine := scanner.Scan()
	var firstLine string

	if hasFirstLine {
		firstLine = scanner.Text()
	}

	// Check if the first line matches the expected pattern (e.g., "// app/projects/page.tsx")
	matched, err := regexp.MatchString(`^\/\/\s*\S+`, firstLine)
	if err != nil {
		return err
	}

	if !matched {
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			relativePath = path
		}
		if dryRun {
			fmt.Printf("Would insert path: %s\n", relativePath)
		} else {
			if goOnly && !goModule.CanAddComment(path) {
				if verbose {
					log.Printf("Skipping comment insertion for Go file: %s\n", path)
				}
				return nil
			}
			// Insert the path at the beginning of the file
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			newContent := append([]byte("// "+strings.TrimPrefix(relativePath, "/")+"\n"), content...)
			err = os.WriteFile(path, newContent, 0644)
			if err != nil {
				return err
			}
			fmt.Printf("Path inserted: %s\n", relativePath)
		}
	} else if verbose {
		fmt.Printf("Path already present: %s\n", path)
	}

	return nil
}

func processFileRun(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if verbose {
			log.Printf("Empty file or error reading: %s\n", path)
		}
		return nil // Empty file or error
	}

	firstLine := scanner.Text()
	if verbose {
		log.Printf("Processing file: %s\nFirst line: %s\n", path, firstLine)
	}

	// More flexible regex to match comments
	re := regexp.MustCompile(`^(?://|#|;)\s*(\S.*)$`)
	match := re.FindStringSubmatch(firstLine)

	if len(match) > 1 {
		// Extract the file path from the comment
		commentPath := strings.TrimSpace(match[1])
		
		// Remove the extension from the commentPath if it exists
		commentPathWithoutExt := strings.TrimSuffix(commentPath, filepath.Ext(commentPath))
		
		// Create the new file name by replacing slashes with hyphens
		newFileName := strings.ReplaceAll(commentPathWithoutExt, "/", "-")
		
		// Add the original file extension
		newFileName += filepath.Ext(path)
		
		// Create the .bb directory if it doesn't exist
		bbDir := filepath.Join(rootPath, ".bb")
		if err := os.MkdirAll(bbDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating .bb directory: %v", err)
		}

		newFilePath := filepath.Join(bbDir, newFileName)

		// Copy the file to the new location
		if err := copyFile(path, newFilePath); err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", path, newFilePath, err)
		}

		fmt.Printf("Copied %s to %s\n", path, newFilePath)
	} else if verbose {
		log.Printf("No matching comment found in: %s\n", path)
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
func init() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show which files would be modified without actually changing them")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().BoolVarP(&runMode, "run", "r", false, "Copy files with path comments to .bb folder")
	rootCmd.Flags().BoolVarP(&nodeOnly, "node", "n", false, "Process only Node.js files (.js, .ts, .jsx, .mjs, .cjs)")
	rootCmd.Flags().BoolVarP(&goOnly, "go", "g", false, "Process only Go files (.go, go.mod, go.sum)")
}

func main() {
	// Initialize the Node module
	nodeModule = node_module.New()
	// Initialize the Go module
	goModule = go_module.New()

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v\n", err)
	}
}