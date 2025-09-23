package tools

import (
	"github.com/cloudwego/eino/schema"
)

// GetAnalyzerTools returns all available tools for the analyzer
func GetAnalyzerTools() []*schema.ToolInfo {
	return []*schema.ToolInfo{
		{
			Name: "read_file",
			Desc: "Reads and returns the content of a specified file. For log analysis, it's recommended to start with reverse=true and limit=100 to read the last 100 lines where recent errors typically appear. The tool supports forward and reverse reading, pagination, and automatic path resolution. Always use absolute paths for file access.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"absolute_path": {
					Type:     schema.String,
					Desc:     "The absolute path to the file to read (e.g., '/home/user/project/file.txt'). Relative paths are not supported. You must provide an absolute path.",
					Required: true,
				},
				"offset": {
					Type:     schema.Integer,
					Desc:     "Optional: 0-based line number to start reading from. When reverse=true, offset is counted from the end (0=last line, 1=second to last). When reverse=false, offset is from the beginning. Requires 'limit' to be set.",
					Required: false,
				},
				"limit": {
					Type:     schema.Integer,
					Desc:     "Optional: Maximum number of lines to read. Recommended: 100 lines for initial log analysis. Use with 'offset' for pagination. If omitted, reads up to 200 lines.",
					Required: false,
				},
				"reverse": {
					Type:     schema.Boolean,
					Desc:     "Optional: If true, read from the end of the file backwards. RECOMMENDED for log analysis as recent errors appear at the end. Default: false (forward reading).",
					Required: false,
				},
			}),
		},
		{
			Name: "search_file_content",
			Desc: "Searches for a regular expression pattern within the content of files in a specified directory. Can filter files by a glob pattern. Returns the lines containing matches, along with their file paths and line numbers. Useful for finding specific error patterns, configuration issues, or code references.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"pattern": {
					Type:     schema.String,
					Desc:     "The regular expression (regex) pattern to search for within file contents (e.g., 'Exception', 'Error', 'OutOfMemoryError', 'ClassNotFoundException').",
					Required: true,
				},
				"path": {
					Type:     schema.String,
					Desc:     "Optional: The absolute path to the directory to search within. If omitted, searches the current working directory.",
					Required: false,
				},
				"include": {
					Type:     schema.String,
					Desc:     "Optional: A glob pattern to filter which files are searched (e.g., '*.log', '*.java', '*.properties'). If omitted, searches all text files.",
					Required: false,
				},
			}),
		},
	}
}
