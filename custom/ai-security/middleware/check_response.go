package middleware

import (
	"bytes"
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

// CheckResponse ai-security 响应检测中间件
// 缓冲非流式 JSON 响应进行检测；流式响应直接放行（避免破坏实时性）
func CheckResponse() gin.HandlerFunc {
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

		blw := &bufferedResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = blw

		c.Next()

		if blw.written {
			return
		}
		if blw.statusCode >= 400 {
			blw.flushOriginal()
			return
		}

		contentType := blw.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			blw.flushOriginal()
			return
		}

		body := blw.body.Bytes()
		content := extractContentFromResponse(body)
		if content == "" {
			blw.flushOriginal()
			return
		}

		modelName := c.GetString("original_model")
		ctx := engine.Context{
			UserID:    userID,
			TokenID:   c.GetInt("token_id"),
			ChannelID: c.GetInt("channel_id"),
			RequestID: c.GetString(common.RequestIdKey),
			ModelName: modelName,
			Direction: constant.DirectionResponse,
			IP:        c.ClientIP(),
		}
		result := service.GetDetector().Detect(content, ctx)
		if !result.Detected {
			blw.flushOriginal()
			return
		}

		switch result.Action {
		case constant.ActionBlock:
			service.LogResponseHit(ctx, result, content, "")
			c.Writer = blw.ResponseWriter
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "响应包含敏感内容，已被拦截。",
			})
			return
		case constant.ActionMask:
			masked := service.MaskWithDefault(content, result)
			newBody := []byte(strings.Replace(string(body), content, masked, -1))
			service.LogResponseHit(ctx, result, content, masked)
			blw.Header().Set("Content-Length", strconv.Itoa(len(newBody)))
			blw.ResponseWriter.WriteHeader(statusOr200(blw.statusCode))
			_, _ = blw.ResponseWriter.Write(newBody)
			return
		case constant.ActionAlert, constant.ActionReview:
			service.LogResponseHit(ctx, result, content, "")
		}

		blw.flushOriginal()
	}
}

// bufferedResponseWriter 缓冲响应内容
type bufferedResponseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

func (w *bufferedResponseWriter) WriteHeader(code int) {
	w.statusCode = code
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *bufferedResponseWriter) flushOriginal() {
	if w.written {
		return
	}
	w.written = true
	w.ResponseWriter.WriteHeader(statusOr200(w.statusCode))
	_, _ = w.ResponseWriter.Write(w.body.Bytes())
}

func (w *bufferedResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func statusOr200(code int) int {
	if code == 0 {
		return http.StatusOK
	}
	return code
}

// extractContentFromResponse 从响应体提取 AI 生成内容（OpenAI / Claude）
func extractContentFromResponse(body []byte) string {
	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := common.Unmarshal(body, &resp); err != nil {
		return ""
	}

	var contents []string
	for _, choice := range resp.Choices {
		if choice.Message.Content != "" {
			contents = append(contents, choice.Message.Content)
		} else if choice.Delta.Content != "" {
			contents = append(contents, choice.Delta.Content)
		}
	}
	for _, block := range resp.Content {
		if block.Type == "text" && block.Text != "" {
			contents = append(contents, block.Text)
		}
	}

	return strings.Join(contents, "\n")
}
