package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/service/security"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
)

// SecurityCheck 请求内容安全检测中间件
func SecurityCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 统一开关：环境变量 SECURITY_ENABLED 或旧系统 CheckSensitiveEnabled 任一关闭即跳过
		if !security.IsSecurityEnabled() || !setting.CheckSensitiveEnabled {
			c.Next()
			return
		}

		// 只对聊天补全接口进行检测
		if !isChatCompletionEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 获取当前用户
		userId := c.GetInt("id")
		if userId == 0 {
			c.Next()
			return
		}

		// 读取请求体（限制最大 10MB）
		const maxBodySize = 10 * 1024 * 1024
		bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize+1))
		if err != nil {
			common.SysLog("读取请求体失败: " + err.Error())
			c.Next()
			return
		}
		if len(bodyBytes) > maxBodySize {
			common.SysLog("请求体超过 10MB，跳过安全检测")
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 解析请求内容
		content := extractContentFromRequest(bodyBytes)
		if content == "" {
			c.Next()
			return
		}

		modelName := extractModelFromRequest(bodyBytes)

		// 执行检测
		ctx := context.Background()
		result, err := security.GetDetectionEngine().Detect(ctx, userId, content, constant.SecurityContentTypeRequest, modelName)
		if err != nil {
			common.SysLog("安全检测错误: " + err.Error())
			c.Next()
			return
		}

		if result.Detected {
			common.SysLog(fmt.Sprintf("[security:middleware] user=%d detected=%v action=%d contentLen=%d",
				userId, result.Detected, result.Action, len(content)))
			switch result.Action {
			case constant.SecurityActionBlock:
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"message": getBlockMessage(userId),
				})
				c.Abort()
				return
			case constant.SecurityActionMask:
				// 替换请求体中的敏感内容（支持多条消息按消息粒度脱敏）
				newBody, replaced := replaceMaskedRequest(bodyBytes, result)
				if !replaced {
					common.SysLog("请求体 Mask 替换未生效，原始内容长度: " + strconv.Itoa(len(bodyBytes)))
				}
				c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
				c.Request.ContentLength = int64(len(newBody))
			}
		}

		c.Next()
	}
}

// SecurityCheckResponse 响应内容安全检测中间件
func SecurityCheckResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 统一开关：环境变量 SECURITY_ENABLED 或旧系统 CheckSensitiveEnabled 任一关闭即跳过
		if !security.IsSecurityEnabled() || !setting.CheckSensitiveEnabled {
			c.Next()
			return
		}

		// 只对聊天补全接口进行检测
		if !isChatCompletionEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		userId := c.GetInt("id")
		if userId == 0 {
			c.Next()
			return
		}

		// 使用自定义 ResponseWriter 拦截响应
		blw := &bufferedResponseWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = blw

		c.Next()

		// 如果响应已经写入（例如流式响应），跳过
		if blw.written {
			return
		}

		// 如果响应已经是错误状态码，直接返回原始响应
		if blw.statusCode >= 400 {
			blw.flushOriginal()
			return
		}

		// 只处理 JSON 响应
		contentType := blw.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			// 非 JSON 直接写回原始响应
			blw.flushOriginal()
			return
		}

		// 从响应体中提取 AI 生成内容
		body := blw.body.Bytes()
		content := extractContentFromResponse(body)
		if content == "" {
			blw.flushOriginal()
			return
		}

		// 执行检测
		ctx := context.Background()
		result, err := security.GetDetectionEngine().Detect(ctx, userId, content, constant.SecurityContentTypeResponse, "")
		if err != nil {
			common.SysLog("响应安全检测错误: " + err.Error())
			blw.flushOriginal()
			return
		}

		if result.Detected {
			common.SysLog(fmt.Sprintf("[security:middleware] user=%d detected=%v action=%d contentLen=%d",
				userId, result.Detected, result.Action, len(content)))
			switch result.Action {
			case constant.SecurityActionBlock:
				// 重写为拦截响应
				c.Writer = blw.ResponseWriter
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": getBlockMessage(userId),
				})
				return
			case constant.SecurityActionMask:
				// 替换响应中的敏感内容（支持多个 choice 按 choice 粒度脱敏）
				newBody, replaced := replaceMaskedResponse(body, result)
				if !replaced {
					common.SysLog("响应体 Mask 替换未生效，原始内容长度: " + strconv.Itoa(len(body)))
					blw.flushOriginal()
					return
				}
				blw.Header().Set("Content-Length", strconv.Itoa(len(newBody)))
				blw.ResponseWriter.WriteHeader(blw.statusCode)
				blw.ResponseWriter.Write(newBody)
				return
			}
		}

		blw.flushOriginal()
	}
}

