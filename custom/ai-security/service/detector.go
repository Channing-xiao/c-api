package service

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/engine"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
)

// Detector 检测编排器
type Detector struct {
	engines []engine.Engine
}

var (
	detector     *Detector
	detectorOnce sync.Once
)

// GetDetector 获取单例检测器
func GetDetector() *Detector {
	detectorOnce.Do(func() {
		aiModel := model.GetStringConfig(constant.ConfigKeyAIModel, constant.DefaultAIModel)
		timeoutSec := model.GetIntConfig(constant.ConfigKeyAITimeout, constant.DefaultAITimeoutSeconds)
		detector = NewDetector(aiModel, time.Duration(timeoutSec)*time.Second)
	})
	return detector
}

// NewDetector 创建检测器
func NewDetector(aiModel string, aiTimeout time.Duration) *Detector {
	return &Detector{
		engines: []engine.Engine{
			engine.KeywordEngine{},
			engine.RegexEngine{},
			engine.NEREngine{},
			engine.NewAIEngine(aiModel, aiTimeout),
		},
	}
}

// Detect 对内容执行多引擎并行检测
func (d *Detector) Detect(content string, ctx engine.Context) *dto.DetectionResult {
	if !model.IsEnabled() {
		return &dto.DetectionResult{Detected: false, Action: constant.ActionAllow, ActionName: "allow"}
	}

	// 仅加载作用域匹配当前方向的策略对应的规则
	rules := d.loadRulesForUser(ctx.UserID, ctx.Direction)
	if len(rules) == 0 {
		return &dto.DetectionResult{Detected: false, Action: constant.ActionAllow, ActionName: "allow"}
	}

	var wg sync.WaitGroup
	results := make([]*dto.DetectionResult, len(d.engines))
	var mu sync.Mutex

	for i, eng := range d.engines {
		wg.Add(1)
		go func(index int, e engine.Engine) {
			defer wg.Done()
			res := e.Detect(content, rules, ctx)
			mu.Lock()
			results[index] = res
			mu.Unlock()
		}(i, eng)
	}

	wg.Wait()
	return mergeResults(results, content)
}

// loadRulesForUser 加载用户策略对应分组的规则；仅纳入作用域匹配当前方向的策略
func (d *Detector) loadRulesForUser(userID int, direction int) []*model.Rule {
	policies, err := d.loadPolicies(userID)
	if err != nil || len(policies) == 0 {
		return nil
	}

	var groupIDs []int64
	for _, p := range policies {
		if !policyMatchesDirection(p.Scope, direction) {
			continue
		}
		groupIDs = append(groupIDs, p.GroupID)
	}
	if len(groupIDs) == 0 {
		return nil
	}

	if cached, ok := GetCachedRulesByGroups(groupIDs); ok {
		return cached
	}

	rules, err := model.ListActiveRulesByGroupIDs(groupIDs)
	if err != nil {
		common.SysError("[ai-security] load rules failed: " + err.Error())
		return nil
	}

	SetCachedRulesByGroups(groupIDs, rules)
	return rules
}

// policyMatchesDirection 判断策略作用域是否覆盖当前检测方向
func policyMatchesDirection(scope int, direction int) bool {
	switch scope {
	case constant.ScopeBoth:
		return true
	case constant.ScopeRequestOnly:
		return direction == constant.DirectionRequest
	case constant.ScopeResponseOnly:
		return direction == constant.DirectionResponse
	default:
		// 未知作用域按双向处理，避免漏检
		return true
	}
}

// loadPolicies 加载用户策略（带缓存）
func (d *Detector) loadPolicies(userID int) ([]*model.Policy, error) {
	if cached, ok := GetCachedPolicies(userID); ok {
		return cached, nil
	}

	var policies []*model.Policy
	err := newapimodel.DB.Where("user_id = ? AND status = ?", userID, constant.StatusEnabled).
		Order("priority ASC, id DESC").Find(&policies).Error
	if err != nil {
		return nil, err
	}

	SetCachedPolicies(userID, policies)
	return policies, nil
}

// mergeResults 合并多引擎结果，按动作优先级与风险分计算最终结果
func mergeResults(results []*dto.DetectionResult, content string) *dto.DetectionResult {
	final := &dto.DetectionResult{
		Detected: false,
		Action:   constant.ActionAllow,
		Matches:  []*dto.MatchResult{},
	}

	for _, res := range results {
		if res == nil || !res.Detected {
			continue
		}
		final.Detected = true
		final.Matches = append(final.Matches, res.Matches...)

		if final.Action == 0 || constant.GetActionPriority(res.Action) > constant.GetActionPriority(final.Action) {
			final.Action = res.Action
			final.RiskScore = res.RiskScore
			final.RiskLevel = res.RiskLevel
		}
	}

	final.ActionName = constant.GetActionName(final.Action)
	return final
}
