package api

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/service"
	"github.com/gin-gonic/gin"
)

// ListRules 获取规则列表
func ListRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	groupID, _ := strconv.ParseInt(c.DefaultQuery("group_id", "0"), 10, 64)
	ruleType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	rules, total, err := service.ListRules(page, pageSize, groupID, ruleType, status)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: dto.ListResponse{
		Items:    rules,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}})
}

// CreateRule 创建规则
func CreateRule(c *gin.Context) {
	var req service.RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	rule, err := service.CreateRule(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "规则创建成功", Data: rule})
}

// UpdateRule 更新规则
func UpdateRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req service.RuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.UpdateRule(id, &req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "规则更新成功"})
}

// DeleteRule 删除规则
func DeleteRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := service.DeleteRule(id); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "规则删除成功"})
}

// UpdateRuleStatus 更新规则状态
func UpdateRuleStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.UpdateRuleStatus(id, req.Status); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "状态更新成功"})
}

// BatchDeleteRules 批量删除规则
func BatchDeleteRules(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.BatchDeleteRules(req.IDs); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "批量删除成功"})
}

// BatchUpdateRuleStatus 批量更新规则状态
func BatchUpdateRuleStatus(c *gin.Context) {
	var req struct {
		IDs    []int64 `json:"ids" binding:"required"`
		Status int     `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.BatchUpdateRuleStatus(req.IDs, req.Status); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "批量状态更新成功"})
}

// TestRule 测试规则
func TestRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	result, err := service.TestRule(id, req.Content)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: result})
}
