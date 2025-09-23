package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// TailInput represents the input parameters for the tail tool
type TailInput struct {
	FilePath string `json:"file_path" description:"Path to the file to tail"`
	NumLines int    `json:"num_lines" description:"Number of lines to return from the end of the file"`
}

// TailOutput represents the output of the tail tool
type TailOutput struct {
	Lines string `json:"lines" description:"The last N lines of the file"`
}

// TailTool is a tool that tails a file.
var TailTool tool.InvokableTool

func init() {
	var err error
	TailTool, err = utils.InferTool(
		"tail",
		"Tails a file and returns the last N lines.",
		tailFile,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create tail tool: %v", err))
	}
}

// tailFile tails a file and returns the last N lines.
func tailFile(ctx context.Context, input TailInput) (TailOutput, error) {
	file, err := os.Open(input.FilePath)
	if err != nil {
		return TailOutput{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return TailOutput{}, fmt.Errorf("failed to read file: %w", err)
	}

	numLines := input.NumLines
	if numLines > len(lines) {
		numLines = len(lines)
	}

	start := len(lines) - numLines
	result := strings.Join(lines[start:], "\n")

	return TailOutput{Lines: result}, nil
}
