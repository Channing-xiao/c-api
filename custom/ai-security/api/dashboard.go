package api

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/service"
	"github.com/gin-gonic/gin"
)

// GetDashboard 获取安全看板数据
func GetDashboard(c *gin.Context) {
	startTime, _ := strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64)
	endTime, _ := strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64)
	userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	groupID, _ := strconv.ParseInt(c.DefaultQuery("group_id", "0"), 10, 64)
	ruleID, _ := strconv.ParseInt(c.DefaultQuery("rule_id", "0"), 10, 64)

	data, err := service.GetDashboard(service.DashboardFilter{
		StartTime: startTime,
		EndTime:   endTime,
		UserID:    userID,
		GroupID:   groupID,
		RuleID:    ruleID,
	})
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: data})
}

// GetDailyTrend 获取历史每日趋势（从 aisec_daily_stats 读取）
func GetDailyTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	data, err := service.GetDailyTrend(days)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Data: data})
}
