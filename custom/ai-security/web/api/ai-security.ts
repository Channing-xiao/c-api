import { api } from '@/lib/api'

export interface AISecurityGroup {
  id: number
  name: string
  description: string
  status: number
  parent_id: number
  depth: number
  path: string
  sort_order: number
  created_at: number
  updated_at: number
}

export interface AISecurityRule {
  id: number
  group_id: number
  group_name?: string
  name: string
  type: number
  content: string
  extra_config: string
  action: number
  priority: number
  risk_score: number
  status: number
  is_seed?: boolean
  seed_code?: string
  created_at: number
  updated_at: number
}

export interface AISecurityPolicy {
  id: number
  user_id: number
  user_name?: string
  group_id: number
  group_name?: string
  scope: number
  default_action: number
  custom_response: string
  whitelist_ips: string
  priority: number
  status: number
  created_at: number
  updated_at: number
}

export interface AISecurityHitLog {
  id: number
  request_id: string
  user_id: number
  user_name?: string
  model_name: string
  content_type: number
  action: number
  risk_level: number
  risk_score: number
  rule_id?: number
  group_id?: number
  original_content_hash: string
  processed_content?: string
  match_detail?: string
  ip: string
  created_at: number
}

export interface AISecurityDashboardData {
  summary: {
    total_detections: number
    total_interceptions: number
    total_alerts: number
    today_detections: number
    today_interceptions: number
  }
  top_categories: Array<{ category: string; count: number }>
  top_users: Array<{ user_id: number; user_name: string; count: number }>
  top_models: Array<{ model_name: string; count: number }>
  risk_distribution: { low: number; medium: number; high: number; critical: number }
}

export interface AISecurityRuleTestResult {
  detected: boolean
  action: number
  action_name: string
  risk_score: number
  risk_level: number
  processed_content?: string
  matches: Array<{
    rule_id: number
    group_id: number
    type: number
    matched_text: string
    position: [number, number]
  }>
}

export interface AISecurityStatus {
  enabled: boolean
  version: string
  group_count: number
  rule_count: number
  policy_count: number
  hit_count: number
}

export interface AISecurityConfigs {
  enabled: boolean
  max_group_depth: number
  ai_timeout_ms: number
  log_retention_days: number
  default_action: number
  mask_strategy: string
  mask_preserve_chars: number
}

export const aiSecurityApi = {
  // Configs
  getConfigs: () => api.get('/api/ai-security/configs').then((r) => r.data),
  updateConfigs: (data: Partial<AISecurityConfigs>) =>
    api.put('/api/ai-security/configs', data).then((r) => r.data),

  // Status
  getStatus: () => api.get('/api/ai-security/status').then((r) => r.data),

  // Install
  install: () => api.post('/api/ai-security/install').then((r) => r.data),

  // Groups
  getGroups: (params?: { page?: number; page_size?: number; status?: number; parent_id?: number; name?: string }) => api.get('/api/ai-security/groups', { params }).then((r) => r.data),
  createGroup: (data: Partial<AISecurityGroup>) => api.post('/api/ai-security/groups', data).then((r) => r.data),
  updateGroup: (id: number, data: Partial<AISecurityGroup>) =>
    api.put(`/api/ai-security/groups/${id}`, data).then((r) => r.data),
  updateGroupStatus: (id: number, status: number) =>
    api.patch(`/api/ai-security/groups/${id}/status`, { status }).then((r) => r.data),
  deleteGroup: (id: number) => api.delete(`/api/ai-security/groups/${id}`).then((r) => r.data),
  copyGroup: (id: number) => api.post(`/api/ai-security/groups/${id}/copy`).then((r) => r.data),

  // Rules
  getRules: (params?: { page?: number; page_size?: number; group_id?: number; type?: number; status?: number }) => api.get('/api/ai-security/rules', { params }).then((r) => r.data),
  createRule: (data: Partial<AISecurityRule>) => api.post('/api/ai-security/rules', data).then((r) => r.data),
  updateRule: (id: number, data: Partial<AISecurityRule>) =>
    api.put(`/api/ai-security/rules/${id}`, data).then((r) => r.data),
  deleteRule: (id: number) => api.delete(`/api/ai-security/rules/${id}`).then((r) => r.data),
  testRule: (id: number, content: string) =>
    api.post(`/api/ai-security/rules/${id}/test`, { content }).then((r) => r.data),
  updateRuleStatus: (id: number, status: number) =>
    api.patch(`/api/ai-security/rules/${id}/status`, { status }).then((r) => r.data),
  batchDeleteRules: (ids: number[]) =>
    api.post('/api/ai-security/rules/batch-delete', { ids }).then((r) => r.data),
  batchUpdateRuleStatus: (ids: number[], status: number) =>
    api.post('/api/ai-security/rules/batch-status', { ids, status }).then((r) => r.data),

  // Policies
  getPolicies: (params?: { page?: number; page_size?: number; user_id?: number; status?: number }) => api.get('/api/ai-security/policies', { params }).then((r) => r.data),
  createPolicy: (data: Partial<AISecurityPolicy>) => api.post('/api/ai-security/policies', data).then((r) => r.data),
  updatePolicy: (id: number, data: Partial<AISecurityPolicy>) =>
    api.put(`/api/ai-security/policies/${id}`, data).then((r) => r.data),
  deletePolicy: (id: number) => api.delete(`/api/ai-security/policies/${id}`).then((r) => r.data),

  // Logs
  getLogs: (params?: {
    page?: number
    page_size?: number
    user_id?: number
    action?: number
    risk_level?: number
    content_type?: number
    rule_id?: number
    group_id?: number
    start_time?: number
    end_time?: number
    model_name?: string
  }) => api.get('/api/ai-security/logs', { params }).then((r) => r.data),
  exportLogs: (params?: {
    format?: 'csv' | 'excel'
    user_id?: number
    action?: number
    risk_level?: number
    content_type?: number
    rule_id?: number
    group_id?: number
    start_time?: number
    end_time?: number
    model_name?: string
  }) =>
    api.get('/api/ai-security/logs/export', {
      params,
      responseType: 'blob',
      skipBusinessError: true,
    }),

  // Dashboard
  getDashboard: (params?: {
    start_time?: number
    end_time?: number
    user_id?: number
    group_id?: number
    rule_id?: number
  }) => api.get('/api/ai-security/dashboard', { params }).then((r) => r.data),

  // Dashboard historical trend (from aisec_daily_stats)
  getDailyTrend: (params?: { days?: number }) =>
    api.get('/api/ai-security/dashboard/trend', { params }).then((r) => r.data),

  // Sync
  syncOfficialSensitiveWords: () =>
    api.post('/api/ai-security/sync/official-sensitive-words').then((r) => r.data),
}
