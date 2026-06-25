package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/engine"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
	"github.com/QuantumNous/new-api/custom/ai-security/service"

	"github.com/gin-gonic/gin"
)

const maxBodySize = 10 * 1024 * 1024

// CheckRequest ai-security 请求检测中间件
// 与官方 sensitive-words 完全解耦：只读取 aisec_configs 开关，不依赖 setting.CheckSensitiveEnabled
func CheckRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !model.IsEnabled() {
			c.Next()
			return
		}

		if !isChatCompletionEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		userID := c.GetInt("id")
		if userID == 0 {
			c.Next()
			return
		}

		bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize+1))
		if err != nil {
			common.SysLog("[ai-security] 读取请求体失败: " + err.Error())
			c.Next()
			return
		}
		if len(bodyBytes) > maxBodySize {
			common.SysLog("[ai-security] 请求体超过 10MB，跳过检测")
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		content := extractContentFromRequest(bodyBytes)
		if content == "" {
			c.Next()
			return
		}

		modelName := extractModelFromRequest(bodyBytes)
		detector := service.GetDetector()
		ctx := engine.Context{
			UserID:    userID,
			TokenID:   c.GetInt("token_id"),
			ChannelID: c.GetInt("channel_id"),
			RequestID: c.GetString(common.RequestIdKey),
			ModelName: modelName,
			Direction: constant.DirectionRequest,
			IP:        c.ClientIP(),
		}
		result := detector.Detect(content, ctx)

		if !result.Detected {
			c.Next()
			return
		}

		switch result.Action {
		case constant.ActionBlock:
			service.LogRequestHit(ctx, result, content, "")
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "请求包含敏感内容，已被拦截。",
			})
			c.Abort()
			return
		case constant.ActionMask:
			masked := service.MaskWithDefault(content, result)
			newBody := replaceContentInRequest(bodyBytes, content, masked)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
			c.Request.ContentLength = int64(len(newBody))
			if bs, err := common.CreateBodyStorage(newBody); err == nil {
				c.Set(common.KeyBodyStorage, bs)
			} else {
				common.SysLog("[ai-security] BodyStorage 同步失败: " + err.Error())
			}
			service.LogRequestHit(ctx, result, content, masked)
		case constant.ActionAlert, constant.ActionReview:
			service.LogRequestHit(ctx, result, content, "")
		}

		c.Next()
	}
}

func isChatCompletionEndpoint(path string) bool {
	return strings.HasSuffix(path, "/chat/completions") ||
		strings.HasSuffix(path, "/completions") ||
		strings.HasSuffix(path, "/messages")
}

func replaceContentInRequest(body []byte, oldContent, newContent string) []byte {
	return []byte(strings.Replace(string(body), oldContent, newContent, -1))
}

func extractContentFromRequest(body []byte) string {
	var req struct {
		System   string `json:"system"`
		Messages []struct {
			Role    string      `json:"role"`
			Content interface{} `json:"content"`
		} `json:"messages"`
	}

	if err := common.Unmarshal(body, &req); err != nil {
		return ""
	}

	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			text := extractTextFromContent(req.Messages[i].Content)
			if text != "" {
				return text
			}
		}
	}

	if req.System != "" {
		return req.System
	}

	return ""
}

func extractTextFromContent(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, item := range v {
			block, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if block["type"] != "text" {
				continue
			}
			text, ok := block["text"].(string)
			if !ok || text == "" {
				continue
			}
			texts = append(texts, text)
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

func extractModelFromRequest(body []byte) string {
	var req struct {
		Model string `json:"model"`
	}
	if err := common.Unmarshal(body, &req); err != nil {
		return ""
	}
	return req.Model
}

// unused 占位，避免 strconv 在裁剪时报未引用
var _ = strconv.Itoa