// bufferedResponseWriter 缓冲响应内容的 ResponseWriter
type bufferedResponseWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	statusCode int
	written bool
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
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	w.ResponseWriter.WriteHeader(w.statusCode)
	w.ResponseWriter.Write(w.body.Bytes())
}

// 实现 http.Flusher 接口以支持流式响应检测降级
func (w *bufferedResponseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// extractContentFromResponse 从响应体中提取 AI 生成的内容
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

	return strings.Join(contents, "\n")
}

// replaceContentInRequest 替换请求体中的内容（保留用于简单场景兼容）
func replaceContentInRequest(body []byte, oldContent, newContent string) []byte {
	return []byte(strings.Replace(string(body), oldContent, newContent, -1))
}

// replaceMaskedRequest 将 Mask 后的内容按消息粒度写回请求体，保留所有其他字段
func replaceMaskedRequest(body []byte, result *security.DetectionResult) ([]byte, bool) {
	if !result.Detected || result.Action != constant.SecurityActionMask || len(result.Matches) == 0 {
		return body, false
	}

	var reqMap map[string]json.RawMessage
	if err := common.Unmarshal(body, &reqMap); err != nil {
		common.SysLog("脱敏请求体解析失败: " + err.Error())
		return body, false
	}

	var messages []json.RawMessage
	if err := common.Unmarshal(reqMap["messages"], &messages); err != nil {
		return body, false
	}

	type msgRef struct {
		index   int
		content string
		start   int
		end     int
	}
	var refs []msgRef
	var joined strings.Builder
	for i, raw := range messages {
		var msg struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		if err := common.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Role == "user" && msg.Content != "" {
			if joined.Len() > 0 {
				joined.WriteByte('\n')
			}
			start := joined.Len()
			joined.WriteString(msg.Content)
			refs = append(refs, msgRef{index: i, content: msg.Content, start: start, end: joined.Len()})
		}
	}

	if len(refs) == 0 {
		return body, false
	}

	// 单条消息：沿用整体字符串替换
	if len(refs) == 1 {
		newBody := []byte(strings.Replace(string(body), refs[0].content, result.ProcessedContent, -1))
		return newBody, !bytes.Equal(newBody, body)
	}

	// 多条消息：按消息偏移量拆分 matches，逐条脱敏
	modified := false
	for _, ref := range refs {
		var matches []*dto.SecurityMatchResult
		for _, m := range result.Matches {
			if m.Position[0] >= ref.start && m.Position[1] <= ref.end {
				adjusted := *m
				adjusted.Position[0] -= ref.start
				adjusted.Position[1] -= ref.start
				matches = append(matches, &adjusted)
			}
		}
		if len(matches) == 0 {
			continue
		}

		masked := security.ApplyMasking(ref.content, matches, nil)
		if masked == ref.content {
			continue
		}

		var msgMap map[string]json.RawMessage
		if err := common.Unmarshal(messages[ref.index], &msgMap); err != nil {
			continue
		}
		contentBytes, err := common.Marshal(masked)
		if err != nil {
			continue
		}
		msgMap["content"] = contentBytes
		newMsg, err := common.Marshal(msgMap)
		if err != nil {
			continue
		}
		messages[ref.index] = newMsg
		modified = true
	}

	if !modified {
		return body, false
	}

	messagesBytes, err := common.Marshal(messages)
	if err != nil {
		common.SysLog("脱敏后重新编码 messages 失败: " + err.Error())
		return body, false
	}
	reqMap["messages"] = messagesBytes
	newBody, err := common.Marshal(reqMap)
	if err != nil {
		common.SysLog("脱敏后重新编码请求体失败: " + err.Error())
		return body, false
	}
	return newBody, true
}

