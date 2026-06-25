package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"github.com/gin-gonic/gin"
)

// ExportHitLogs 导出命中日志（CSV）
func ExportHitLogs(c *gin.Context) {
	format := c.DefaultQuery("format", "csv")
	userID, _ := strconv.Atoi(c.DefaultQuery("user_id", "0"))
	action, _ := strconv.Atoi(c.DefaultQuery("action", "0"))
	riskLevel, _ := strconv.Atoi(c.DefaultQuery("risk_level", "0"))
	direction, _ := strconv.Atoi(c.DefaultQuery("direction", "0"))
	modelName := c.DefaultQuery("model_name", "")
	ruleID, _ := strconv.ParseInt(c.DefaultQuery("rule_id", "0"), 10, 64)
	groupID, _ := strconv.ParseInt(c.DefaultQuery("group_id", "0"), 10, 64)
	startTime, _ := strconv.ParseInt(c.DefaultQuery("start_time", "0"), 10, 64)
	endTime, _ := strconv.ParseInt(c.DefaultQuery("end_time", "0"), 10, 64)

	filter := model.HitLogFilter{
		UserID:    userID,
		Action:    action,
		RiskLevel: riskLevel,
		Direction: direction,
		ModelName: modelName,
		RuleID:    ruleID,
		GroupID:   groupID,
		StartTime: startTime,
		EndTime:   endTime,
	}

	// 导出上限 10000 条
	logs, _, err := model.ListHitLogs(1, 10000, filter)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	if format == "excel" {
		// Excel 导出复用 CSV（兼容 Excel 打开），设置 BOM 以便中文正常显示
		buf := buildCSV(logs, true)
		c.Header("Content-Disposition", "attachment; filename=ai_security_logs.csv")
		c.Data(http.StatusOK, "application/vnd.ms-excel", buf.Bytes())
		return
	}

	buf := buildCSV(logs, true)
	c.Header("Content-Disposition", "attachment; filename=ai_security_logs.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
}

func buildCSV(logs []*model.HitLog, withBOM bool) *bytes.Buffer {
	buf := &bytes.Buffer{}
	if withBOM {
		buf.WriteString("\xEF\xBB\xBF")
	}
	w := csv.NewWriter(buf)
	_ = w.Write([]string{
		"ID", "RequestID", "UserID", "ModelName", "Direction",
		"Action", "RiskLevel", "RiskScore", "MatchedText", "IP", "CreatedAt",
	})
	for _, log := range logs {
		_ = w.Write([]string{
			strconv.FormatInt(log.ID, 10),
			log.RequestID,
			strconv.Itoa(log.UserID),
			log.ModelName,
			directionName(log.Direction),
			constant.GetActionName(log.Action),
			constant.GetRiskLevelName(log.RiskLevel),
			strconv.Itoa(log.RiskScore),
			log.MatchedText,
			log.IP,
			fmt.Sprintf("%d", log.CreatedAt),
		})
	}
	w.Flush()
	return buf
}

func directionName(direction int) string {
	if direction == constant.DirectionResponse {
		return "response"
	}
	return "request"
}
