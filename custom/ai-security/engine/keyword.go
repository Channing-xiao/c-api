package engine

import (
	"strings"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// KeywordEngine 关键词引擎
type KeywordEngine struct{}

func (KeywordEngine) Type() int { return constant.RuleTypeKeyword }

func (e KeywordEngine) Detect(content string, rules []*model.Rule, ctx Context) *dto.DetectionResult {
	result := &dto.DetectionResult{
		Detected: false,
		Action:   constant.ActionAllow,
		Matches:  []*dto.MatchResult{},
	}

	lowerContent := strings.ToLower(content)

	for _, rule := range rules {
		if rule.Type != constant.RuleTypeKeyword || rule.Status != constant.StatusEnabled {
			continue
		}

		keywords := strings.Split(rule.Content, ",")
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}

			lowerKeyword := strings.ToLower(keyword)
			idx := strings.Index(lowerContent, lowerKeyword)
			if idx == -1 {
				continue
			}

			result.Detected = true
			result.Matches = append(result.Matches, &dto.MatchResult{
				RuleID:      rule.ID,
				GroupID:     rule.GroupID,
				Type:        rule.Type,
				MatchedText: content[idx : idx+len(keyword)],
				Position:    [2]int{idx, idx + len(keyword)},
			})

			if result.Action == 0 || constant.GetActionPriority(rule.Action) > constant.GetActionPriority(result.Action) {
				result.Action = rule.Action
				result.RiskScore = rule.RiskScore
				result.RiskLevel = constant.GetRiskLevelByScore(rule.RiskScore)
			}
		}
	}

	result.ActionName = constant.GetActionName(result.Action)
	return result
}
