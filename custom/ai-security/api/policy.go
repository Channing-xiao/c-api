package api

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/service"
	"github.com/gin-gonic/gin"
)

// ListPolicies 获取策略列表
func ListPolicies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))

	policies, total, err := service.ListPolicies(page, pageSize, userID, status)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: dto.ListResponse{
		Items:    policies,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}})
}

// CreatePolicy 创建策略
func CreatePolicy(c *gin.Context) {
	var req service.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	policy, err := service.CreatePolicy(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "策略创建成功", Data: policy})
}

// UpdatePolicy 更新策略
func UpdatePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req service.PolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.UpdatePolicy(id, &req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "策略更新成功"})
}

// DeletePolicy 删除策略
func DeletePolicy(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := service.DeletePolicy(id); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "策略删除成功"})
}
