package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ReadFileInput represents the input parameters for the read_file tool
type ReadFileInput struct {
	AbsolutePath string `json:"absolute_path" description:"The absolute path to the file to read (e.g., '/home/user/project/file.txt'). Relative paths are not supported. You must provide an absolute path."`
	Offset       *int   `json:"offset,omitempty" description:"Optional: 0-based line number to start reading from. When reverse=true, offset is counted from the end (0=last line, 1=second to last). When reverse=false, offset is from the beginning. Requires 'limit' to be set."`
	Limit        *int   `json:"limit,omitempty" description:"Optional: Maximum number of lines to read. Recommended: 100 lines for initial log analysis. Use with 'offset' for pagination. If omitted, reads up to 200 lines."`
	Reverse      *bool  `json:"reverse,omitempty" description:"Optional: If true, read from the end of the file backwards. RECOMMENDED for log analysis as recent errors appear at the end. Default: false (forward reading)."`
}

// ReadFileOutput represents the output of the read_file tool
type ReadFileOutput struct {
	Content    string `json:"content" description:"The content of the file"`
	Truncated  bool   `json:"truncated" description:"Whether the content was truncated due to file size"`
	TotalLines int    `json:"total_lines" description:"Total number of lines in the file (for text files)"`
	ReadLines  int    `json:"read_lines" description:"Number of lines actually read"`
}

// ReadFileTool is a tool that reads file content.
var ReadFileTool tool.InvokableTool

func init() {
	var err error
	ReadFileTool, err = utils.InferTool(
		"read_file",
		"Reads and returns the content of a specified file. For log analysis, start with reverse=true and limit=100 to read the last 100 lines where recent errors typically appear. Supports forward/reverse reading, pagination, and automatic path resolution. Always use absolute paths.",
		readFile,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create read_file tool: %v", err))
	}
}

// readFile reads a file and returns its content.
func readFile(ctx context.Context, input ReadFileInput) (ReadFileOutput, error) {
	// Validate absolute path
	if !filepath.IsAbs(input.AbsolutePath) {
		return ReadFileOutput{}, fmt.Errorf("path must be absolute: %s", input.AbsolutePath)
	}

	// Check if file exists
	fileInfo, err := os.Stat(input.AbsolutePath)
	if err != nil {
		return ReadFileOutput{}, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return ReadFileOutput{}, fmt.Errorf("path is a directory, not a file: %s", input.AbsolutePath)
	}

	// For now, we'll focus on text files
	// TODO: Add support for binary files (images, PDFs) in the future
	file, err := os.Open(input.AbsolutePath)
	if err != nil {
		return ReadFileOutput{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read all lines first to get total count
	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return ReadFileOutput{}, fmt.Errorf("failed to read file: %w", err)
	}

	totalLines := len(allLines)

	// Determine which lines to read
	var startLine, endLine int
	var truncated bool

	// Check if reverse reading is requested
	reverse := input.Reverse != nil && *input.Reverse

	if input.Offset != nil && input.Limit != nil {
		// Read specific range
		if reverse {
			// Reverse reading: offset is from the end
			startLine = totalLines - *input.Offset - *input.Limit
			if startLine < 0 {
				startLine = 0
			}
			endLine = totalLines - *input.Offset
			if endLine > totalLines {
				endLine = totalLines
			}
		} else {
			// Forward reading: offset is from the beginning
			startLine = *input.Offset
			if startLine < 0 {
				startLine = 0
			}
			if startLine >= totalLines {
				return ReadFileOutput{
					Content:    "",
					Truncated:  false,
					TotalLines: totalLines,
					ReadLines:  0,
				}, nil
			}
			endLine = startLine + *input.Limit
			if endLine > totalLines {
				endLine = totalLines
			}
		}
	} else if input.Offset != nil {
		// Read from offset to end
		if reverse {
			// Reverse reading: from offset to beginning
			endLine = totalLines - *input.Offset
			if endLine > totalLines {
				endLine = totalLines
			}
			startLine = 0
		} else {
			// Forward reading: from offset to end
			startLine = *input.Offset
			if startLine < 0 {
				startLine = 0
			}
			if startLine >= totalLines {
				return ReadFileOutput{
					Content:    "",
					Truncated:  false,
					TotalLines: totalLines,
					ReadLines:  0,
				}, nil
			}
			endLine = totalLines
		}
	} else {
		// Read entire file, but limit to reasonable size
		const maxLines = 200
		if reverse {
			// Read last N lines
			if totalLines > maxLines {
				startLine = totalLines - maxLines
				endLine = totalLines
				truncated = true
			} else {
				startLine = 0
				endLine = totalLines
			}
		} else {
			// Read first N lines
			startLine = 0
			if totalLines > maxLines {
				endLine = maxLines
				truncated = true
			} else {
				endLine = totalLines
			}
		}
	}

	// Extract the requested lines
	selectedLines := allLines[startLine:endLine]
	content := strings.Join(selectedLines, "\n")
	readLines := len(selectedLines)

	return ReadFileOutput{
		Content:    content,
		Truncated:  truncated,
		TotalLines: totalLines,
		ReadLines:  readLines,
	}, nil
}
