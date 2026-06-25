package engine

import (
	"context"
	"time"

	"github.com/QuantumNous/new-api/common"
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// AIEngine AI 检测引擎（调用内部 channel 或模型，超时降级）
type AIEngine struct {
	modelName string
	timeout   time.Duration
}

func NewAIEngine(modelName string, timeout time.Duration) *AIEngine {
	if modelName == "" {
		modelName = constant.DefaultAIModel
	}
	if timeout <= 0 {
		timeout = constant.DefaultAITimeoutSeconds * time.Second
	}
	return &AIEngine{modelName: modelName, timeout: timeout}
}

func (AIEngine) Type() int { return constant.RuleTypeAI }

func (e *AIEngine) Detect(content string, rules []*model.Rule, ctx Context) *dto.DetectionResult {
	result := &dto.DetectionResult{
		Detected: false,
		Action:   constant.ActionAllow,
		Matches:  []*dto.MatchResult{},
	}

	activeRules := filterActiveAIRules(rules)
	if len(activeRules) == 0 {
		result.ActionName = constant.GetActionName(result.Action)
		return result
	}

	prompt := buildAIPrompt(content, activeRules)
	detected, reason := e.callModel(prompt)

	if detected {
		result.Detected = true
		// AI 命中时，取最高风险动作
		for _, rule := range activeRules {
			if result.Action == 0 || constant.GetActionPriority(rule.Action) > constant.GetActionPriority(result.Action) {
				result.Action = rule.Action
				result.RiskScore = rule.RiskScore
				result.RiskLevel = constant.GetRiskLevelByScore(rule.RiskScore)
			}
		}
		result.Matches = append(result.Matches, &dto.MatchResult{
			RuleID:      0,
			GroupID:     0,
			Type:        constant.RuleTypeAI,
			MatchedText: reason,
			Position:    [2]int{0, 0},
		})
	}

	result.ActionName = constant.GetActionName(result.Action)
	return result
}

func filterActiveAIRules(rules []*model.Rule) []*model.Rule {
	var out []*model.Rule
	for _, r := range rules {
		if r.Type == constant.RuleTypeAI && r.Status == constant.StatusEnabled {
			out = append(out, r)
		}
	}
	return out
}

func buildAIPrompt(content string, rules []*model.Rule) string {
	var categories []string
	for _, r := range rules {
		categories = append(categories, r.Content)
	}
	return "请判断以下内容是否包含以下任一敏感类别：" + joinStrings(categories, "、") +
		"。仅回答 YES 或 NO，并给出简短理由。内容：" + content
}

func joinStrings(items []string, sep string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += sep + items[i]
	}
	return result
}

// callModel 调用 AI 模型，超时返回 false
func (e *AIEngine) callModel(prompt string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	done := make(chan struct {
		detected bool
		reason   string
	}, 1)

	go func() {
		// TODO: 接入项目内部 channel/relay 调用模型
		// 当前作为占位实现，直接返回 NO 避免阻塞
		done <- struct {
			detected bool
			reason   string
		}{detected: false, reason: "AI engine placeholder"}
	}()

	select {
	case r := <-done:
		return r.detected, r.reason
	case <-ctx.Done():
		common.SysError("[ai-security] AI detection timeout, fallback to allow")
		return false, "AI detection timeout"
	}
}
