package service

import (
	"encoding/json"
	"time"

	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// AuditActionType 审计操作类型
const (
	AuditActionCreate       = "create"
	AuditActionUpdate       = "update"
	AuditActionDelete       = "delete"
	AuditActionCopy         = "copy"
	AuditActionEnable       = "enable"
	AuditActionDisable      = "disable"
	AuditActionBatchDelete  = "batch_delete"
	AuditActionBatchStatus  = "batch_status"
	AuditActionSync         = "sync"
)

// AuditTargetType 审计对象类型
const (
	AuditTargetGroup   = "group"
	AuditTargetRule    = "rule"
	AuditTargetPolicy  = "policy"
	AuditTargetConfig  = "config"
)

// CreateAuditLog 创建审计日志
func CreateAuditLog(userID int, actionType, targetType string, targetID int64, oldValue, newValue interface{}) {
	oldJSON, _ := json.Marshal(oldValue)
	newJSON, _ := json.Marshal(newValue)

	log := &model.AuditLog{
		UserID:     userID,
		ActionType: actionType,
		TargetType: targetType,
		TargetID:   targetID,
		OldValue:   string(oldJSON),
		NewValue:   string(newJSON),
		CreatedAt:  time.Now().Unix(),
	}

	_ = model.CreateAuditLog(log)
}
