package security

import (
	"container/list"
	"fmt"
	"sync"
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
	listElem *list.Element
}

var (
	regexCache   = make(map[string]*regexCacheEntry)
	regexCacheMu sync.Mutex
	regexLRU     = list.New()
)

// RegexDetector 正则检测引擎
type RegexDetector struct{}

func (rd *RegexDetector) Name() string { return "regex" }

func (rd *RegexDetector) getCompiled(pattern string) (*regexp2.Regexp, error) {
	regexCacheMu.Lock()
	if entry, ok := regexCache[pattern]; ok {
		regexLRU.MoveToFront(entry.listElem)
		regexCacheMu.Unlock()
		return entry.re, nil
	}
	regexCacheMu.Unlock()

	re, err := regexp2.Compile(pattern, 0)
	if err != nil {
		return nil, err
	}

	regexCacheMu.Lock()
	defer regexCacheMu.Unlock()

	// 编译期间可能已有其他 goroutine 写入同一 pattern，再次检查避免重复缓存
	if entry, ok := regexCache[pattern]; ok {
		regexLRU.MoveToFront(entry.listElem)
		return entry.re, nil
	}

	// 缓存达到上限时淘汰最久未使用的条目（O(1) LRU）
	if len(regexCache) >= regexCacheMaxSize {
		if evict := regexLRU.Back(); evict != nil {
			regexLRU.Remove(evict)
			delete(regexCache, evict.Value.(string))
		}
	}

	elem := regexLRU.PushFront(pattern)
	regexCache[pattern] = &regexCacheEntry{re: re, listElem: elem}
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
