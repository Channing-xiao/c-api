package model

import "github.com/QuantumNous/new-api/model"

// HitLog 命中日志
type HitLog struct {
	ID                  int64  `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	RequestID           string `json:"request_id" gorm:"column:request_id;size:64;not null;index:idx_aisec_hit_request"`
	UserID              int    `json:"user_id" gorm:"column:user_id;type:int;not null;index:idx_aisec_hit_user"`
	TokenID             int    `json:"token_id" gorm:"column:token_id;type:int;default:0"`
	ModelName           string `json:"model_name" gorm:"column:model_name;size:128;default:'';index:idx_aisec_hit_model"`
	ChannelID           int    `json:"channel_id" gorm:"column:channel_id;type:int;default:0"`
	Direction           int    `json:"direction" gorm:"column:direction;type:int;default:1"`
	RuleID              int64  `json:"rule_id" gorm:"column:rule_id;type:bigint;default:0;index:idx_aisec_hit_rule"`
	GroupID             int64  `json:"group_id" gorm:"column:group_id;type:bigint;default:0;index:idx_aisec_hit_group"`
	RiskLevel           int    `json:"risk_level" gorm:"column:risk_level;type:int;default:1;index:idx_aisec_hit_risk"`
	RiskScore           int    `json:"risk_score" gorm:"column:risk_score;type:int;default:0"`
	Action              int    `json:"action" gorm:"column:action;type:int;not null;index:idx_aisec_hit_action"`
	MatchedText         string `json:"matched_text" gorm:"column:matched_text;size:500;default:''"`
	HitReason           string `json:"hit_reason" gorm:"column:hit_reason;size:255;default:''"`
	OriginalContentHash string `json:"original_content_hash" gorm:"column:original_content_hash;size:64;default:''"`
	ProcessedContent    string `json:"processed_content" gorm:"column:processed_content;type:text;default:null"`
	IP                  string `json:"ip" gorm:"column:ip;size:64;default:''"`
	CreatedAt           int64  `json:"created_at" gorm:"column:created_at;type:bigint;default:0;index:idx_aisec_hit_created"`
}

func (HitLog) TableName() string { return "aisec_hit_logs" }

// CreateHitLogs 批量创建命中日志
func CreateHitLogs(logs []*HitLog) error {
	if len(logs) == 0 {
		return nil
	}
	return model.DB.CreateInBatches(logs, 100).Error
}

// HitLogFilter 命中日志查询过滤条件
type HitLogFilter struct {
	UserID    int
	Action    int
	RiskLevel int
	Direction int
	ModelName string
	RuleID    int64
	GroupID   int64
	StartTime int64
	EndTime   int64
}

// ListHitLogs 分页查询命中日志
func ListHitLogs(page, pageSize int, filter HitLogFilter) ([]*HitLog, int64, error) {
	var logs []*HitLog
	var total int64

	db := model.DB.Model(&HitLog{})
	if filter.UserID > 0 {
		db = db.Where("user_id = ?", filter.UserID)
	}
	if filter.Action > 0 {
		db = db.Where("action = ?", filter.Action)
	}
	if filter.RiskLevel > 0 {
		db = db.Where("risk_level = ?", filter.RiskLevel)
	}
	if filter.Direction > 0 {
		db = db.Where("direction = ?", filter.Direction)
	}
	if filter.ModelName != "" {
		db = db.Where("model_name = ?", filter.ModelName)
	}
	if filter.RuleID > 0 {
		db = db.Where("rule_id = ?", filter.RuleID)
	}
	if filter.GroupID > 0 {
		db = db.Where("group_id = ?", filter.GroupID)
	}
	if filter.StartTime > 0 {
		db = db.Where("created_at >= ?", filter.StartTime)
	}
	if filter.EndTime > 0 {
		db = db.Where("created_at <= ?", filter.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteHitLogsBefore 删除指定时间之前的命中日志（用于保留期清理）
func DeleteHitLogsBefore(timestamp int64) (int64, error) {
	result := model.DB.Where("created_at < ?", timestamp).Delete(&HitLog{})
	return result.RowsAffected, result.Error
}
