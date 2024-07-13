# BigBrain (BB) - File Path Processor

BigBrain (BB) is a Go-based command-line tool designed to process and manage file paths in your projects. It's particularly useful for preparing project files for AI RAG (Retrieval-Augmented Generation) systems by adding file path comments to the beginning of each file.

## Build Instructions

To build the BigBrain tool, follow these steps:

1. Ensure you have Go installed on your system (version 1.16 or later recommended).
2. Clone this repository or navigate to the project directory.
3. Run the following command to build the executable:

   ```
   go build -o bb main.go
   ```

   This will create an executable named `bb` in your current directory.

## Usage

Run the tool using the following command structure:

```
./bb [flags] [path]
```

If no path is specified, BB will use the current directory.

### Flags

- `-n, --node`: Search for Node.js related files (ts, js, jsx, mjs, cjs)
- `-d, --dry-run`: Show which files would be modified without actually changing them
- `-v, --verbose`: Enable verbose logging
- `-r, --run`: Copy files with path comments to .bb folder

### Examples

1. Process all files in the current directory (dry run):
   ```
   ./bb --dry-run .
   ```

2. Process only Node.js files in a specific directory:
   ```
   ./bb --node /path/to/project
   ```

3. Run in verbose mode and copy matching files to .bb folder:
   ```
   ./bb --run --verbose .
   ```

## Use Case: AI RAG Preparation

BigBrain is designed to facilitate the preparation of project files for AI Retrieval-Augmented Generation (RAG) systems. It does this by:

1. Adding file path comments to the beginning of each processed file.
2. Optionally copying files with specific path comments to a dedicated `.bb` folder.

This preparation makes it easier for AI systems to understand the structure and context of your project, enhancing the AI's ability to provide accurate and relevant information when queried about your codebase.

## Gitignore Integration

BB respects `.gitignore` rules and has built-in ignore patterns for common development artifacts. This ensures that only relevant files are processed and included in your AI RAG dataset.

## Note

Always run the tool with the `--dry-run` flag first to review which files would be modified before making actual changes to your project files.