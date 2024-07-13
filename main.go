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
)

var (
	rootPath    string
	nodeFiles   bool
	dryRun      bool
	verbose     bool
	runMode     bool
	ignoreRules []string
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

	// Ignore 'public' and '.next' folders for Node.js projects
	if nodeFiles && (relPath == "public" || strings.HasPrefix(relPath, "public/") ||
		relPath == ".next" || strings.HasPrefix(relPath, ".next/")) {
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
	Use:   "bigbrain [path]",
	Short: "Recursively search for and update file paths",
	Args:  cobra.MaximumNArgs(1),
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

			ext := filepath.Ext(path)
			if nodeFiles && (ext != ".ts" && ext != ".js" && ext != ".jsx" && ext != ".mjs" && ext != ".cjs") {
				if verbose {
					log.Printf("Skipping %s (not a Node.js file)\n", path)
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
		newFileName := strings.ReplaceAll(commentPath, "/", "-") + filepath.Ext(path)
		newFilePath := filepath.Join(rootPath, ".bb", newFileName)

		// Copy the file to the new location
		err = copyFile(path, newFilePath)
		if err != nil {
			log.Printf("Error copying file %s to %s: %v\n", path, newFilePath, err)
			return err
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
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.Flags().BoolVarP(&nodeFiles, "node", "n", false, "Search for Node.js related files (ts, js, jsx, mjs, cjs)")
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show which files would be modified without actually changing them")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().BoolVarP(&runMode, "run", "r", false, "Copy files with path comments to .bb folder")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v\n", err)
	}
}
