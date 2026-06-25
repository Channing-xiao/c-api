export const AISecurityRuleType = {
  KEYWORD: 1,
  REGEX: 2,
  NER: 3,
  AI: 4,
} as const

export const AISecurityAction = {
  ALLOW: 1,
  ALERT: 2,
  MASK: 3,
  BLOCK: 4,
  REVIEW: 5,
} as const

export const AISecurityRiskLevel = {
  LOW: 1,
  MEDIUM: 2,
  HIGH: 3,
  CRITICAL: 4,
} as const

export const AISecurityScope = {
  REQUEST: 1,
  RESPONSE: 2,
  BOTH: 3,
} as const

export const AISecurityStatus = {
  DISABLED: 0,
  ENABLED: 1,
} as const

export const ruleTypeOptions = [
  { value: AISecurityRuleType.KEYWORD, labelKey: 'aiSecurity.ruleTypes.keyword' },
  { value: AISecurityRuleType.REGEX, labelKey: 'aiSecurity.ruleTypes.regex' },
  { value: AISecurityRuleType.NER, labelKey: 'aiSecurity.ruleTypes.ner' },
  { value: AISecurityRuleType.AI, labelKey: 'aiSecurity.ruleTypes.ai' },
]

export const actionOptions = [
  { value: AISecurityAction.ALLOW, labelKey: 'aiSecurity.actions.allow', color: 'text-green-600' },
  { value: AISecurityAction.ALERT, labelKey: 'aiSecurity.actions.alert', color: 'text-yellow-600' },
  { value: AISecurityAction.MASK, labelKey: 'aiSecurity.actions.mask', color: 'text-blue-600' },
  { value: AISecurityAction.BLOCK, labelKey: 'aiSecurity.actions.block', color: 'text-red-600' },
  { value: AISecurityAction.REVIEW, labelKey: 'aiSecurity.actions.review', color: 'text-purple-600' },
]

export const scopeOptions = [
  { value: AISecurityScope.REQUEST, labelKey: 'aiSecurity.scopes.request' },
  { value: AISecurityScope.RESPONSE, labelKey: 'aiSecurity.scopes.response' },
  { value: AISecurityScope.BOTH, labelKey: 'aiSecurity.scopes.both' },
]

export const riskLevelOptions = [
  { value: AISecurityRiskLevel.LOW, labelKey: 'aiSecurity.riskLevels.low', color: 'bg-green-100 text-green-700' },
  { value: AISecurityRiskLevel.MEDIUM, labelKey: 'aiSecurity.riskLevels.medium', color: 'bg-yellow-100 text-yellow-700' },
  { value: AISecurityRiskLevel.HIGH, labelKey: 'aiSecurity.riskLevels.high', color: 'bg-orange-100 text-orange-700' },
  { value: AISecurityRiskLevel.CRITICAL, labelKey: 'aiSecurity.riskLevels.critical', color: 'bg-red-100 text-red-700' },
]

export const statusOptions = [
  { value: AISecurityStatus.ENABLED, labelKey: 'aiSecurity.status.enabled' },
  { value: AISecurityStatus.DISABLED, labelKey: 'aiSecurity.status.disabled' },
]
