export const aiSecurityQueryKeys = {
  groups: (params?: Record<string, unknown>) => ['ai-security', 'groups', params] as const,
  rules: (params?: Record<string, unknown>) => ['ai-security', 'rules', params] as const,
  policies: (params?: Record<string, unknown>) => ['ai-security', 'policies', params] as const,
  logs: (params?: Record<string, unknown>) => ['ai-security', 'logs', params] as const,
  dashboard: (params?: Record<string, unknown>) => ['ai-security', 'dashboard', params] as const,
  status: () => ['ai-security', 'status'] as const,
  configs: () => ['ai-security', 'configs'] as const,
} as const