// replaceMaskedResponse 将 Mask 后的内容按 choice 粒度写回响应体，保留所有其他字段
func replaceMaskedResponse(body []byte, result *security.DetectionResult) ([]byte, bool) {
	if !result.Detected || result.Action != constant.SecurityActionMask || len(result.Matches) == 0 {
		return body, false
	}

	var respMap map[string]json.RawMessage
	if err := common.Unmarshal(body, &respMap); err != nil {
		common.SysLog("脱敏响应体解析失败: " + err.Error())
		return body, false
	}

	var choices []json.RawMessage
	if err := common.Unmarshal(respMap["choices"], &choices); err != nil {
		return body, false
	}

	type contentRef struct {
		index   int
		field   string
		content string
		start   int
		end     int
	}
	var refs []contentRef
	var joined strings.Builder
	for i, raw := range choices {
		var choice struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		}
		if err := common.Unmarshal(raw, &choice); err != nil {
			continue
		}
		if choice.Message.Content != "" {
			if joined.Len() > 0 {
				joined.WriteByte('\n')
			}
			start := joined.Len()
			joined.WriteString(choice.Message.Content)
			refs = append(refs, contentRef{index: i, field: "message", content: choice.Message.Content, start: start, end: joined.Len()})
		} else if choice.Delta.Content != "" {
			if joined.Len() > 0 {
				joined.WriteByte('\n')
			}
			start := joined.Len()
			joined.WriteString(choice.Delta.Content)
			refs = append(refs, contentRef{index: i, field: "delta", content: choice.Delta.Content, start: start, end: joined.Len()})
		}
	}

	if len(refs) == 0 {
		return body, false
	}

	if len(refs) == 1 {
		newBody := []byte(strings.Replace(string(body), refs[0].content, result.ProcessedContent, -1))
		return newBody, !bytes.Equal(newBody, body)
	}

	modified := false
	for _, ref := range refs {
		var matches []*dto.SecurityMatchResult
		for _, m := range result.Matches {
			if m.Position[0] >= ref.start && m.Position[1] <= ref.end {
				adjusted := *m
				adjusted.Position[0] -= ref.start
				adjusted.Position[1] -= ref.start
				matches = append(matches, &adjusted)
			}
		}
		if len(matches) == 0 {
			continue
		}

		masked := security.ApplyMasking(ref.content, matches, nil)
		if masked == ref.content {
			continue
		}

		var choiceMap map[string]json.RawMessage
		if err := common.Unmarshal(choices[ref.index], &choiceMap); err != nil {
			continue
		}
		var fieldMap map[string]json.RawMessage
		if err := common.Unmarshal(choiceMap[ref.field], &fieldMap); err != nil {
			continue
		}
		contentBytes, err := common.Marshal(masked)
		if err != nil {
			continue
		}
		fieldMap["content"] = contentBytes
		newField, err := common.Marshal(fieldMap)
		if err != nil {
			continue
		}
		choiceMap[ref.field] = newField
		newChoice, err := common.Marshal(choiceMap)
		if err != nil {
			continue
		}
		choices[ref.index] = newChoice
		modified = true
	}

	if !modified {
		return body, false
	}

	choicesBytes, err := common.Marshal(choices)
	if err != nil {
		common.SysLog("脱敏后重新编码 choices 失败: " + err.Error())
		return body, false
	}
	respMap["choices"] = choicesBytes
	newBody, err := common.Marshal(respMap)
	if err != nil {
		common.SysLog("脱敏后重新编码响应体失败: " + err.Error())
		return body, false
	}
	return newBody, true
}

// replaceContentInResponse 替换响应体中的内容（保留用于简单场景兼容）
func replaceContentInResponse(body []byte, oldContent, newContent string) []byte {
	return []byte(strings.Replace(string(body), oldContent, newContent, -1))
}

// isChatCompletionEndpoint 判断是否为聊天补全接口
func isChatCompletionEndpoint(path string) bool {
	return strings.HasSuffix(path, "/chat/completions") || strings.HasSuffix(path, "/completions")
}

// extractContentFromRequest 从请求体中提取用户内容
func extractContentFromRequest(body []byte) string {
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := common.Unmarshal(body, &req); err != nil {
		return ""
	}

	var contents []string
	for _, msg := range req.Messages {
		if msg.Role == "user" && msg.Content != "" {
			contents = append(contents, msg.Content)
		}
	}

	return strings.Join(contents, "\n")
}

// extractModelFromRequest 从请求体中提取模型名称
func extractModelFromRequest(body []byte) string {
	var req struct {
		Model string `json:"model"`
	}
	if err := common.Unmarshal(body, &req); err != nil {
		return ""
	}
	return req.Model
}

// getBlockMessage 获取拦截提示消息
func getBlockMessage(userId int) string {
	// 尝试获取用户的自定义拦截消息
	policies, err := security.GetUserPolicies(userId)
	if err != nil {
		return "请求包含敏感内容，已被拦截。"
	}

	for _, policy := range policies {
		if policy.CustomResponse != "" {
			return policy.CustomResponse
		}
	}

	return "请求包含敏感内容，已被拦截。"
}
