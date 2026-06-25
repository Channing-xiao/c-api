package engine

import (
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// Context 检测上下文
type Context struct {
	UserID    int
	TokenID   int
	ChannelID int
	RequestID string
	ModelName string
	Direction int
	IP        string
}

// Engine 检测引擎接口
type Engine interface {
	// Detect 对内容执行检测，返回匹配结果
	Detect(content string, rules []*model.Rule, ctx Context) *dto.DetectionResult
	// Type 返回引擎处理的规则类型
	Type() int
}

// BaseEngine 提供通用辅助方法
type BaseEngine struct{}

// NewResult 创建空结果
func (BaseEngine) NewResult() *dto.DetectionResult {
	return &dto.DetectionResult{
		Detected: false,
		Action:   0,
		Matches:  []*dto.MatchResult{},
	}
}
