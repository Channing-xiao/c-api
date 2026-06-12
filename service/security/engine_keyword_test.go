package security

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"
)

func TestKeywordDetector_ConsecutiveDetections(t *testing.T) {
	detector := &KeywordDetector{}
	rules := []*model.SecurityRule{
		{
			ID:        1,
			GroupID:   1,
			Type:      constant.SecurityRuleTypeKeyword,
			Content:   "敏感词",
			Action:    constant.SecurityActionBlock,
			RiskScore: 80,
			Status:    constant.SecurityStatusEnabled,
		},
	}

	// 第一次命中
	result, err := detector.Detect("包含敏感词的内容", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Detected {
		t.Fatalf("expected first request detected")
	}

	// 第二次未命中（回归：修复前可能因状态污染而误报）
	result, err = detector.Detect("普通正常内容", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Detected {
		t.Fatalf("expected second request not detected, got detected")
	}

	// 第三次再次命中
	result, err = detector.Detect("又有敏感词出现", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Detected {
		t.Fatalf("expected third request detected")
	}
}

func TestKeywordDetector_Detect(t *testing.T) {
	detector := &KeywordDetector{}
	rules := []*model.SecurityRule{
		{
			ID:        1,
			GroupID:   1,
			Type:      constant.SecurityRuleTypeKeyword,
			Content:   "机密, 密码",
			Action:    constant.SecurityActionBlock,
			RiskScore: 80,
			Status:    constant.SecurityStatusEnabled,
		},
		{
			ID:        2,
			GroupID:   1,
			Type:      constant.SecurityRuleTypeRegex,
			Content:   `\d{18}`,
			Action:    constant.SecurityActionMask,
			RiskScore: 60,
			Status:    constant.SecurityStatusEnabled,
		},
	}

	result, err := detector.Detect("这是机密信息", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Detected {
		t.Fatalf("expected detected, got not detected")
	}
	if result.RiskScore != 80 {
		t.Fatalf("expected risk score 80, got %d", result.RiskScore)
	}
	if len(result.Matches) == 0 {
		t.Fatalf("expected matches, got none")
	}

	// 未命中
	result, err = detector.Detect("普通内容", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Detected {
		t.Fatalf("expected not detected, got detected")
	}

	// 禁用规则不应命中
	rules[0].Status = constant.SecurityStatusDisabled
	result, err = detector.Detect("这是机密信息", rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Detected {
		t.Fatalf("expected not detected when rule disabled")
	}
}
