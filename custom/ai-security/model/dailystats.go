package model

import "github.com/QuantumNous/new-api/model"

// DailyStat 每日统计
type DailyStat struct {
	ID              int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Date            string `gorm:"column:date;size:10;not null;uniqueIndex:idx_aisec_daily_date"`
	TotalDetected   int    `gorm:"column:total_detected;type:int;default:0"`
	TotalBlocked    int    `gorm:"column:total_blocked;type:int;default:0"`
	TotalMasked     int    `gorm:"column:total_masked;type:int;default:0"`
	TotalAlerted    int    `gorm:"column:total_alerted;type:int;default:0"`
	TotalReviewed   int    `gorm:"column:total_reviewed;type:int;default:0"`
	TopCategory     string `gorm:"column:top_category;type:text;default:null"`
	TopUser         string `gorm:"column:top_user;type:text;default:null"`
	TopModel        string `gorm:"column:top_model;type:text;default:null"`
	CreatedAt       int64  `gorm:"column:created_at;type:bigint;default:0"`
	UpdatedAt       int64  `gorm:"column:updated_at;type:bigint;default:0"`
}

func (DailyStat) TableName() string { return "aisec_daily_stats" }

// GetDailyStatByDate 根据日期获取统计
func GetDailyStatByDate(date string) (*DailyStat, error) {
	var stat DailyStat
	err := model.DB.Where("date = ?", date).First(&stat).Error
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

// UpsertDailyStat 更新或创建每日统计
func UpsertDailyStat(stat *DailyStat) error {
	var existing DailyStat
	err := model.DB.Where("date = ?", stat.Date).First(&existing).Error
	if err != nil {
		return model.DB.Create(stat).Error
	}
	return model.DB.Model(&existing).Updates(stat).Error
}

// ListDailyStatsRange 按日期区间（含端点，格式 2006-01-02）升序返回历史统计
func ListDailyStatsRange(startDate, endDate string) ([]*DailyStat, error) {
	var stats []*DailyStat
	err := model.DB.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date ASC").Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}
