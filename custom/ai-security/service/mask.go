package service

import (
	"strings"

	constant "github.com/QuantumNous/new-api/custom/ai-security/constant"
	"github.com/QuantumNous/new-api/custom/ai-security/dto"
	"github.com/QuantumNous/new-api/custom/ai-security/model"
)

// MaskStrategy 脱敏策略
type MaskStrategy string

const (
	MaskStrategyFull    MaskStrategy = "full"
	MaskStrategyPreserve MaskStrategy = "preserve"
	MaskStrategyReplace  MaskStrategy = "replace"
)

// Masker 脱敏器
type Masker struct {
	strategy      MaskStrategy
	preserveChars int
	replaceChar   rune
}

// NewMasker 创建脱敏器
func NewMasker(strategy string, preserveChars int) *Masker {
	s := MaskStrategy(strategy)
	if s == "" {
		s = MaskStrategyFull
	}
	if preserveChars < 0 {
		preserveChars = 0
	}
	return &Masker{
		strategy:      s,
		preserveChars: preserveChars,
		replaceChar:   '*',
	}
}

// Mask 对文本按检测结果执行脱敏
func (m *Masker) Mask(content string, result *dto.DetectionResult) string {
	if !result.Detected || len(result.Matches) == 0 {
		return content
	}

	// 按位置从后向前替换，避免索引偏移
	matches := make([]*dto.MatchResult, len(result.Matches))
	copy(matches, result.Matches)
	sortMatches(matches)

	out := []rune(content)
	for _, match := range matches {
		if match.Position[0] < 0 || match.Position[1] > len(out) || match.Position[0] >= match.Position[1] {
			continue
		}
		start := match.Position[0]
		end := match.Position[1]
		maskedRunes := []rune(m.maskSegment(string(out[start:end])))
		out = append(out[:start], append(maskedRunes, out[end:]...)...)
	}

	return string(out)
}

// maskSegment 对单个片段脱敏
func (m *Masker) maskSegment(segment string) string {
	runes := []rune(segment)
	switch m.strategy {
	case MaskStrategyPreserve:
		if m.preserveChars >= len(runes) {
			return strings.Repeat(string(m.replaceChar), len(runes))
		}
		keep := m.preserveChars / 2
		if keep == 0 {
			keep = 1
		}
		prefix := runes[:keep]
		suffix := runes[len(runes)-keep:]
		middle := strings.Repeat(string(m.replaceChar), len(runes)-keep*2)
		return string(prefix) + middle + string(suffix)
	case MaskStrategyReplace:
		return strings.Repeat(string(m.replaceChar), len(runes))
	case MaskStrategyFull:
		fallthrough
	default:
		return strings.Repeat(string(m.replaceChar), len(runes))
	}
}

// MaskWithDefault 使用默认配置脱敏
func MaskWithDefault(content string, result *dto.DetectionResult) string {
	strategy := model.GetStringConfig(constant.ConfigKeyMaskStrategy, string(MaskStrategyFull))
	preserve := model.GetIntConfig(constant.ConfigKeyMaskPreserveChars, 0)
	return NewMasker(strategy, preserve).Mask(content, result)
}

func sortMatches(matches []*dto.MatchResult) {
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Position[0] < matches[j].Position[0] {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}
