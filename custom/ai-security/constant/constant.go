package ai_security

// Rule types
const (
	RuleTypeKeyword = iota + 1
	RuleTypeRegex
	RuleTypeNER
	RuleTypeAI
)

// Actions
const (
	ActionAllow = iota + 1
	ActionAlert
	ActionMask
	ActionBlock
	ActionReview
)

// Action priorities (higher value wins)
const (
	ActionPriorityAllow  = 1
	ActionPriorityAlert  = 2
	ActionPriorityMask   = 3
	ActionPriorityBlock  = 4
	ActionPriorityReview = 5
)

// Risk levels
const (
	RiskLevelLow = iota + 1
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

// Content directions
const (
	DirectionRequest = iota + 1
	DirectionResponse
)

// Scopes
const (
	ScopeRequestOnly = iota + 1
	ScopeResponseOnly
	ScopeBoth
)

// Status
const (
	StatusDisabled = iota
	StatusEnabled
)

// Config keys
const (
	ConfigKeyEnabled             = "ai_security_enabled"
	ConfigKeyAIModel             = "ai_model_name"
	ConfigKeyAITimeout           = "ai_timeout_seconds"
	ConfigKeyLogRetention        = "log_retention_days"
	ConfigKeyAuditLogRetention   = "audit_log_retention_days"
	ConfigKeyDefaultRiskScore    = "default_risk_score"
	ConfigKeyMaxGroupDepth       = "max_group_depth"
	ConfigKeyMaskStrategy        = "mask_strategy"
	ConfigKeyMaskPreserveChars   = "mask_preserve_chars"
)

// Defaults
const (
	DefaultAIModel           = "gpt-4o-mini"
	DefaultAITimeoutSeconds  = 3
	DefaultLogRetentionDays  = 30
	DefaultAuditRetentionDays = 90
	DefaultRiskScore         = 50
	DefaultMaxGroupDepth     = 5
)

// GetActionPriority returns priority for an action
func GetActionPriority(action int) int {
	switch action {
	case ActionAllow:
		return ActionPriorityAllow
	case ActionAlert:
		return ActionPriorityAlert
	case ActionMask:
		return ActionPriorityMask
	case ActionBlock:
		return ActionPriorityBlock
	case ActionReview:
		return ActionPriorityReview
	default:
		return ActionPriorityAllow
	}
}

// GetRiskLevelByScore returns risk level for a score
func GetRiskLevelByScore(score int) int {
	switch {
	case score <= 25:
		return RiskLevelLow
	case score <= 50:
		return RiskLevelMedium
	case score <= 75:
		return RiskLevelHigh
	default:
		return RiskLevelCritical
	}
}

// GetRiskLevelName returns human-readable risk level
func GetRiskLevelName(level int) string {
	switch level {
	case RiskLevelLow:
		return "low"
	case RiskLevelMedium:
		return "medium"
	case RiskLevelHigh:
		return "high"
	case RiskLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// GetActionName returns human-readable action name
func GetActionName(action int) string {
	switch action {
	case ActionAllow:
		return "allow"
	case ActionAlert:
		return "alert"
	case ActionMask:
		return "mask"
	case ActionBlock:
		return "block"
	case ActionReview:
		return "review"
	default:
		return "unknown"
	}
}
