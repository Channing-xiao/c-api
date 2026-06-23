export const securityQueryKeys = {
  groups: (params?: Record<string, unknown>) => ['security', 'groups', params] as const,
  rules: (params?: Record<string, unknown>) => ['security', 'rules', params] as const,
  policies: (params?: Record<string, unknown>) => ['security', 'policies', params] as const,
  logs: (params?: Record<string, unknown>) => ['security', 'logs', params] as const,
  dashboard: (params?: Record<string, unknown>) => ['security', 'dashboard', params] as const,
  status: () => ['security', 'status'] as const,
  migration: () => ['security', 'migration'] as const,
}
