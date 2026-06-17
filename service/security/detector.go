package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"golang.org/x/sync/errgroup"
)

// DetectionLogContext 命中日志上下文信息
// 用于把请求链路中的 request_id、channel_id、token_id 写入安全日志，与网关请求日志关联
type DetectionLogContext struct {
	RequestID string
	ChannelID int
	TokenID   int
}

// MaskConfig 脱敏配置（通过规则 extra_config JSON 字段配置）
type MaskConfig struct {
	MaskType      string `json:"mask_type"`      // full: 整体替换（默认）; preserve: 保留前后
	PreserveStart int    `json:"preserve_start"` // 保留前 N 位
	PreserveEnd   int    `json:"preserve_end"`   // 保留后 N 位
	MaskChar      string `json:"mask_char"`      // 替换字符，默认 *
	ReplaceWith   string `json:"replace_with"`   // 整体替换时的固定字符串（优先级高于 MaskChar）
}

// DetectionResult 检测结果
type DetectionResult struct {
	Detected         bool
	Action           int
	RiskScore        int
	RiskLevel        int
	ProcessedContent string
	Matches          []*dto.SecurityMatchResult
	EngineResults    map[string]*EngineResult
}

// EngineResult 单个引擎检测结果
type EngineResult struct {
	EngineName string
	Detected   bool
	Matches    []*dto.SecurityMatchResult
	RiskScore  int
	Error      error
}

// ContentDetector 内容检测器接口
type ContentDetector interface {
	Name() string
	Detect(content string, rules []*model.SecurityRule) (*EngineResult, error)
}

// DetectionEngine 检测引擎
type DetectionEngine struct {
	detectors []ContentDetector
}

var (
	detectionEngine     *DetectionEngine
	detectionEngineOnce sync.Once
)

// GetDetectionEngine 获取检测引擎单例
func GetDetectionEngine() *DetectionEngine {
	detectionEngineOnce.Do(func() {
		detectionEngine = &DetectionEngine{
			detectors: []ContentDetector{
				&KeywordDetector{},
				&RegexDetector{},
				&NERDetector{},
				&AIDetector{},
			},
		}
	})
	return detectionEngine
}

