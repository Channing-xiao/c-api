package security

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	goahocorasick "github.com/anknown/ahocorasick"
)

// KeywordDetector 关键词检测引擎
type KeywordDetector struct {
}

func (kd *KeywordDetector) Name() string {
	return "keyword"
}

// Detect 使用 Aho-Corasick 自动机检测关键词
func (kd *KeywordDetector) Detect(content string, rules []*model.SecurityRule) (*EngineResult, error) {
	result := &EngineResult{
		EngineName: kd.Name(),
		Detected:   false,
		Matches:    make([]*dto.SecurityMatchResult, 0),
		RiskScore:  0,
	}

	type keywordRule struct {
		keyword string
		rule    *model.SecurityRule
	}

	var allKeywords []string
	var keywordRules []keywordRule

	for _, rule := range rules {
		if rule.Type != constant.SecurityRuleTypeKeyword || rule.Status != constant.SecurityStatusEnabled {
			continue
		}

		keywords := strings.Split(rule.Content, ",")
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword == "" {
				continue
			}
			allKeywords = append(allKeywords, strings.ToLower(keyword))
			keywordRules = append(keywordRules, keywordRule{keyword: keyword, rule: rule})
		}
	}

	if len(allKeywords) == 0 {
		return result, nil
	}

	// 构建 AC 自动机
	m := new(goahocorasick.Machine)
	runesDict := make([][]rune, 0, len(allKeywords))
	for _, kw := range allKeywords {
		runesDict = append(runesDict, []rune(kw))
	}
	if err := m.Build(runesDict); err != nil {
		return result, err
	}

	// 在原始内容上建立 rune -> byte 位置映射，避免 ToLower 后字节长度变化导致切片错位
	contentRunes := []rune(content)
	runeToByte := make([]int, len(contentRunes)+1)
	byteIdx := 0
	for i, r := range contentRunes {
		runeToByte[i] = byteIdx
		byteIdx += utf8.RuneLen(r)
	}
	runeToByte[len(contentRunes)] = byteIdx

	// 搜索时转为小写以实现大小写不敏感匹配
	// 注意：Go 的 strings.ToLower 对大多数 Unicode 字符保持 rune 数量不变，
	// 但存在极少数特殊 locale 字符（如土耳其语 İ）会改变 rune 数量。
	// 此类场景下命中位置可能错位，属于可接受的边界情况。
	contentLowerRunes := []rune(strings.ToLower(content))
	hits := m.MultiPatternSearch(contentLowerRunes, false)

	// 建立小写关键词到规则列表的映射
	keywordRuleMap := make(map[string][]keywordRule)
	for _, kr := range keywordRules {
		lk := strings.ToLower(kr.keyword)
		keywordRuleMap[lk] = append(keywordRuleMap[lk], kr)
	}

	// 收集所有命中位置，同一规则在同一位置去重
	type matchKey struct {
		ruleID int64
		start  int
		end    int
	}
	seenMatches := make(map[matchKey]bool)
	seenWords := make(map[string]bool)

	for _, hit := range hits {
		word := string(hit.Word)
		rulesForWord, ok := keywordRuleMap[word]
		if !ok {
			continue
		}
		seenWords[word] = true
		start := runeToByte[hit.Pos]
		end := runeToByte[hit.Pos+len(hit.Word)]

		for _, kr := range rulesForWord {
			key := matchKey{ruleID: kr.rule.ID, start: start, end: end}
			if seenMatches[key] {
				continue
			}
			seenMatches[key] = true

			result.Detected = true
			if kr.rule.RiskScore > result.RiskScore {
				result.RiskScore = kr.rule.RiskScore
			}
			result.Matches = append(result.Matches, &dto.SecurityMatchResult{
				RuleID:      kr.rule.ID,
				GroupID:     kr.rule.GroupID,
				Type:        kr.rule.Type,
				MatchedText: content[start:end],
				Position:    [2]int{start, end},
			})
		}
	}

	var matchedRuleIDs []int64
	var matchedWords []string
	for id := range seenMatches {
		matchedRuleIDs = append(matchedRuleIDs, id.ruleID)
	}
	for word := range seenWords {
		matchedWords = append(matchedWords, word)
	}
	common.SysLog(fmt.Sprintf("[security:keyword] keywords=%d hits=%d detected=%v matchedRules=%v matchedWords=%v",
		len(allKeywords), len(hits), result.Detected, matchedRuleIDs, matchedWords))

	return result, nil
}
