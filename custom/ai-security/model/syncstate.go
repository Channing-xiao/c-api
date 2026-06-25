package model

import "github.com/QuantumNous/new-api/model"

// SyncState 同步状态
type SyncState struct {
	ID           int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Source       string `gorm:"column:source;size:64;not null;uniqueIndex:idx_aisec_sync_source"`
	LastSyncAt   int64  `gorm:"column:last_sync_at;type:bigint;default:0"`
	SyncCount    int    `gorm:"column:sync_count;type:int;default:0"`
	CreatedAt    int64  `gorm:"column:created_at;type:bigint;default:0"`
	UpdatedAt    int64  `gorm:"column:updated_at;type:bigint;default:0"`
}

func (SyncState) TableName() string { return "aisec_sync_state" }

// GetSyncStateBySource 根据来源获取同步状态
func GetSyncStateBySource(source string) (*SyncState, error) {
	var state SyncState
	err := model.DB.Where("source = ?", source).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// SaveSyncState 保存同步状态
func SaveSyncState(state *SyncState) error {
	var existing SyncState
	err := model.DB.Where("source = ?", state.Source).First(&existing).Error
	if err != nil {
		return model.DB.Create(state).Error
	}
	return model.DB.Model(&existing).Updates(state).Error
}
