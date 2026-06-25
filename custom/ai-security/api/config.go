package api

import (
	"net/http"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"github.com/QuantumNous/new-api/custom/ai-security/service"
	"github.com/gin-gonic/gin"
)

// GetConfigs 获取配置
func GetConfigs(c *gin.Context) {
	configs, err := model.ConfigMap()
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	// 确保默认值存在
	if _, ok := configs["ai_security_enabled"]; !ok {
		configs["ai_security_enabled"] = "true"
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: configs})
}

// UpdateConfigs 更新配置
func UpdateConfigs(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := model.UpdateConfigs(req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	service.InvalidateConfigCache()

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "配置更新成功"})
}