// Detect 执行内容检测
func (de *DetectionEngine) Detect(ctx context.Context, userID int, content string, contentType int, modelName string, logCtx DetectionLogContext) (*DetectionResult, error) {
	result := &DetectionResult{
		Detected:      false,
		Action:        constant.SecurityActionPass,
		RiskScore:     0,
		EngineResults: make(map[string]*EngineResult),
		Matches:       make([]*dto.SecurityMatchResult, 0),
	}

	if !IsSecurityEnabled() {
		return result, nil
	}

	// 获取用户策略（使用缓存）
	policies, err := GetCachedUserPolicies(userID)
	if err != nil {
		common.SysLog("获取用户安全策略失败: " + err.Error())
		return result, nil
	}

	if len(policies) == 0 {
		common.SysLog(fmt.Sprintf("[security] user=%d no policies", userID))
		return result, nil
	}

	// 过滤生效范围并去重
	var effectiveGroupIds []int64
	seen := make(map[int64]bool)
	for _, policy := range policies {
		if contentType == constant.SecurityContentTypeRequest && policy.Scope == constant.SecurityScopeResponseOnly {
			continue
		}
		if contentType == constant.SecurityContentTypeResponse && policy.Scope == constant.SecurityScopeRequestOnly {
			continue
		}
		if !seen[policy.GroupID] {
			seen[policy.GroupID] = true
			effectiveGroupIds = append(effectiveGroupIds, policy.GroupID)
		}
	}

	if len(effectiveGroupIds) == 0 {
		common.SysLog(fmt.Sprintf("[security] user=%d no effective groups for contentType=%d", userID, contentType))
		return result, nil
	}

	// 获取规则（优先从缓存）
	rules, err := GetCachedRulesByGroupIds(effectiveGroupIds)
	if err != nil {
		common.SysLog("获取安全规则失败: " + err.Error())
		return result, nil
	}

	if len(rules) == 0 {
		common.SysLog(fmt.Sprintf("[security] user=%d no rules for groups=%v", userID, effectiveGroupIds))
		return result, nil
	}

	// 并行执行本地检测引擎
	var mu sync.Mutex
	var detected bool
	var maxRiskScore int
	var allMatches []*dto.SecurityMatchResult

	g, ctx := errgroup.WithContext(ctx)

	// Keyword 和 Regex 并行执行
	for _, detector := range de.detectors[:2] {
		d := detector
		g.Go(func() error {
			engineResult, err := d.Detect(content, rules)
			if err != nil {
				common.SysLog(fmt.Sprintf("检测引擎 %s 错误: %v", d.Name(), err))
				return nil
			}
			mu.Lock()
			defer mu.Unlock()
			result.EngineResults[d.Name()] = engineResult
			if engineResult.Detected {
				detected = true
				if engineResult.RiskScore > maxRiskScore {
					maxRiskScore = engineResult.RiskScore
				}
				allMatches = append(allMatches, engineResult.Matches...)
			}
			return nil
		})
	}

	_ = g.Wait()

	// NER 检测
	if nerDetector, ok := de.detectors[2].(*NERDetector); ok {
		engineResult, err := nerDetector.Detect(content, rules)
		if err == nil && engineResult.Detected {
			result.EngineResults[nerDetector.Name()] = engineResult
			detected = true
			if engineResult.RiskScore > maxRiskScore {
				maxRiskScore = engineResult.RiskScore
			}
			allMatches = append(allMatches, engineResult.Matches...)
		}
	}

	// AI 检测（异步，带超时）
	aiCtx, cancel := context.WithTimeout(ctx, time.Duration(constant.SecurityAITimeoutSeconds)*time.Second)
	defer cancel()

	if aiDetector, ok := de.detectors[3].(*AIDetector); ok {
		aiResult, err := aiDetector.DetectWithContext(aiCtx, content, rules)
		if err == nil && aiResult.Detected {
			result.EngineResults[aiDetector.Name()] = aiResult
			detected = true
			if aiResult.RiskScore > maxRiskScore {
				maxRiskScore = aiResult.RiskScore
			}
			allMatches = append(allMatches, aiResult.Matches...)
		} else if err != nil {
			common.SysLog("AI 检测超时或失败，降级到本地规则: " + err.Error())
		}
	}

	result.Detected = detected
	result.RiskScore = maxRiskScore
	result.RiskLevel = constant.GetSecurityRiskLevelByScore(maxRiskScore)

	// 过滤掉无效匹配（空文本或非法位置），避免误报
	validMatches := make([]*dto.SecurityMatchResult, 0, len(allMatches))
	for _, m := range allMatches {
		if m == nil || m.MatchedText == "" {
			continue
		}
		if m.Position[0] < 0 || m.Position[1] <= m.Position[0] || m.Position[1] > len(content) {
			common.SysLog(fmt.Sprintf("[security] invalid match position dropped: ruleID=%d position=%v", m.RuleID, m.Position))
			continue
		}
		validMatches = append(validMatches, m)
	}
	result.Matches = validMatches

	// 重新计算 Detected：没有有效匹配时不得判定为命中
	if len(result.Matches) == 0 {
		result.Detected = false
		result.Action = constant.SecurityActionPass
		result.RiskScore = 0
		result.RiskLevel = constant.GetSecurityRiskLevelByScore(0)
	}

	if result.Detected {
		// 计算最终动作（取最高优先级）
		result.Action = resolveAction(result.Matches, rules)
		// 执行脱敏
		if result.Action == constant.SecurityActionMask {
			result.ProcessedContent = applyMasking(content, result.Matches, rules)
		}
	}

	// 结构化调试日志（不记录敏感内容）
	var matchRuleIDs []int64
	for _, m := range result.Matches {
		matchRuleIDs = append(matchRuleIDs, m.RuleID)
	}
	common.SysLog(fmt.Sprintf("[security] detect user=%d contentLen=%d contentType=%d policies=%d groups=%d rules=%d matches=%d matchRuleIDs=%v action=%d riskScore=%d",
		userID, len(content), contentType, len(policies), len(effectiveGroupIds), len(rules), len(result.Matches), matchRuleIDs, result.Action, result.RiskScore))

	// 详细命中诊断日志（记录命中的规则和文本，便于排查误报）
	for _, m := range result.Matches {
		common.SysLog(fmt.Sprintf("[security:match] ruleID=%d groupID=%d type=%d matchedText=%q position=%v",
			m.RuleID, m.GroupID, m.Type, m.MatchedText, m.Position))
	}

	// 异步记录日志
	go recordHitLog(userID, content, result, contentType, modelName, logCtx)

	return result, nil
}

