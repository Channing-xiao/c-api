import { useTranslation } from 'react-i18next'

// 规则类型（与后端 constant/security.go SecurityRuleType 对应）
export const RULE_TYPES = [
  { value: 1, labelKey: 'Keyword Match' },
  { value: 2, labelKey: 'Regex Match' },
  { value: 3, labelKey: 'NER' },
  { value: 4, labelKey: 'AI Detection' },
] as const

// 处理动作（与后端 constant/security.go SecurityAction 对应）
export const ACTIONS = [
  { value: 1, labelKey: 'Pass' },
  { value: 2, labelKey: 'Alert' },
  { value: 3, labelKey: 'Mask' },
  { value: 4, labelKey: 'Block' },
  { value: 5, labelKey: 'Review' },
] as const

// 检测范围（与后端 constant/security.go SecurityScope 对应）
export const SCOPES = [
  { value: 1, labelKey: 'Request Only' },
  { value: 2, labelKey: 'Response Only' },
  { value: 3, labelKey: 'Both' },
] as const

// 启用状态（与后端 constant/security.go SecurityStatus 对应）
export const STATUSES = [
  { value: 0, labelKey: 'Disabled' },
  { value: 1, labelKey: 'Enabled' },
] as const

// 风险等级（与后端 constant/security.go SecurityRiskLevel 对应）
export const RISK_LEVELS = [
  { value: 1, labelKey: 'Low', color: 'bg-green-100 text-green-800' },
  { value: 2, labelKey: 'Medium', color: 'bg-yellow-100 text-yellow-800' },
  { value: 3, labelKey: 'High', color: 'bg-orange-100 text-orange-800' },
  { value: 4, labelKey: 'Critical', color: 'bg-red-100 text-red-800' },
] as const

export type OptionItem = {
  value: number
  labelKey: string
  color?: string
}

export function useSecurityOptions() {
  const { t } = useTranslation()

  const withLabel = (items: readonly OptionItem[]) =>
    items.map((item) => ({
      ...item,
      label: t(item.labelKey),
    }))

  return {
    ruleTypes: withLabel(RULE_TYPES),
    actions: withLabel(ACTIONS),
    scopes: withLabel(SCOPES),
    statuses: withLabel(STATUSES),
    riskLevels: withLabel(RISK_LEVELS),
    getLabel: (items: readonly OptionItem[], value: number, fallback = t('Unknown')) =>
      t(items.find((item) => item.value === value)?.labelKey ?? fallback),
    getRiskLevel: (value: number) =>
      RISK_LEVELS.find((item) => item.value === value),
  }
}
