package engine

import (
	"regexp"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// RegexEngine 正则引擎
type RegexEngine struct{}

func (RegexEngine) Type() int { return constant.RuleTypeRegex }

func (e RegexEngine) Detect(content string, rules []*model.Rule, ctx Context) *dto.DetectionResult {
	result := &dto.DetectionResult{
		Detected: false,
		Action:   constant.ActionAllow,
		Matches:  []*dto.MatchResult{},
	}

	for _, rule := range rules {
		if rule.Type != constant.RuleTypeRegex || rule.Status != constant.StatusEnabled {
			continue
		}

		re, err := regexp.Compile(rule.Content)
		if err != nil {
			continue
		}

		loc := re.FindStringIndex(content)
		if loc == nil {
			continue
		}

		result.Detected = true
		result.Matches = append(result.Matches, &dto.MatchResult{
			RuleID:      rule.ID,
			GroupID:     rule.GroupID,
			Type:        rule.Type,
			MatchedText: content[loc[0]:loc[1]],
			Position:    [2]int{loc[0], loc[1]},
		})

		if result.Action == 0 || constant.GetActionPriority(rule.Action) > constant.GetActionPriority(result.Action) {
			result.Action = rule.Action
			result.RiskScore = rule.RiskScore
			result.RiskLevel = constant.GetRiskLevelByScore(rule.RiskScore)
		}
	}

	result.ActionName = constant.GetActionName(result.Action)
	return result
}
