package analyzer

import "time"

// AnalysisResult 分析结果
type AnalysisResult struct {
	// 基本信息
	Timestamp    time.Time     `json:"timestamp"`
	LogSize      int           `json:"log_size"`
	AnalysisTime time.Duration `json:"analysis_time"`

	// 分析结果
	Status        string   `json:"status"`         // success, failure, warning
	ErrorType     string   `json:"error_type"`     // 错误类型
	ErrorMessage  string   `json:"error_message"`  // 错误消息
	RootCause     string   `json:"root_cause"`     // 根本原因
	Suggestions   []string `json:"suggestions"`    // 解决建议
	RelatedErrors []string `json:"related_errors"` // 相关错误

	// 详细信息
	Summary    string            `json:"summary"`    // 分析摘要
	Details    string            `json:"details"`    // 详细分析
	Confidence float64           `json:"confidence"` // 置信度 (0-1)
	Metadata   map[string]string `json:"metadata"`   // 元数据
}

// NewAnalysisResult 创建新的分析结果
func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Timestamp:     time.Now(),
		Suggestions:   make([]string, 0),
		RelatedErrors: make([]string, 0),
		Metadata:      make(map[string]string),
		Confidence:    0.0,
	}
}
