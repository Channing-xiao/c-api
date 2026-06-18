package middleware

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/service/security"
	"github.com/gin-gonic/gin"
)

// securityStreamingWriter 拦截 SSE 流式响应，逐行进行内容安全检测
type securityStreamingWriter struct {
	gin.ResponseWriter
	c          *gin.Context
	userID     int
	modelName  string
	logCtx     security.DetectionLogContext
	detectFunc func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error)

	mu      sync.Mutex
	scanBuf []byte
	blocked bool
}

// Write 按行缓冲 SSE 数据，处理完整 data: 行时触发检测
func (w *securityStreamingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.blocked {
		return len(p), nil
	}

	w.scanBuf = append(w.scanBuf, p...)

	for {
		idx := bytes.IndexByte(w.scanBuf, '\n')
		if idx < 0 {
			break
		}
		line := w.scanBuf[:idx+1]
		w.scanBuf = w.scanBuf[idx+1:]
		if err := w.processLine(line); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

// Flush 实现 http.Flusher，保证 SSE 实时下传
func (w *securityStreamingWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// processLine 处理单行 SSE；非 data: 行直接透传
func (w *securityStreamingWriter) processLine(line []byte) error {
	trimmed := bytes.TrimSpace(line)

	// 空行、注释行、event 行直接透传
	if len(trimmed) == 0 || trimmed[0] == ':' || !bytes.HasPrefix(trimmed, []byte("data:")) {
		_, err := w.ResponseWriter.Write(line)
		return err
	}

	data := bytes.TrimSpace(trimmed[5:])
	if bytes.Equal(data, []byte("[DONE]")) {
		_, err := w.ResponseWriter.Write(line)
		return err
	}

	content, reasoning := extractStreamContent(data)
	if content == "" && reasoning == "" {
		_, err := w.ResponseWriter.Write(line)
		return err
	}

	detect := w.detectFunc
	if detect == nil {
		detect = security.GetDetectionEngine().DetectStreaming
	}

	// 分别检测 content 和 reasoning_content，各自应用动作
	var (
		newContent        = content
		newReasoning      = reasoning
		modified          bool
		overallAction     = constant.SecurityActionPass
		overallContentLen int
		overallMatches    int
		hit               bool
	)

	check := func(text string) (string, bool) {
		if text == "" {
			return text, false
		}
		result, err := detect(
			context.Background(),
			w.userID,
			text,
			constant.SecurityContentTypeResponse,
			w.modelName,
			w.logCtx,
		)
		if err != nil {
			common.SysLog("流式响应安全检测错误: " + err.Error())
			return text, false
		}
		if !result.Detected {
			return text, false
		}
		if result.Action > overallAction {
			overallAction = result.Action
		}
		overallContentLen += len(text)
		overallMatches += len(result.Matches)
		if result.Action == constant.SecurityActionMask {
			return result.ProcessedContent, true
		}
		return text, true
	}

	if content != "" {
		newContent, hit = check(content)
		if hit {
			modified = true
		}
	}
	if reasoning != "" {
		newReasoning, hit = check(reasoning)
		if hit {
			modified = true
		}
	}

	if modified {
		common.SysLog(fmt.Sprintf("[security:stream] user=%d detected=true action=%d contentLen=%d matches=%d",
			w.userID, overallAction, overallContentLen, overallMatches))
	}

	switch overallAction {
	case constant.SecurityActionPass, constant.SecurityActionAlert:
		_, err := w.ResponseWriter.Write(line)
		return err
	case constant.SecurityActionMask:
		newData, ok := replaceStreamContent(data, newContent, newReasoning)
		if !ok {
			common.SysLog("流式响应 Mask 替换未生效")
			_, err := w.ResponseWriter.Write(line)
			return err
		}
		_, err := fmt.Fprintf(w.ResponseWriter, "data: %s\n", string(newData))
		return err
	case constant.SecurityActionBlock:
		w.blocked = true
		return w.writeBlockFrame()
	default:
		_, err := w.ResponseWriter.Write(line)
		return err
	}
}

// writeBlockFrame 向客户端发送流式终止帧
func (w *securityStreamingWriter) writeBlockFrame() error {
	// Claude /messages 接口返回 Claude 格式终止帧
	if strings.HasSuffix(w.c.Request.URL.Path, "/messages") {
		_, err := w.ResponseWriter.Write([]byte("event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n"))
		return err
	}
	// 默认 OpenAI 格式
	_, err := w.ResponseWriter.Write([]byte("data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"content_filter\"}]}\n\ndata: [DONE]\n\n"))
	return err
}

// extractStreamContent 从 SSE data 负载中提取文本内容
// 支持 OpenAI chat.completions delta.content / delta.reasoning_content 与 Claude delta.text
func extractStreamContent(data []byte) (content string, reasoning string) {
	var payload struct {
		Choices []struct {
			Delta struct {
				Content          string `json:"content"`
				ReasoningContent string `json:"reasoning_content"`
			} `json:"delta"`
		} `json:"choices"`
		Delta struct {
			Text string `json:"text"`
		} `json:"delta"`
	}
	if err := common.Unmarshal(data, &payload); err != nil {
		return "", ""
	}
	if len(payload.Choices) > 0 {
		return payload.Choices[0].Delta.Content, payload.Choices[0].Delta.ReasoningContent
	}
	return payload.Delta.Text, ""
}

// replaceStreamContent 将 SSE data 负载中的文本替换为脱敏后内容
func replaceStreamContent(data []byte, content, reasoning string) ([]byte, bool) {
	var payload map[string]interface{}
	if err := common.Unmarshal(data, &payload); err != nil {
		return nil, false
	}

	// OpenAI 格式：choices[0].delta.content / reasoning_content
	if choicesRaw, ok := payload["choices"].([]interface{}); ok && len(choicesRaw) > 0 {
		if choice, ok := choicesRaw[0].(map[string]interface{}); ok {
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if content != "" {
					if _, ok := delta["content"]; ok {
						delta["content"] = content
					}
				}
				if reasoning != "" {
					if _, ok := delta["reasoning_content"]; ok {
						delta["reasoning_content"] = reasoning
					}
				}
				newData, err := common.Marshal(payload)
				return newData, err == nil
			}
		}
	}

	// Claude 格式：delta.text
	if delta, ok := payload["delta"].(map[string]interface{}); ok {
		if _, ok := delta["text"]; ok {
			delta["text"] = content
			newData, err := common.Marshal(payload)
			return newData, err == nil
		}
	}

	return nil, false
}
