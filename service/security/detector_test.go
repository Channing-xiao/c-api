package security

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
)

func TestResolveAction(t *testing.T) {
	rules := []*model.SecurityRule{
		{ID: 1, Action: constant.SecurityActionPass},
		{ID: 2, Action: constant.SecurityActionAlert},
		{ID: 3, Action: constant.SecurityActionMask},
		{ID: 4, Action: constant.SecurityActionBlock},
		{ID: 5, Action: constant.SecurityActionReview},
	}

	matches := []*dto.SecurityMatchResult{
		{RuleID: 1},
		{RuleID: 3},
	}

	action := resolveAction(matches, rules)
	if action != constant.SecurityActionMask {
		t.Fatalf("expected mask (3), got %d", action)
	}

	matches = []*dto.SecurityMatchResult{
		{RuleID: 2},
		{RuleID: 4},
	}

	action = resolveAction(matches, rules)
	if action != constant.SecurityActionBlock {
		t.Fatalf("expected block (4), got %d", action)
	}
}

func TestApplyMasking(t *testing.T) {
	rules := []*model.SecurityRule{{ID: 1}}
	matches := []*dto.SecurityMatchResult{
		{RuleID: 1, Position: [2]int{10, 21}},
	}

	result := applyMasking("请联系 13800138000", matches, rules)
	expected := "请联系 ***"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}

	// 验证乱序 matches 也能正确替换
	matches2 := []*dto.SecurityMatchResult{
		{RuleID: 1, Position: [2]int{13, 24}},
		{RuleID: 1, Position: [2]int{0, 11}},
	}
	result2 := applyMasking("13800138000 请联系 13800138000", matches2, rules)
	expected2 := "*** 请联系 ***"
	if result2 != expected2 {
		t.Fatalf("expected %q, got %q", expected2, result2)
	}
}

func TestMaskText(t *testing.T) {
	if maskText("ab", nil) != "***" {
		t.Fatalf("expected ***, got %s", maskText("ab", nil))
	}
	if maskText("abc", nil) != "***" {
		t.Fatalf("expected ***, got %s", maskText("abc", nil))
	}
	if maskText("abcd", nil) != "***" {
		t.Fatalf("expected ***, got %s", maskText("abcd", nil))
	}

	// 自定义保留前后位数的脱敏
	config := &MaskConfig{MaskType: "preserve", PreserveStart: 6, PreserveEnd: 4, MaskChar: "*"}
	if got := maskText("110101199001011234", config); got != "110101********1234" {
		t.Fatalf("expected 110101********1234, got %s", got)
	}

	// 边界：保留位数超过文本长度
	config2 := &MaskConfig{MaskType: "preserve", PreserveStart: 10, PreserveEnd: 10, MaskChar: "*"}
	if got := maskText("12345", config2); got != "*" {
		t.Fatalf("expected *, got %s", got)
	}

	// 自定义整体替换字符串
	config3 := &MaskConfig{MaskType: "full", ReplaceWith: "[已脱敏]"}
	if got := maskText("110101199001011234", config3); got != "[已脱敏]" {
		t.Fatalf("expected [已脱敏], got %s", got)
	}
}

func TestApplyMaskingWithCustomConfig(t *testing.T) {
	rules := []*model.SecurityRule{{
		ID:          1,
		ExtraConfig: `{"mask_type":"preserve","preserve_start":6,"preserve_end":4,"mask_char":"*"}`,
	}}
	matches := []*dto.SecurityMatchResult{
		{RuleID: 1, Position: [2]int{7, 25}},
	}

	result := applyMasking("身份证 110101199001011234 请查收", matches, rules)
	expected := "身份证 110101********1234 请查收"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

