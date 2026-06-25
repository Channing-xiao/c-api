package model

import "github.com/QuantumNous/new-api/model"

// Rule 检测规则
type Rule struct {
	ID             int64  `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Code           string `json:"code" gorm:"column:code;size:128;uniqueIndex:idx_aisec_rule_code"`
	GroupID        int64  `json:"group_id" gorm:"column:group_id;type:bigint;not null;index:idx_aisec_rule_group"`
	Name           string `json:"name" gorm:"column:name;size:128;not null"`
	Type           int    `json:"type" gorm:"column:type;type:int;not null;index:idx_aisec_rule_type"`
	Content        string `json:"content" gorm:"column:content;type:text;not null"`
	ExtraConfig    string `json:"extra_config" gorm:"column:extra_config;type:text;default:null"`
	Action         int    `json:"action" gorm:"column:action;type:int;default:4"`
	Priority       int    `json:"priority" gorm:"column:priority;type:int;default:0"`
	RiskScore      int    `json:"risk_score" gorm:"column:risk_score;type:int;default:50"`
	Status         int    `json:"status" gorm:"column:status;type:int;default:1;index:idx_aisec_rule_status"`
	IsSeed         bool   `json:"is_seed" gorm:"column:is_seed;default:false"`
	SeedModifiedAt int64  `json:"seed_modified_at" gorm:"column:seed_modified_at;type:bigint;default:0"`
	CreatedAt      int64  `json:"created_at" gorm:"column:created_at;type:bigint;default:0"`
	UpdatedAt      int64  `json:"updated_at" gorm:"column:updated_at;type:bigint;default:0"`
}

func (Rule) TableName() string { return "aisec_rules" }

// GetRuleByID 根据 ID 获取规则
func GetRuleByID(id int64) (*Rule, error) {
	var rule Rule
	err := model.DB.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetRuleByCode 根据编码获取规则
func GetRuleByCode(code string) (*Rule, error) {
	var rule Rule
	err := model.DB.Where("code = ?", code).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListActiveRulesByGroupIDs 根据分组 ID 列表获取启用规则
func ListActiveRulesByGroupIDs(groupIDs []int64) ([]*Rule, error) {
	var rules []*Rule
	err := model.DB.Where("group_id IN (?) AND status = ?", groupIDs, 1).
		Order("priority DESC, id ASC").Find(&rules).Error
	return rules, err
}
