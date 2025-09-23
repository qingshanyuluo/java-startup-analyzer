package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// SearchFileContentInput represents the input parameters for the search_file_content tool
type SearchFileContentInput struct {
	Pattern string `json:"pattern" description:"The regular expression (regex) pattern to search for within file contents (e.g., 'function\\s+myFunction', 'import\\s+\\{.*\\}\\s+from\\s+.*')."`
	Path    string `json:"path,omitempty" description:"Optional: The absolute path to the directory to search within. If omitted, searches the current working directory."`
	Include string `json:"include,omitempty" description:"Optional: A glob pattern to filter which files are searched (e.g., '*.js', '*.{ts,tsx}', 'src/**'). If omitted, searches all files (respecting potential global ignores)."`
}

// SearchResult represents a single search match
type SearchResult struct {
	FilePath   string `json:"file_path" description:"The path of the file containing the match"`
	LineNumber int    `json:"line_number" description:"The line number where the match was found"`
	Content    string `json:"content" description:"The content of the line containing the match"`
}

// SearchFileContentOutput represents the output of the search_file_content tool
type SearchFileContentOutput struct {
	Results      []SearchResult `json:"results" description:"List of search results with file paths, line numbers, and content"`
	TotalFiles   int            `json:"total_files" description:"Total number of files searched"`
	TotalMatches int            `json:"total_matches" description:"Total number of matches found"`
}

// SearchFileContentTool is a tool that searches for patterns in files.
var SearchFileContentTool tool.InvokableTool

func init() {
	var err error
	SearchFileContentTool, err = utils.InferTool(
		"search_file_content",
		"Searches for a regular expression pattern within the content of files in a specified directory. Can filter files by a glob pattern. Returns the lines containing matches, along with their file paths and line numbers. Useful for finding specific error patterns, configuration issues, or code references.",
		searchFileContent,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create search_file_content tool: %v", err))
	}
}

// searchFileContent searches for a pattern in files within a directory.
func searchFileContent(ctx context.Context, input SearchFileContentInput) (SearchFileContentOutput, error) {
	// Validate pattern
	if input.Pattern == "" {
		return SearchFileContentOutput{}, fmt.Errorf("pattern cannot be empty")
	}

	// Compile regex pattern
	regex, err := regexp.Compile(input.Pattern)
	if err != nil {
		return SearchFileContentOutput{}, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Determine search directory
	searchDir := input.Path
	if searchDir == "" {
		searchDir = "."
	}

	// Convert to absolute path
	absSearchDir, err := filepath.Abs(searchDir)
	if err != nil {
		return SearchFileContentOutput{}, fmt.Errorf("failed to resolve search directory: %w", err)
	}

	// Check if directory exists
	if _, err := os.Stat(absSearchDir); os.IsNotExist(err) {
		return SearchFileContentOutput{}, fmt.Errorf("search directory does not exist: %s", absSearchDir)
	}

	var results []SearchResult
	totalFiles := 0

	// Walk through directory
	err = filepath.Walk(absSearchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Apply include filter if specified
		if input.Include != "" {
			matched, err := filepath.Match(input.Include, filepath.Base(path))
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		// Skip binary files and common non-text files
		if isBinaryFile(path) {
			return nil
		}

		totalFiles++

		// Search in file
		fileResults, err := searchInFile(path, regex)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: failed to search in file %s: %v\n", path, err)
			return nil
		}

		results = append(results, fileResults...)
		return nil
	})

	if err != nil {
		return SearchFileContentOutput{}, fmt.Errorf("failed to walk directory: %w", err)
	}

	return SearchFileContentOutput{
		Results:      results,
		TotalFiles:   totalFiles,
		TotalMatches: len(results),
	}, nil
}

// searchInFile searches for a pattern in a specific file.
func searchInFile(filePath string, regex *regexp.Regexp) ([]SearchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []SearchResult
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		if regex.MatchString(line) {
			results = append(results, SearchResult{
				FilePath:   filePath,
				LineNumber: lineNumber,
				Content:    strings.TrimSpace(line),
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// isBinaryFile checks if a file is likely to be binary.
func isBinaryFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".dat": true, ".db": true, ".sqlite": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".bmp": true, ".ico": true, ".svg": true, ".webp": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true,
		".xlsx": true, ".ppt": true, ".pptx": true,
	}

	return binaryExts[ext]
}
