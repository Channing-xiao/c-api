package api

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"github.com/gin-gonic/gin"
)

// ListHitLogs 命中日志列表
func ListHitLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	action, _ := strconv.Atoi(c.DefaultQuery("action", "0"))
	riskLevel, _ := strconv.Atoi(c.DefaultQuery("risk_level", "0"))
	direction, _ := strconv.Atoi(c.DefaultQuery("direction", "0"))
	modelName := c.DefaultQuery("model_name", "")
	ruleID, _ := strconv.ParseInt(c.DefaultQuery("rule_id", "0"), 10, 64)
	groupID, _ := strconv.ParseInt(c.DefaultQuery("group_id", "0"), 10, 64)
	startTime, _ := strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64)
	endTime, _ := strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64)

	logs, total, err := model.ListHitLogs(page, pageSize, model.HitLogFilter{
		UserID:    userID,
		Action:    action,
		RiskLevel: riskLevel,
		Direction: direction,
		ModelName: modelName,
		RuleID:    ruleID,
		GroupID:   groupID,
		StartTime: startTime,
		EndTime:   endTime,
	})
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: dto.ListResponse{
		Items:    logs,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}})
}
