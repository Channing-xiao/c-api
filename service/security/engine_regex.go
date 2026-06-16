package security

import (
	"fmt"
	"math"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/dlclark/regexp2"
)

const regexCacheMaxSize = 1000

type regexCacheEntry struct {
	re       *regexp2.Regexp
	lastUsed int64 // UnixNano，用于 LRU 淘汰
}

var (
	regexCache   = make(map[string]*regexCacheEntry)
	regexCacheMu sync.RWMutex
)

// RegexDetector 正则检测引擎
type RegexDetector struct{}

func (rd *RegexDetector) Name() string { return "regex" }

func (rd *RegexDetector) getCompiled(pattern string) (*regexp2.Regexp, error) {
	regexCacheMu.RLock()
	entry, ok := regexCache[pattern]
	regexCacheMu.RUnlock()
	if ok {
		// 更新最近使用时间
		regexCacheMu.Lock()
		entry.lastUsed = time.Now().UnixNano()
		regexCacheMu.Unlock()
		return entry.re, nil
	}

	re, err := regexp2.Compile(pattern, 0)
	if err != nil {
		return nil, err
	}

	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()

	// 缓存达到上限时淘汰最久未使用的条目
	if len(regexCache) >= regexCacheMaxSize {
		var oldestKey string
		var oldestTime int64 = math.MaxInt64
		for k, v := range regexCache {
			if v.lastUsed < oldestTime {
				oldestTime = v.lastUsed
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(regexCache, oldestKey)
		}
	}

	regexCache[pattern] = &regexCacheEntry{re: re, lastUsed: time.Now().UnixNano()}
	return re, nil
}

// Detect 使用正则表达式检测
func (rd *RegexDetector) Detect(content string, rules []*model.SecurityRule) (*EngineResult, error) {
	result := &EngineResult{
		EngineName: rd.Name(),
		Detected:   false,
		Matches:    make([]*dto.SecurityMatchResult, 0),
		RiskScore:  0,
	}

	// 建立 rune 位置到 byte 位置的映射（与 keyword 引擎保持一致）
	contentRunes := []rune(content)
	runeToByte := make([]int, len(contentRunes)+1)
	byteIdx := 0
	for i, r := range contentRunes {
		runeToByte[i] = byteIdx
		byteIdx += utf8.RuneLen(r)
	}
	runeToByte[len(contentRunes)] = byteIdx

	for _, rule := range rules {
		if rule.Type != constant.SecurityRuleTypeRegex || rule.Status != constant.SecurityStatusEnabled {
			continue
		}

		re, err := rd.getCompiled(rule.Content)
		if err != nil {
			continue // 跳过无效正则
		}

		match, err := re.FindStringMatch(content)
		if err != nil {
			continue
		}

		// 遍历该正则的所有命中位置
		for match != nil {
			result.Detected = true
			if rule.RiskScore > result.RiskScore {
				result.RiskScore = rule.RiskScore
			}

			matchedText := match.String()
			common.SysLog(fmt.Sprintf("[security:regex] ruleID=%d matchedText=%q", rule.ID, matchedText))

			// regexp2 的 Index/Length 是 rune 位置，转换为 byte 位置
			start := runeToByte[match.Index]
			end := runeToByte[match.Index+match.Length]

			result.Matches = append(result.Matches, &dto.SecurityMatchResult{
				RuleID:      rule.ID,
				GroupID:     rule.GroupID,
				Type:        rule.Type,
				MatchedText: matchedText,
				Position:    [2]int{start, end},
			})

			match, err = re.FindNextMatch(match)
			if err != nil {
				break
			}
		}
	}

	return result, nil
}