// DetectWithRule 使用单条规则检测内容（用于规则测试）
func (de *DetectionEngine) DetectWithRule(rule *model.SecurityRule, content string) (*DetectionResult, error) {
	result := &DetectionResult{
		Detected:      false,
		Action:        constant.SecurityActionPass,
		RiskScore:     0,
		EngineResults: make(map[string]*EngineResult),
		Matches:       make([]*dto.SecurityMatchResult, 0),
	}

	if rule == nil || rule.Status != constant.SecurityStatusEnabled {
		return result, nil
	}

	// 根据规则类型选择检测器
	var detector ContentDetector
	switch rule.Type {
	case constant.SecurityRuleTypeKeyword:
		detector = de.detectors[0] // KeywordDetector
	case constant.SecurityRuleTypeRegex:
		detector = de.detectors[1] // RegexDetector
	case constant.SecurityRuleTypeNER:
		detector = de.detectors[2] // NERDetector
	case constant.SecurityRuleTypeAI:
		detector = de.detectors[3] // AIDetector
	default:
		return result, nil
	}

	engineResult, err := detector.Detect(content, []*model.SecurityRule{rule})
	if err != nil {
		return result, err
	}

	result.EngineResults[detector.Name()] = engineResult
	if engineResult.Detected {
		result.Detected = true
		result.RiskScore = engineResult.RiskScore
		result.RiskLevel = constant.GetSecurityRiskLevelByScore(engineResult.RiskScore)
		result.Matches = engineResult.Matches
		result.Action = rule.Action
		if result.Action == constant.SecurityActionMask {
			result.ProcessedContent = applyMasking(content, result.Matches, []*model.SecurityRule{rule})
		}
	}

	return result, nil
}

// resolveAction 根据匹配结果解析最终动作（取最高优先级）
func resolveAction(matches []*dto.SecurityMatchResult, rules []*model.SecurityRule) int {
	if len(matches) == 0 {
		return constant.SecurityActionPass
	}

	// 构建规则 ID -> 动作的映射
	ruleActionMap := make(map[int64]int)
	for _, rule := range rules {
		ruleActionMap[rule.ID] = rule.Action
	}

	maxPriority := constant.SecurityActionPriorityPass
	finalAction := constant.SecurityActionPass

	for _, match := range matches {
		if action, ok := ruleActionMap[match.RuleID]; ok {
			priority := constant.GetSecurityActionPriority(action)
			if priority > maxPriority {
				maxPriority = priority
				finalAction = action
			}
		}
	}

	return finalAction
}

// ApplyMasking 对外暴露的脱敏处理入口
func ApplyMasking(content string, matches []*dto.SecurityMatchResult, rules []*model.SecurityRule) string {
	return applyMasking(content, matches, rules)
}

// applyMasking 应用脱敏处理
func applyMasking(content string, matches []*dto.SecurityMatchResult, rules []*model.SecurityRule) string {
	if len(matches) == 0 {
		return content
	}

	// 构建规则 ID -> 脱敏配置的映射
	ruleConfigMap := make(map[int64]*MaskConfig)
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		config := parseMaskConfig(rule.ExtraConfig)
		ruleConfigMap[rule.ID] = config
	}

	// 按位置降序排序（从后往前替换，避免位置偏移）
	sortedMatches := make([]*dto.SecurityMatchResult, len(matches))
	copy(sortedMatches, matches)
	sort.Slice(sortedMatches, func(i, j int) bool {
		return sortedMatches[i].Position[0] > sortedMatches[j].Position[0]
	})

	result := content
	for _, match := range sortedMatches {
		if match.Position[1] <= match.Position[0] {
			continue
		}
		if match.Position[0] < 0 || match.Position[1] > len(result) {
			continue
		}
		config := ruleConfigMap[match.RuleID]
		matchedText := result[match.Position[0]:match.Position[1]]
		masked := maskText(matchedText, config)
		result = result[:match.Position[0]] + masked + result[match.Position[1]:]
	}

	return result
}

// parseMaskConfig 解析规则的 extra_config 脱敏配置
func parseMaskConfig(extraConfig string) *MaskConfig {
	if extraConfig == "" {
		return nil
	}
	var config MaskConfig
	if err := common.UnmarshalJsonStr(extraConfig, &config); err != nil {
		common.SysLog(fmt.Sprintf("[security:mask] 解析 extra_config 失败: %v", err))
		return nil
	}
	// 默认替换字符
	if config.MaskChar == "" {
		config.MaskChar = "*"
	}
	return &config
}

// maskText 对文本进行脱敏处理
// 默认整体替换为 ***；若配置 mask_type=preserve 则保留前后指定位数
func maskText(text string, config *MaskConfig) string {
	if config == nil || config.MaskType == "" || config.MaskType == "full" {
		if config != nil && config.ReplaceWith != "" {
			return config.ReplaceWith
		}
		return "***"
	}

	if config.MaskType == "preserve" {
		runes := []rune(text)
		length := len(runes)
		start := config.PreserveStart
		end := config.PreserveEnd
		if start < 0 {
			start = 0
		}
		if end < 0 {
			end = 0
		}
		// 保留位数超过或等于文本长度时，直接整体替换
		if start+end >= length {
			return config.MaskChar
		}
		maskLen := length - start - end
		return string(runes[:start]) + strings.Repeat(config.MaskChar, maskLen) + string(runes[length-end:])
	}

	// 兜底：未知 mask_type 按 full 处理
	return "***"
}

