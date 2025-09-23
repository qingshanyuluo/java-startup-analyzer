package tools

import (
	"github.com/cloudwego/eino/schema"
)

// GetAnalyzerTools returns all available tools for the analyzer
func GetAnalyzerTools() []*schema.ToolInfo {
	return []*schema.ToolInfo{
		{
			Name: "tail",
			Desc: "Tails a file and returns the last N lines. Use this tool to read log files for analysis.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"file_path": {
					Type:     schema.String,
					Desc:     "Path to the file to tail",
					Required: true,
				},
				"num_lines": {
					Type:     schema.Integer,
					Desc:     "Number of lines to return from the end of the file (default: 50)",
					Required: false,
				},
			}),
		},
	}
}
