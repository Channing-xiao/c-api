package model

import (
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/model"
)

// Policy 用户策略
type Policy struct {
	ID             int64  `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	UserID         int    `json:"user_id" gorm:"column:user_id;type:int;not null;index:idx_aisec_policy_user"`
	GroupID        int64  `json:"group_id" gorm:"column:group_id;type:bigint;not null;index:idx_aisec_policy_group"`
	Scope          int    `json:"scope" gorm:"column:scope;type:int;default:3"`
	DefaultAction  int    `json:"default_action" gorm:"column:default_action;type:int;default:4"`
	CustomResponse string `json:"custom_response" gorm:"column:custom_response;type:text;default:null"`
	WhitelistIPs   string `json:"whitelist_ips" gorm:"column:whitelist_ips;type:text;default:null"`
	Priority       int    `json:"priority" gorm:"column:priority;type:int;default:0;index:idx_aisec_policy_priority"`
	Status         int    `json:"status" gorm:"column:status;type:int;default:1;index:idx_aisec_policy_status"`
	CreatedAt      int64  `json:"created_at" gorm:"column:created_at;type:bigint;default:0"`
	UpdatedAt      int64  `json:"updated_at" gorm:"column:updated_at;type:bigint;default:0"`
}

func (Policy) TableName() string { return "aisec_policies" }

// GetPolicyByID 根据 ID 获取策略
func GetPolicyByID(id int64) (*Policy, error) {
	var policy Policy
	err := model.DB.Where("id = ?", id).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// HasActivePolicyForUserGroup 检查用户对某分组是否已有启用策略（排除指定 ID）
func HasActivePolicyForUserGroup(userID int, groupID int64, excludeID int64) (bool, error) {
	var count int64
	db := model.DB.Model(&Policy{}).Where("user_id = ? AND group_id = ? AND status = ?", userID, groupID, constant.StatusEnabled)
	if excludeID > 0 {
		db = db.Where("id != ?", excludeID)
	}
	err := db.Count(&count).Error
	return count > 0, err
}
