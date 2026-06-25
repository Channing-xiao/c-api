package service

import (
	"errors"
	"strings"
	"time"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
)

// RuleRequest 规则请求
type RuleRequest struct {
	GroupID     int64  `json:"group_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=128"`
	Type        int    `json:"type" binding:"required,oneof=1 2 3 4"`
	Content     string `json:"content" binding:"required"`
	ExtraConfig string `json:"extra_config"`
	Action      int    `json:"action" binding:"required,oneof=1 2 3 4 5"`
	Priority    int    `json:"priority"`
	RiskScore   int    `json:"risk_score" binding:"min=0,max=100"`
}

// GetRuleByID 获取规则
func GetRuleByID(id int64) (*model.Rule, error) {
	return model.GetRuleByID(id)
}

// ListRules 获取规则列表
func ListRules(page, pageSize int, groupID int64, ruleType, status int) ([]*model.Rule, int64, error) {
	var rules []*model.Rule
	var total int64

	db := newapimodel.DB.Model(&model.Rule{})
	if groupID > 0 {
		db = db.Where("group_id = ?", groupID)
	}
	if ruleType > 0 {
		db = db.Where("type = ?", ruleType)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("priority DESC, id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// CreateRule 创建规则
func CreateRule(req *RuleRequest) (*model.Rule, error) {
	if _, err := model.GetGroupByID(req.GroupID); err != nil {
		return nil, errors.New("分组不存在")
	}

	rule := &model.Rule{
		GroupID:     req.GroupID,
		Name:        req.Name,
		Type:        req.Type,
		Content:     req.Content,
		ExtraConfig: req.ExtraConfig,
		Action:      req.Action,
		Priority:    req.Priority,
		RiskScore:   req.RiskScore,
		Status:      constant.StatusEnabled,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	if err := newapimodel.DB.Create(rule).Error; err != nil {
		return nil, err
	}

	InvalidateRuleCache()
	return rule, nil
}

// UpdateRule 更新规则
func UpdateRule(id int64, req *RuleRequest) error {
	rule, err := model.GetRuleByID(id)
	if err != nil {
		return errors.New("规则不存在")
	}

	if _, err := model.GetGroupByID(req.GroupID); err != nil {
		return errors.New("分组不存在")
	}

	rule.GroupID = req.GroupID
	rule.Name = req.Name
	rule.Type = req.Type
	rule.Content = req.Content
	rule.ExtraConfig = req.ExtraConfig
	rule.Action = req.Action
	rule.Priority = req.Priority
	rule.RiskScore = req.RiskScore
	rule.UpdatedAt = time.Now().Unix()
	if rule.IsSeed {
		rule.SeedModifiedAt = time.Now().Unix()
	}

	if err := newapimodel.DB.Save(rule).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// DeleteRule 删除规则
func DeleteRule(id int64) error {
	rule, err := model.GetRuleByID(id)
	if err != nil {
		return errors.New("规则不存在")
	}

	if err := newapimodel.DB.Delete(rule).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// UpdateRuleStatus 更新规则状态
func UpdateRuleStatus(id int64, status int) error {
	rule, err := model.GetRuleByID(id)
	if err != nil {
		return errors.New("规则不存在")
	}

	rule.Status = status
	rule.UpdatedAt = time.Now().Unix()
	if err := newapimodel.DB.Save(rule).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// BatchDeleteRules 批量删除规则
func BatchDeleteRules(ids []int64) error {
	if err := newapimodel.DB.Where("id IN (?)", ids).Delete(&model.Rule{}).Error; err != nil {
		return err
	}
	InvalidateRuleCache()
	return nil
}

// BatchUpdateRuleStatus 批量更新规则状态
func BatchUpdateRuleStatus(ids []int64, status int) error {
	if err := newapimodel.DB.Model(&model.Rule{}).Where("id IN (?)", ids).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now().Unix()}).Error; err != nil {
		return err
	}
	InvalidateRuleCache()
	return nil
}

// TestRule 测试规则
func TestRule(id int64, content string) (*dto.DetectionResult, error) {
	rule, err := model.GetRuleByID(id)
	if err != nil {
		return nil, errors.New("规则不存在")
	}

	if rule.Status != constant.StatusEnabled {
		return &dto.DetectionResult{Detected: false, Action: constant.ActionAllow}, nil
	}

	// 这里简化处理，实际应调用 engine
	result := &dto.DetectionResult{
		Detected: false,
		Action:   constant.ActionAllow,
	}

	// 简单 keyword/regex 测试示例
	if rule.Type == constant.RuleTypeKeyword {
		keywords := strings.Split(rule.Content, ",")
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" && strings.Contains(content, keyword) {
				idx := strings.Index(content, keyword)
				result.Detected = true
				result.Action = rule.Action
				result.RiskScore = rule.RiskScore
				result.RiskLevel = constant.GetRiskLevelByScore(rule.RiskScore)
				result.Matches = append(result.Matches, &dto.MatchResult{
					RuleID:      rule.ID,
					GroupID:     rule.GroupID,
					Type:        rule.Type,
					MatchedText: keyword,
					Position:    [2]int{idx, idx + len(keyword)},
				})
				break
			}
		}
	}

	result.ActionName = constant.GetActionName(result.Action)
	return result, nil
}
