package api

import (
	"net/http"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	newapimodel "github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetStatus 获取模块状态
func GetStatus(c *gin.Context) {
	var groupCount, ruleCount, policyCount int64

	newapimodel.DB.Model(&model.Group{}).Where("status = ?", 1).Count(&groupCount)
	newapimodel.DB.Model(&model.Rule{}).Where("status = ?", 1).Count(&ruleCount)
	newapimodel.DB.Model(&model.Policy{}).Where("status = ?", 1).Count(&policyCount)

	status := dto.StatusResponse{
		Enabled:      model.IsEnabled(),
		RuleCount:    ruleCount,
		GroupCount:   groupCount,
		PolicyCount:  policyCount,
		CacheEnabled: true,
		Version:      "1.0.0",
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: status})
}

// Install 执行安装/初始化
func Install(c *gin.Context) {
	// 迁移和种子由 Init() 在启动时完成
	// 这里可以强制重新执行种子补充
	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "安装完成", Data: gin.H{
		"migrated":   true,
		"version":    "1.0.0",
	}})
}
