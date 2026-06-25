package api

import (
	"net/http"

	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/service"
	"github.com/gin-gonic/gin"
)

// SyncOfficialSensitiveWords 同步官方敏感词
func SyncOfficialSensitiveWords(c *gin.Context) {
	adminID := c.GetInt("id")
	result, err := service.SyncOfficialSensitiveWords(adminID)
	if err != nil {
		c.JSON(http.StatusOK, dto.Response{Success: false, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.Response{Success: true, Message: result.Message, Data: result})
}
