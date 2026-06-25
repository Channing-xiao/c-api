package service

import (
	"strconv"
	"strings"
	"time"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
)

// OfficialSyncResult 官方敏感词同步结果
type OfficialSyncResult struct {
	ImportedCount int    `json:"imported_count"`
	SkippedCount  int    `json:"skipped_count"`
	GroupID       int64  `json:"group_id"`
	Message       string `json:"message"`
}

// SyncOfficialSensitiveWords 将官方 sensitive-words 单向导入为 ai-security 规则
// 官方数据只读，不修改；导入后生成独立的 aisec_rules 记录
func SyncOfficialSensitiveWords(adminID int) (*OfficialSyncResult, error) {
	result := &OfficialSyncResult{}

	// 查找或创建"官方敏感词导入"分组
	group, err := getOrCreateOfficialImportGroup()
	if err != nil {
		return nil, err
	}
	result.GroupID = group.ID

	words := setting.SensitiveWords
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}

		code := "official:" + word
		exists, err := model.GetRuleByCode(code)
		if err == nil && exists != nil {
			result.SkippedCount++
			continue
		}

		rule := &model.Rule{
			Code:        code,
			GroupID:     group.ID,
			Name:        "官方导入: " + word,
			Type:        constant.RuleTypeKeyword,
			Content:     word,
			Action:      constant.ActionBlock,
			Priority:    0,
			RiskScore:   constant.DefaultRiskScore,
			Status:      constant.StatusEnabled,
			IsSeed:      false,
			CreatedAt:   time.Now().Unix(),
			UpdatedAt:   time.Now().Unix(),
		}
		if err := newapimodel.DB.Create(rule).Error; err != nil {
			return nil, err
		}
		result.ImportedCount++
	}

	InvalidateRuleCache()
	result.Message = "官方敏感词导入完成"
	return result, nil
}

func getOrCreateOfficialImportGroup() (*model.Group, error) {
	var group model.Group
	err := newapimodel.DB.Where("name = ?", "官方敏感词导入").First(&group).Error
	if err == nil {
		return &group, nil
	}

	group = model.Group{
		Name:        "官方敏感词导入",
		Description: "从官方 Sensitive Words 单向导入的规则",
		ParentID:    0,
		Depth:       0,
		Path:        "",
		Status:      constant.StatusEnabled,
		SortOrder:   0,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}
	if err := newapimodel.DB.Create(&group).Error; err != nil {
		return nil, err
	}
	group.Path = "/" + strconv.FormatInt(group.ID, 10)
	_ = newapimodel.DB.Model(&group).Update("path", group.Path)
	return &group, nil
}
