package seed

import (
	"fmt"
	"time"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
)

// DefaultRuleSeed 默认规则种子
type DefaultRuleSeed struct {
	Code        string
	GroupName   string
	RuleName    string
	Type        int
	Content     string
	ExtraConfig string
	Action      int
	Priority    int
	RiskScore   int
}

var defaultSeeds = []DefaultRuleSeed{
	// 基础安全策略
	{Code: "basic-political", GroupName: "基础安全策略", RuleName: "政治敏感", Type: constant.RuleTypeKeyword, Content: "敏感政治词1,敏感政治词2", Action: constant.ActionBlock, Priority: 10, RiskScore: 80},
	{Code: "basic-violence", GroupName: "基础安全策略", RuleName: "暴力恐怖", Type: constant.RuleTypeKeyword, Content: "暴力,恐怖,爆炸", Action: constant.ActionBlock, Priority: 10, RiskScore: 85},
	// 个人隐私信息
	{Code: "privacy-phone", GroupName: "个人隐私信息", RuleName: "手机号检测", Type: constant.RuleTypeRegex, Content: `\b1[3-9]\d{9}\b`, ExtraConfig: `{"mask_type":"preserve","preserve_start":3,"preserve_end":4}`, Action: constant.ActionMask, Priority: 5, RiskScore: 60},
	{Code: "privacy-idcard", GroupName: "个人隐私信息", RuleName: "身份证号检测", Type: constant.RuleTypeRegex, Content: `\b\d{17}[\dXx]\b`, Action: constant.ActionMask, Priority: 5, RiskScore: 70},
	// 企业机密
	{Code: "corp-internal", GroupName: "企业机密", RuleName: "内部代号", Type: constant.RuleTypeKeyword, Content: "内部代号,机密项目", Action: constant.ActionBlock, Priority: 8, RiskScore: 75},
	// 合规风险
	{Code: "compliance-medical", GroupName: "合规风险", RuleName: "医疗记录", Type: constant.RuleTypeRegex, Content: `病历|诊断报告|就诊记录`, Action: constant.ActionReview, Priority: 6, RiskScore: 65},
	// Prompt 防护
	{Code: "prompt-jailbreak", GroupName: "Prompt防护", RuleName: "越狱提示", Type: constant.RuleTypeKeyword, Content: "ignore previous instructions,DAN,越狱", Action: constant.ActionBlock, Priority: 10, RiskScore: 90},
}

// InitDefaultRules 初始化默认规则
func InitDefaultRules() error {
	for _, seed := range defaultSeeds {
		if err := ensureGroupAndRule(seed); err != nil {
			return fmt.Errorf("failed to seed rule %s: %w", seed.Code, err)
		}
	}
	return nil
}

func ensureGroupAndRule(seed DefaultRuleSeed) error {
	// 查找或创建分组
	var group model.Group
	err := newapimodel.DB.Where("name = ?", seed.GroupName).First(&group).Error
	if err != nil {
		group = model.Group{
			Name:        seed.GroupName,
			Description: "",
			ParentID:    0,
			Depth:       0,
			Path:        "",
			Status:      constant.StatusEnabled,
			SortOrder:   0,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		if err := newapimodel.DB.Create(&group).Error; err != nil {
			return err
		}
		group.Path = fmt.Sprintf("/%d", group.ID)
		if err := newapimodel.DB.Model(&group).Update("path", group.Path).Error; err != nil {
			return err
		}
	}

	// 查找规则
	var existing model.Rule
	err = newapimodel.DB.Where("code = ?", seed.Code).First(&existing).Error
	if err == nil {
		// 已存在且未被用户修改，不覆盖
		return nil
	}

	// 创建规则
	rule := model.Rule{
		Code:        seed.Code,
		GroupID:     group.ID,
		Name:        seed.RuleName,
		Type:        seed.Type,
		Content:     seed.Content,
		ExtraConfig: seed.ExtraConfig,
		Action:      seed.Action,
		Priority:    seed.Priority,
		RiskScore:   seed.RiskScore,
		Status:      constant.StatusEnabled,
		IsSeed:      true,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	return newapimodel.DB.Create(&rule).Error
}
