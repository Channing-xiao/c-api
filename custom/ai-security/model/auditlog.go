package model

import "github.com/QuantumNous/new-api/model"

// AuditLog 操作审计日志
type AuditLog struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int    `gorm:"column:user_id;type:int;not null;index:idx_aisec_audit_user"`
	ActionType string `gorm:"column:action_type;size:32;not null;index:idx_aisec_audit_action"`
	TargetType string `gorm:"column:target_type;size:32;not null"`
	TargetID   int64  `gorm:"column:target_id;type:bigint;not null"`
	OldValue   string `gorm:"column:old_value;type:text;default:null"`
	NewValue   string `gorm:"column:new_value;type:text;default:null"`
	CreatedAt  int64  `gorm:"column:created_at;type:bigint;default:0;index:idx_aisec_audit_created"`
}

func (AuditLog) TableName() string { return "aisec_audit_logs" }

// CreateAuditLog 创建审计日志
func CreateAuditLog(log *AuditLog) error {
	return model.DB.Create(log).Error
}
