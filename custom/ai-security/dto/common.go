package dto

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ListResponse 列表响应结构
type ListResponse struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// MatchResult 检测结果匹配项
type MatchResult struct {
	RuleID      int64  `json:"rule_id"`
	GroupID     int64  `json:"group_id"`
	Type        int    `json:"type"`
	MatchedText string `json:"matched_text"`
	Position    [2]int `json:"position"`
}

// DetectionResult 检测结果
type DetectionResult struct {
	Detected         bool          `json:"detected"`
	Action           int           `json:"action"`
	ActionName       string        `json:"action_name"`
	RiskScore        int           `json:"risk_score"`
	RiskLevel        int           `json:"risk_level"`
	ProcessedContent string        `json:"processed_content,omitempty"`
	Matches          []*MatchResult `json:"matches,omitempty"`
}

// CheckRequest 检测请求
type CheckRequest struct {
	UserID    int    `json:"user_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
	ModelName string `json:"model_name"`
}

// StatusResponse 模块状态响应
type StatusResponse struct {
	Enabled      bool   `json:"enabled"`
	RuleCount    int64  `json:"rule_count"`
	GroupCount   int64  `json:"group_count"`
	PolicyCount  int64  `json:"policy_count"`
	CacheEnabled bool   `json:"cache_enabled"`
	Version      string `json:"version"`
}
