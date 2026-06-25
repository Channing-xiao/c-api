package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
)

var dailyStatsSchedulerOnce sync.Once

// StartDailyStatsScheduler 启动每日统计聚合调度器（幂等，仅启动一次）。
// 启动时立即回填昨日与今日，随后每小时刷新一次：
//   - 今日行保持最新（看板历史趋势可读到当天的滚动值）
//   - 昨日行在跨零点后一小时内被最终确认（即使进程在零点时未运行）
func StartDailyStatsScheduler() {
	dailyStatsSchedulerOnce.Do(func() {
		go dailyStatsLoop()
	})
}

func dailyStatsLoop() {
	aggregateRecentDays()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if !model.IsEnabled() {
			continue
		}
		aggregateRecentDays()
	}
}

// aggregateRecentDays 聚合昨日与今日（Upsert 幂等）
func aggregateRecentDays() {
	now := time.Now()
	if err := AggregateDailyStats(now.Add(-24 * time.Hour)); err != nil {
		common.SysError("[ai-security] aggregate yesterday stats failed: " + err.Error())
	}
	if err := AggregateDailyStats(now); err != nil {
		common.SysError("[ai-security] aggregate today stats failed: " + err.Error())
	}
}

// GetDailyTrend 读取最近 days 天的历史统计（从 aisec_daily_stats）。
// days <= 0 时默认 7 天，最大 365 天。
func GetDailyTrend(days int) ([]*model.DailyStat, error) {
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	now := time.Now()
	endDate := now.Format("2006-01-02")
	startDate := now.Add(-time.Duration(days-1) * 24 * time.Hour).Format("2006-01-02")
	return model.ListDailyStatsRange(startDate, endDate)
}

// AggregateDailyStats 按天聚合命中日志并写入 aisec_daily_stats
func AggregateDailyStats(date time.Time) error {
	dateStr := date.Format("2006-01-02")
	start := date.Truncate(24 * time.Hour).Unix()
	end := start + 86400 - 1

	var stats struct {
		TotalDetected int64
		TotalBlocked  int64
		TotalMasked   int64
		TotalAlerted  int64
		TotalReviewed int64
	}

	db := newapimodel.DB.Model(&model.HitLog{}).Where("created_at >= ? AND created_at <= ?", start, end)
	if err := db.Count(&stats.TotalDetected).Error; err != nil {
		return err
	}
	if err := db.Where("action = ?", constant.ActionBlock).Count(&stats.TotalBlocked).Error; err != nil {
		return err
	}
	if err := db.Where("action = ?", constant.ActionMask).Count(&stats.TotalMasked).Error; err != nil {
		return err
	}
	if err := db.Where("action = ?", constant.ActionAlert).Count(&stats.TotalAlerted).Error; err != nil {
		return err
	}
	if err := db.Where("action = ?", constant.ActionReview).Count(&stats.TotalReviewed).Error; err != nil {
		return err
	}

	topCat, err := topCategoryForDate(start, end)
	if err != nil {
		return err
	}
	topUser, err := topUserForDate(start, end)
	if err != nil {
		return err
	}
	topModel, err := topModelForDate(start, end)
	if err != nil {
		return err
	}

	stat := &model.DailyStat{
		Date:          dateStr,
		TotalDetected: int(stats.TotalDetected),
		TotalBlocked:  int(stats.TotalBlocked),
		TotalMasked:   int(stats.TotalMasked),
		TotalAlerted:  int(stats.TotalAlerted),
		TotalReviewed: int(stats.TotalReviewed),
		TopCategory:   topCat,
		TopUser:       topUser,
		TopModel:      topModel,
		CreatedAt:     time.Now().Unix(),
		UpdatedAt:     time.Now().Unix(),
	}
	return model.UpsertDailyStat(stat)
}

// AggregateYesterdayStats 聚合昨日数据
func AggregateYesterdayStats() error {
	yesterday := time.Now().Add(-24 * time.Hour)
	return AggregateDailyStats(yesterday)
}

func topCategoryForDate(start, end int64) (string, error) {
	var row struct {
		GroupID int64 `gorm:"column:group_id"`
		Count   int64 `gorm:"column:count"`
	}
	err := newapimodel.DB.Model(&model.HitLog{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Select("group_id, COUNT(*) as count").Group("group_id").
		Order("count DESC").Limit(1).Scan(&row).Error
	if err != nil || row.GroupID == 0 {
		return "", err
	}
	g, err := model.GetGroupByID(row.GroupID)
	if err != nil {
		return fmt.Sprintf("group:%d", row.GroupID), nil
	}
	return g.Name, nil
}

func topUserForDate(start, end int64) (string, error) {
	var row struct {
		UserID int   `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	err := newapimodel.DB.Model(&model.HitLog{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Select("user_id, COUNT(*) as count").Group("user_id").
		Order("count DESC").Limit(1).Scan(&row).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", row.UserID), nil
}

func topModelForDate(start, end int64) (string, error) {
	var row struct {
		ModelName string `gorm:"column:model_name"`
		Count     int64  `gorm:"column:count"`
	}
	err := newapimodel.DB.Model(&model.HitLog{}).
		Where("created_at >= ? AND created_at <= ?", start, end).
		Select("model_name, COUNT(*) as count").Group("model_name").
		Order("count DESC").Limit(1).Scan(&row).Error
	if err != nil {
		return "", err
	}
	return row.ModelName, nil
}
