package service

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/engine"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// HitLogger 命中日志记录器
type HitLogger struct {
	buffer []*model.HitLog
	mu     sync.Mutex
	ticker *time.Ticker
	done   chan struct{}
}

var (
	hitLogger *HitLogger
	once      sync.Once
)

// GetHitLogger 获取单例
func GetHitLogger() *HitLogger {
	once.Do(func() {
		hitLogger = &HitLogger{
			buffer: make([]*model.HitLog, 0, 100),
			ticker: time.NewTicker(5 * time.Second),
			done:   make(chan struct{}),
		}
		go hitLogger.loop()
	})
	return hitLogger
}

// Log 记录命中日志
func (h *HitLogger) Log(ctx engine.Context, direction int, result *dto.DetectionResult, original string, processed string) {
	if !model.IsEnabled() {
		return
	}

	hash := sha256.Sum256([]byte(original))
	hashStr := fmt.Sprintf("%x", hash)

	var ruleID, groupID int64
	if len(result.Matches) > 0 {
		ruleID = result.Matches[0].RuleID
		groupID = result.Matches[0].GroupID
	}

	var matchedText string
	if len(result.Matches) > 0 {
		matchedText = result.Matches[0].MatchedText
		if len(matchedText) > 500 {
			matchedText = matchedText[:500]
		}
	}

	log := &model.HitLog{
		RequestID:           ctx.RequestID,
		UserID:              ctx.UserID,
		TokenID:             ctx.TokenID,
		ModelName:           ctx.ModelName,
		ChannelID:           ctx.ChannelID,
		Direction:           direction,
		RuleID:              ruleID,
		GroupID:             groupID,
		RiskLevel:           result.RiskLevel,
		RiskScore:           result.RiskScore,
		Action:              result.Action,
		MatchedText:         matchedText,
		HitReason:           result.ActionName,
		OriginalContentHash: hashStr,
		ProcessedContent:    processed,
		IP:                  ctx.IP,
		CreatedAt:           time.Now().Unix(),
	}

	h.mu.Lock()
	h.buffer = append(h.buffer, log)
	shouldFlush := len(h.buffer) >= 100
	h.mu.Unlock()

	if shouldFlush {
		h.Flush()
	}
}

// Flush 批量写入日志
func (h *HitLogger) Flush() {
	h.mu.Lock()
	if len(h.buffer) == 0 {
		h.mu.Unlock()
		return
	}
	logs := make([]*model.HitLog, len(h.buffer))
	copy(logs, h.buffer)
	h.buffer = h.buffer[:0]
	h.mu.Unlock()

	if err := model.CreateHitLogs(logs); err != nil {
		common.SysError("[ai-security] flush hit logs failed: " + err.Error())
	}
}

// Stop 停止日志记录器
func (h *HitLogger) Stop() {
	close(h.done)
	h.ticker.Stop()
	h.Flush()
}

func (h *HitLogger) loop() {
	for {
		select {
		case <-h.ticker.C:
			h.Flush()
		case <-h.done:
			return
		}
	}
}

// LogRequestHit 便捷方法：记录请求命中
func LogRequestHit(ctx engine.Context, result *dto.DetectionResult, original string, processed string) {
	GetHitLogger().Log(ctx, constant.DirectionRequest, result, original, processed)
}

// LogResponseHit 便捷方法：记录响应命中
func LogResponseHit(ctx engine.Context, result *dto.DetectionResult, original string, processed string) {
	GetHitLogger().Log(ctx, constant.DirectionResponse, result, original, processed)
}
