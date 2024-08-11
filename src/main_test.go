package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/The-Mines/BigBrain/pkg/file_processor"
	"github.com/The-Mines/BigBrain/pkg/node_module"
	"github.com/The-Mines/BigBrain/pkg/go_module"
	"io"
)

func TestMainCommandFunctionality(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bb-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("NoArguments", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		
		os.Args = []string{"bb"}
		os.Chdir(tmpDir)

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		main()

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		assert.Contains(t, string(out), "Processing directory:")
	})

	t.Run("SpecificPath", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		
		os.Args = []string{"bb", tmpDir}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		main()

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		assert.Contains(t, string(out), "Processing directory: "+tmpDir)
	})

	t.Run("InvalidPath", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		
		invalidPath := "/path/does/not/exist"
		os.Args = []string{"bb", invalidPath}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		main()

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stderr = oldStderr

		assert.Contains(t, string(out), "Error during processing")
	})
}

func TestFlagFunctionality(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bb-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	createSampleFiles(t, tmpDir)

	t.Run("DryRun", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		
		os.Args = []string{"bb", "--dry-run", tmpDir}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		main()

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		assert.Contains(t, string(out), "Would insert path:")
		assert.NotContains(t, string(out), "Path inserted:")
	})

	t.Run("Verbose", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		
		os.Args = []string{"bb", "--verbose", tmpDir}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		main()

		w.Close()
		out, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		assert.Contains(t, string(out), "Ignoring hidden file/directory:")
	})}




func TestFileProcessing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bb-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	nodeModule := node_module.New()
	goModule := go_module.New()
	processor := fileprocessor.New(tmpDir, false, nodeModule, goModule, false, false)

	t.Run("ProcessFileWithoutComment", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "test.js")
		err := os.WriteFile(filePath, []byte("console.log('Hello');"), 0644)
		assert.NoError(t, err)

		err = processor.ProcessFile(filePath, false)
		assert.NoError(t, err)

		content, err := os.ReadFile(filePath)
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(string(content), "// test.js"))
	})

	t.Run("ProcessFileWithExistingComment", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "existing.js")
		err := os.WriteFile(filePath, []byte("// existing.js\nconsole.log('Hello');"), 0644)
		assert.NoError(t, err)

		err = processor.ProcessFile(filePath, false)
		assert.NoError(t, err)

		content, err := os.ReadFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, "// existing.js\nconsole.log('Hello');", string(content))
	})
}

func TestGitignoreFunctionality(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bb-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err = os.WriteFile(gitignorePath, []byte("ignored.txt\nignored_dir/"), 0644)
	assert.NoError(t, err)

	ignoredFile := filepath.Join(tmpDir, "ignored.txt")
	err = os.WriteFile(ignoredFile, []byte("This should be ignored"), 0644)
	assert.NoError(t, err)

	ignoredDir := filepath.Join(tmpDir, "ignored_dir")
	err = os.Mkdir(ignoredDir, 0755)
	assert.NoError(t, err)

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"bb", "--verbose", tmpDir}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	assert.Contains(t, string(out), "Ignoring: "+ignoredFile)
	assert.Contains(t, string(out), "Ignoring: "+ignoredDir)
}

func TestNodeModule(t *testing.T) {
	nodeModule := node_module.New()

	t.Run("IsNodeFile", func(t *testing.T) {
		assert.True(t, nodeModule.IsNodeFile("test.js"))
		assert.True(t, nodeModule.IsNodeFile("test.ts"))
		assert.False(t, nodeModule.IsNodeFile("test.go"))
	})

	t.Run("ShouldIgnoreNodePath", func(t *testing.T) {
		assert.True(t, nodeModule.ShouldIgnoreNodePath("public/index.html"))
		assert.True(t, nodeModule.ShouldIgnoreNodePath(".next/build"))
		assert.False(t, nodeModule.ShouldIgnoreNodePath("src/components/App.js"))
	})
}

func TestGoModule(t *testing.T) {
	goModule := go_module.New()

	t.Run("IsGoFile", func(t *testing.T) {
		assert.True(t, goModule.IsGoFile("main.go"))
		assert.True(t, goModule.IsGoFile("go.mod"))
		assert.False(t, goModule.IsGoFile("main.js"))
	})

	t.Run("ShouldIgnoreGoPath", func(t *testing.T) {
		assert.True(t, goModule.ShouldIgnoreGoPath("vendor/github.com/example/pkg"))
		assert.True(t, goModule.ShouldIgnoreGoPath(".git/config"))
		assert.False(t, goModule.ShouldIgnoreGoPath("pkg/file_processor/file_processor.go"))
	})

	t.Run("CanAddComment", func(t *testing.T) {
		assert.True(t, goModule.CanAddComment("main.go"))
		assert.False(t, goModule.CanAddComment("go.mod"))
		assert.False(t, goModule.CanAddComment("go.sum"))
	})
}

func createSampleFiles(t *testing.T, dir string) {
	files := []struct {
		name    string
		content string
	}{
		{"test.js", "console.log('Hello');"},
		{"test.go", "package main\n\nfunc main() {}"},
		{"test.ts", "const greeting: string = 'Hello';"},
	}

	for _, f := range files {
		err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.content), 0644)
		assert.NoError(t, err)
	}
}