const (
	hitLogChanSize      = 1000            // 命中日志缓冲通道大小
	hitLogBatchSize     = 50              // 批量写入阈值
	hitLogFlushInterval = 2 * time.Second // 批量写入最大等待时间
	hitLogMaxRetries    = 3               // 批量写入失败最大重试次数
)

var (
	hitLogChan       chan *model.SecurityHitLog
	hitLogWorkerOnce sync.Once
	hitLogWg         sync.WaitGroup
	droppedHitLogs   atomic.Uint64
)

// initHitLogWorker 启动命中日志后台 worker（只执行一次）
func initHitLogWorker() {
	hitLogWorkerOnce.Do(func() {
		hitLogChan = make(chan *model.SecurityHitLog, hitLogChanSize)
		hitLogWg.Add(1)
		go func() {
			defer hitLogWg.Done()
			hitLogWorker()
		}()
	})
}

// hitLogWorker 批量消费命中日志并写入数据库
// worker 内部每次 select 都 recover，防止单条日志处理 panic 导致整个 worker 退出
func hitLogWorker() {
	batch := make([]*model.SecurityHitLog, 0, hitLogBatchSize)
	ticker := time.NewTicker(hitLogFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := flushHitLogs(batch); err != nil {
			common.SysLog("批量写入安全命中日志失败（已重试）: " + err.Error())
		}
		batch = batch[:0]
	}
	defer flush() // 退出前刷新剩余日志

	for {
		shouldExit := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					common.SysLog(fmt.Sprintf("安全命中日志 worker panic: %v", r))
				}
			}()

			select {
			case log, ok := <-hitLogChan:
				if !ok {
					shouldExit = true
					return
				}
				batch = append(batch, log)
				if len(batch) >= hitLogBatchSize {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}()
		if shouldExit {
			return
		}
	}
}

// flushHitLogs 批量写入命中日志，失败时自动重试
func flushHitLogs(logs []*model.SecurityHitLog) error {
	var lastErr error
	for i := 0; i < hitLogMaxRetries; i++ {
		if err := model.DB.CreateInBatches(logs, hitLogBatchSize).Error; err != nil {
			lastErr = err
			common.SysLog(fmt.Sprintf("批量写入安全命中日志失败（重试 %d/%d）: %v", i+1, hitLogMaxRetries, err))
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
			continue
		}
		return nil
	}
	return lastErr
}

// FlushSecurityHitLogs 安全刷新剩余命中日志，用于服务优雅退出
func FlushSecurityHitLogs() {
	if hitLogChan == nil {
		return
	}
	close(hitLogChan)
	hitLogWg.Wait()
	hitLogChan = nil
}

// recordHitLog 异步记录命中日志
func recordHitLog(userID int, originalContent string, result *DetectionResult, contentType int, modelName string, logCtx DetectionLogContext) {
	defer func() {
		if r := recover(); r != nil {
			common.SysLog(fmt.Sprintf("记录安全日志 panic: %v", r))
		}
	}()

	initHitLogWorker()

	hash := sha256.Sum256([]byte(originalContent))
	hashStr := hex.EncodeToString(hash[:])

	var processedContent string
	if result.Action == constant.SecurityActionMask {
		processedContent = result.ProcessedContent
	}

	var ruleID, groupID int64
	if len(result.Matches) > 0 {
		ruleID = result.Matches[0].RuleID
		groupID = result.Matches[0].GroupID
	}

	requestID := logCtx.RequestID
	if requestID == "" {
		requestID = generateRequestID()
	}

	log := &model.SecurityHitLog{
		RequestID:           requestID,
		UserID:              userID,
		ChannelID:           logCtx.ChannelID,
		ModelName:           modelName,
		TokenID:             logCtx.TokenID,
		RuleID:              ruleID,
		GroupID:             groupID,
		ContentType:         contentType,
		Action:              result.Action,
		RiskLevel:           result.RiskLevel,
		RiskScore:           result.RiskScore,
		OriginalContentHash: hashStr,
		ProcessedContent:    processedContent,
		CreatedAt:           time.Now().Unix(),
	}

	ch := hitLogChan
	if ch == nil {
		return
	}
	select {
	case ch <- log:
	default:
		dropped := droppedHitLogs.Add(1)
		common.SysLog(fmt.Sprintf("安全命中日志队列已满，丢弃一条日志（累计丢弃 %d）", dropped))
	}
}

func generateRequestID() string {
	return fmt.Sprintf("sec-%d-%d", time.Now().UnixNano(), common.GetRandomInt(10000))
}
