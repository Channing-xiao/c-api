package api

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/service"
	"github.com/gin-gonic/gin"
)

// ListGroups 获取分组列表
func ListGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	parentID, _ := strconv.ParseInt(c.DefaultQuery("parent_id", "-1"), 10, 64)
	name := c.DefaultQuery("name", "")

	groups, total, err := service.ListGroups(page, pageSize, status, parentID, name)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: dto.ListResponse{
		Items:    groups,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}})
}

// CreateGroup 创建分组
func CreateGroup(c *gin.Context) {
	var req service.GroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	group, err := service.CreateGroup(&req)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "分组创建成功", Data: group})
}

// UpdateGroup 更新分组
func UpdateGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req service.GroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.UpdateGroup(id, &req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "分组更新成功"})
}

// UpdateGroupStatus 更新分组状态
func UpdateGroupStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status int `json:"status" binding:"oneof=0 1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if err := service.UpdateGroupStatus(id, req.Status); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "状态更新成功"})
}

// DeleteGroup 删除分组
func DeleteGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := service.DeleteGroup(id); err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "分组删除成功"})
}

// CopyGroup 复制分组
func CopyGroup(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	group, err := service.CopyGroup(id)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: "分组复制成功", Data: group})
}
