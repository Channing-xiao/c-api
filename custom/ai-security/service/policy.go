package service

import (
	"errors"
	"time"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
)

// PolicyRequest 策略请求
type PolicyRequest struct {
	UserID         int    `json:"user_id" binding:"required"`
	GroupID        int64  `json:"group_id" binding:"required"`
	Scope          int    `json:"scope" binding:"required,oneof=1 2 3"`
	DefaultAction  int    `json:"default_action" binding:"required,oneof=1 2 3 4 5"`
	CustomResponse string `json:"custom_response"`
	WhitelistIPs   string `json:"whitelist_ips"`
	Priority       int    `json:"priority"`
	Status         int    `json:"status" binding:"oneof=0 1"`
}

// GetPolicyByID 获取策略
func GetPolicyByID(id int64) (*model.Policy, error) {
	return model.GetPolicyByID(id)
}

// ListPolicies 获取策略列表
func ListPolicies(page, pageSize int, userID int, status int) ([]*model.Policy, int64, error) {
	var policies []*model.Policy
	var total int64

	db := newapimodel.DB.Model(&model.Policy{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status >= 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("priority ASC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&policies).Error; err != nil {
		return nil, 0, err
	}

	return policies, total, nil
}

// CreatePolicy 创建策略
func CreatePolicy(req *PolicyRequest) (*model.Policy, error) {
	if _, err := model.GetGroupByID(req.GroupID); err != nil {
		return nil, errors.New("分组不存在")
	}

	if exists, err := model.HasActivePolicyForUserGroup(req.UserID, req.GroupID, 0); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("该用户已存在同一分组的启用策略")
	}

	policy := &model.Policy{
		UserID:         req.UserID,
		GroupID:        req.GroupID,
		Scope:          req.Scope,
		DefaultAction:  req.DefaultAction,
		CustomResponse: req.CustomResponse,
		WhitelistIPs:   req.WhitelistIPs,
		Priority:       req.Priority,
		Status:         req.Status,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
	}

	if policy.Status == 0 {
		policy.Status = constant.StatusEnabled
	}

	if err := newapimodel.DB.Create(policy).Error; err != nil {
		return nil, err
	}

	InvalidateRuleCache()
	return policy, nil
}

// UpdatePolicy 更新策略
func UpdatePolicy(id int64, req *PolicyRequest) error {
	policy, err := model.GetPolicyByID(id)
	if err != nil {
		return errors.New("策略不存在")
	}

	if _, err := model.GetGroupByID(req.GroupID); err != nil {
		return errors.New("分组不存在")
	}

	if policy.Status == constant.StatusEnabled || req.Status == constant.StatusEnabled {
		if exists, err := model.HasActivePolicyForUserGroup(req.UserID, req.GroupID, id); err != nil {
			return err
		} else if exists {
			return errors.New("该用户已存在同一分组的启用策略")
		}
	}

	policy.UserID = req.UserID
	policy.GroupID = req.GroupID
	policy.Scope = req.Scope
	policy.DefaultAction = req.DefaultAction
	policy.CustomResponse = req.CustomResponse
	policy.WhitelistIPs = req.WhitelistIPs
	policy.Priority = req.Priority
	policy.Status = req.Status
	policy.UpdatedAt = time.Now().Unix()

	if err := newapimodel.DB.Save(policy).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// DeletePolicy 删除策略
func DeletePolicy(id int64) error {
	policy, err := model.GetPolicyByID(id)
	if err != nil {
		return errors.New("策略不存在")
	}

	if err := newapimodel.DB.Delete(policy).Error; err != nil {
		return err
	}

	InvalidateRuleCache()
	return nil
}

// ListActivePoliciesByUserID 获取用户启用的策略
func ListActivePoliciesByUserID(userID int) ([]*model.Policy, error) {
	var policies []*model.Policy
	err := newapimodel.DB.Where("user_id = ? AND status = ?", userID, constant.StatusEnabled).
		Order("priority ASC, id DESC").Find(&policies).Error
	return policies, err
}
