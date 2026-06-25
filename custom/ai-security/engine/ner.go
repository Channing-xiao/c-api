package engine

import (
	"strings"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// NEREngine 命名实体识别引擎（轻量占位实现）
type NEREngine struct{}

func (NEREngine) Type() int { return constant.RuleTypeNER }

// 默认敏感实体类型标签
var defaultSensitiveEntities = []string{
	"PERSON", "ORG", "GPE", "PHONE", "EMAIL", "ID_CARD", "BANK_CARD", "ADDRESS",
}

func (e NEREngine) Detect(content string, rules []*model.Rule, ctx Context) *dto.DetectionResult {
	result := &dto.DetectionResult{
		Detected: false,
		Action:   constant.ActionAllow,
		Matches:  []*dto.MatchResult{},
	}

	// 收集所有配置的实体标签
	targetSet := make(map[string]struct{})
	for _, tag := range defaultSensitiveEntities {
		targetSet[tag] = struct{}{}
	}

	for _, rule := range rules {
		if rule.Type != constant.RuleTypeNER || rule.Status != constant.StatusEnabled {
			continue
		}

		ruleTags := parseTags(rule.Content)
		for _, tag := range ruleTags {
			targetSet[tag] = struct{}{}
		}
	}

	// 简单占位：检测形如 <TAG>...</TAG> 的伪标签或常见模式
	for tag := range targetSet {
		openTag := "<" + tag + ">"
		closeTag := "</" + tag + ">"
		start := 0
		for {
			idx := strings.Index(content[start:], openTag)
			if idx == -1 {
				break
			}
			idx += start
			end := strings.Index(content[idx:], closeTag)
			if end == -1 {
				break
			}
			end += idx + len(closeTag)

			result.Detected = true
			result.Matches = append(result.Matches, &dto.MatchResult{
				RuleID:      0,
				GroupID:     0,
				Type:        constant.RuleTypeNER,
				MatchedText: content[idx:end],
				Position:    [2]int{idx, end},
			})

			start = end
		}
	}

	if result.Detected && len(rules) > 0 {
		// 取最高优先级动作与风险分
		for _, rule := range rules {
			if rule.Type != constant.RuleTypeNER || rule.Status != constant.StatusEnabled {
				continue
			}
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

func parseTags(content string) []string {
	if strings.TrimSpace(content) == "" {
		return []string{}
	}
	parts := strings.Split(content, ",")
	var tags []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
