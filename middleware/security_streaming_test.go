package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/service/security"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestExtractStreamContent_OpenAI(t *testing.T) {
	data := []byte(`{"choices":[{"index":0,"delta":{"content":"hello world"}}]}`)
	require.Equal(t, "hello world", extractStreamContent(data))
}

func TestExtractStreamContent_Claude(t *testing.T) {
	data := []byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hello world"}}`)
	require.Equal(t, "hello world", extractStreamContent(data))
}

func TestExtractStreamContent_Empty(t *testing.T) {
	data := []byte(`{"choices":[{"index":0,"delta":{}}]}`)
	require.Equal(t, "", extractStreamContent(data))
}

func TestReplaceStreamContent_OpenAI(t *testing.T) {
	data := []byte(`{"choices":[{"index":0,"delta":{"content":"请联系 13800138000"}}]}`)
	newData, ok := replaceStreamContent(data, "请联系 ***")
	require.True(t, ok)
	require.Contains(t, string(newData), `"content":"请联系 ***"`)
	require.NotContains(t, string(newData), "13800138000")
}

func TestReplaceStreamContent_Claude(t *testing.T) {
	data := []byte(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"请联系 13800138000"}}`)
	newData, ok := replaceStreamContent(data, "请联系 ***")
	require.True(t, ok)
	require.Contains(t, string(newData), `"text":"请联系 ***"`)
	require.NotContains(t, string(newData), "13800138000")
}

func TestSecurityStreamingWriter_PassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	writer := &securityStreamingWriter{
		ResponseWriter: c.Writer,
		c:              c,
		userID:         1,
		detectFunc: func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error) {
			return &security.DetectionResult{Detected: false, Action: constant.SecurityActionPass}, nil
		},
	}

	input := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hello\"}}]}\n\n"
	n, err := writer.Write([]byte(input))
	require.NoError(t, err)
	require.Equal(t, len(input), n)

	body := recorder.Body.String()
	require.Contains(t, body, `"content":"hello"`)
}

func TestSecurityStreamingWriter_PartialLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	writer := &securityStreamingWriter{
		ResponseWriter: c.Writer,
		c:              c,
		userID:         1,
		detectFunc: func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error) {
			return &security.DetectionResult{Detected: false, Action: constant.SecurityActionPass}, nil
		},
	}

	_, err := writer.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hel"))
	require.NoError(t, err)
	require.Empty(t, recorder.Body.String())

	_, err = writer.Write([]byte("lo\"}}]}\n\n"))
	require.NoError(t, err)
	require.Contains(t, recorder.Body.String(), `"content":"hello"`)
}

func TestSecurityStreamingWriter_MaskChunk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	writer := &securityStreamingWriter{
		ResponseWriter: c.Writer,
		c:              c,
		userID:         1,
		detectFunc: func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error) {
			return &security.DetectionResult{
				Detected:         true,
				Action:           constant.SecurityActionMask,
				ProcessedContent: "请联系 ***",
			}, nil
		},
	}

	input := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"请联系 13800138000\"}}]}\n\n"
	_, err := writer.Write([]byte(input))
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, `"content":"请联系 ***"`)
	require.NotContains(t, body, "13800138000")
}

func TestSecurityStreamingWriter_BlockChunk(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest(http.MethodPost, "/v1/chat/completions", nil)

	writer := &securityStreamingWriter{
		ResponseWriter: c.Writer,
		c:              c,
		userID:         1,
		detectFunc: func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error) {
			return &security.DetectionResult{
				Detected: true,
				Action:   constant.SecurityActionBlock,
			}, nil
		},
	}

	input := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"敏感内容\"}}]}\n\n"
	_, err := writer.Write([]byte(input))
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, `"finish_reason":"content_filter"`)
	require.Contains(t, body, "data: [DONE]")

	// 阻断后后续写入应被丢弃
	recorder.Body.Reset()
	_, err = writer.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"more\"}}]}\n\n"))
	require.NoError(t, err)
	require.Empty(t, recorder.Body.String())
}

func TestSecurityStreamingWriter_ClaudeFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request, _ = http.NewRequest(http.MethodPost, "/v1/messages", nil)

	writer := &securityStreamingWriter{
		ResponseWriter: c.Writer,
		c:              c,
		userID:         1,
		detectFunc: func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error) {
			return &security.DetectionResult{
				Detected: true,
				Action:   constant.SecurityActionBlock,
			}, nil
		},
	}

	input := "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"敏感\"}}\n\n"
	_, err := writer.Write([]byte(input))
	require.NoError(t, err)

	body := recorder.Body.String()
	require.True(t, strings.Contains(body, "event: message_stop") || strings.Contains(body, "message_stop"), "expected Claude stop frame, got: %s", body)
}

func TestSecurityStreamingWriter_EventLinePassThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	writer := &securityStreamingWriter{
		ResponseWriter: c.Writer,
		c:              c,
		userID:         1,
		detectFunc: func(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx security.DetectionLogContext) (*security.DetectionResult, error) {
			return &security.DetectionResult{Detected: false, Action: constant.SecurityActionPass}, nil
		},
	}

	input := "event: content_block_start\ndata: {\"type\":\"content_block_start\"}\n\n"
	_, err := writer.Write([]byte(input))
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, "event: content_block_start")
	require.Contains(t, body, "data: {\"type\":\"content_block_start\"}")
}
