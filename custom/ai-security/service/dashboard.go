package service

import (
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

// DashboardData 看板数据
type DashboardData struct {
	Summary          DashboardSummary          `json:"summary"`
	RiskDistribution DashboardRiskDistribution `json:"risk_distribution"`
	TopCategories    []DashboardCategoryItem   `json:"top_categories"`
	TopUsers         []DashboardUserItem       `json:"top_users"`
	TopModels        []DashboardModelItem      `json:"top_models"`
}

type DashboardSummary struct {
	TotalDetections    int64 `json:"total_detections"`
	TotalInterceptions int64 `json:"total_interceptions"`
	TotalAlerts        int64 `json:"total_alerts"`
	TodayDetections    int64 `json:"today_detections"`
	TodayInterceptions int64 `json:"today_interceptions"`
}

type DashboardRiskDistribution struct {
	Low      int64 `json:"low"`
	Medium   int64 `json:"medium"`
	High     int64 `json:"high"`
	Critical int64 `json:"critical"`
}

type DashboardCategoryItem struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type DashboardUserItem struct {
	UserID   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Count    int64  `json:"count"`
}

type DashboardModelItem struct {
	ModelName string `json:"model_name"`
	Count     int64  `json:"count"`
}

// DashboardFilter 看板过滤条件
type DashboardFilter struct {
	StartTime int64
	EndTime   int64
	UserID    int
	GroupID   int64
	RuleID    int64
}

// dashboardScope 构建带时间范围与维度过滤的基础查询
func dashboardScope(f DashboardFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Model(&model.HitLog{})
		if f.StartTime > 0 {
			db = db.Where("created_at >= ?", f.StartTime)
		}
		if f.EndTime > 0 {
			db = db.Where("created_at <= ?", f.EndTime)
		}
		if f.UserID > 0 {
			db = db.Where("user_id = ?", f.UserID)
		}
		if f.GroupID > 0 {
			db = db.Where("group_id = ?", f.GroupID)
		}
		if f.RuleID > 0 {
			db = db.Where("rule_id = ?", f.RuleID)
		}
		return db
	}
}

// GetDashboard 获取看板数据
func GetDashboard(f DashboardFilter) (*DashboardData, error) {
	data := &DashboardData{
		TopCategories: []DashboardCategoryItem{},
		TopUsers:      []DashboardUserItem{},
		TopModels:     []DashboardModelItem{},
	}
	scope := dashboardScope(f)

	if err := newapimodel.DB.Scopes(scope).Count(&data.Summary.TotalDetections).Error; err != nil {
		return nil, err
	}
	if err := newapimodel.DB.Scopes(scope).Where("action = ?", constant.ActionBlock).
		Count(&data.Summary.TotalInterceptions).Error; err != nil {
		return nil, err
	}
	if err := newapimodel.DB.Scopes(scope).Where("action = ?", constant.ActionAlert).
		Count(&data.Summary.TotalAlerts).Error; err != nil {
		return nil, err
	}

	// 风险等级分布
	var riskCounts []struct {
		RiskLevel int   `gorm:"column:risk_level"`
		Count     int64 `gorm:"column:count"`
	}
	if err := newapimodel.DB.Scopes(scope).
		Select("risk_level, COUNT(*) as count").Group("risk_level").Scan(&riskCounts).Error; err != nil {
		return nil, err
	}
	for _, rc := range riskCounts {
		switch rc.RiskLevel {
		case constant.RiskLevelLow:
			data.RiskDistribution.Low = rc.Count
		case constant.RiskLevelMedium:
			data.RiskDistribution.Medium = rc.Count
		case constant.RiskLevelHigh:
			data.RiskDistribution.High = rc.Count
		case constant.RiskLevelCritical:
			data.RiskDistribution.Critical = rc.Count
		}
	}

	// 热门分组（按 group_id）
	var categoryRows []struct {
		GroupID int64 `gorm:"column:group_id"`
		Count   int64 `gorm:"column:count"`
	}
	if err := newapimodel.DB.Scopes(scope).
		Select("group_id, COUNT(*) as count").Group("group_id").
		Order("count DESC").Limit(10).Scan(&categoryRows).Error; err != nil {
		return nil, err
	}
	for _, row := range categoryRows {
		name := "未分组"
		if row.GroupID > 0 {
			if g, err := model.GetGroupByID(row.GroupID); err == nil {
				name = g.Name
			}
		}
		data.TopCategories = append(data.TopCategories, DashboardCategoryItem{Category: name, Count: row.Count})
	}

	// 热门用户
	var userRows []struct {
		UserID int   `gorm:"column:user_id"`
		Count  int64 `gorm:"column:count"`
	}
	if err := newapimodel.DB.Scopes(scope).
		Select("user_id, COUNT(*) as count").Group("user_id").
		Order("count DESC").Limit(10).Scan(&userRows).Error; err != nil {
		return nil, err
	}
	for _, row := range userRows {
		data.TopUsers = append(data.TopUsers, DashboardUserItem{UserID: row.UserID, Count: row.Count})
	}

	// 热门模型
	var modelRows []struct {
		ModelName string `gorm:"column:model_name"`
		Count     int64  `gorm:"column:count"`
	}
	if err := newapimodel.DB.Scopes(scope).
		Select("model_name, COUNT(*) as count").Group("model_name").
		Order("count DESC").Limit(10).Scan(&modelRows).Error; err != nil {
		return nil, err
	}
	for _, row := range modelRows {
		data.TopModels = append(data.TopModels, DashboardModelItem{ModelName: row.ModelName, Count: row.Count})
	}

	return data, nil
}
