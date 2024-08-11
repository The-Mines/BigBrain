// pkg/file_processor/file_processor.go
package fileprocessor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/The-Mines/BigBrain/pkg/go_module"
	"github.com/The-Mines/BigBrain/pkg/node_module"
)

type FileProcessor interface {
	ProcessFile(path string, dryRun bool) error
	ProcessFileRun(path string) error
}

type fileProcessor struct {
	rootPath    string
	verbose     bool
	nodeModule  node_module.NodeModule
	goModule    go_module.GoModule
	nodeOnly    bool
	goOnly      bool
}

func New(rootPath string, verbose bool, nodeModule node_module.NodeModule, goModule go_module.GoModule, nodeOnly, goOnly bool) FileProcessor {
	return &fileProcessor{
		rootPath:   rootPath,
		verbose:    verbose,
		nodeModule: nodeModule,
		goModule:   goModule,
		nodeOnly:   nodeOnly,
		goOnly:     goOnly,
	}
}

func (fp *fileProcessor) ProcessFile(path string, dryRun bool) error {
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
		relativePath, err := filepath.Rel(fp.rootPath, path)
		if err != nil {
			relativePath = path
		}
		if dryRun {
			fmt.Printf("Would insert path: %s\n", relativePath)
		} else {
			if fp.goOnly && !fp.goModule.CanAddComment(path) {
				if fp.verbose {
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
	} else if fp.verbose {
		fmt.Printf("Path already present: %s\n", path)
	}

	return nil
}

func (fp *fileProcessor) ProcessFileRun(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if fp.verbose {
			log.Printf("Empty file or error reading: %s\n", path)
		}
		return nil // Empty file or error
	}

	firstLine := scanner.Text()
	if fp.verbose {
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
		bbDir := filepath.Join(fp.rootPath, ".bb")
		if err := os.MkdirAll(bbDir, os.ModePerm); err != nil {
			return fmt.Errorf("error creating .bb directory: %v", err)
		}

		newFilePath := filepath.Join(bbDir, newFileName)

		// Copy the file to the new location
		if err := copyFile(path, newFilePath); err != nil {
			return fmt.Errorf("error copying file %s to %s: %v", path, newFilePath, err)
		}

		fmt.Printf("Copied %s to %s\n", path, newFilePath)
	} else if fp.verbose {
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