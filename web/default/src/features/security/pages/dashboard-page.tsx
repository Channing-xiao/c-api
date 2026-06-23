import React, { Suspense, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Activity,
  AlertTriangle,
  Bot,
  CalendarClock,
  CalendarDays,
  Filter,
  RefreshCw,
  Shield,
  ShieldCheck,
  Users,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { EmptyState } from '@/components/empty-state'
import { securityApi, type DashboardData } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
import { SecurityStatCard } from '../components/security-stat-card'
import { TopUsersTable } from '../components/top-users-table'

const RiskDistributionChart = React.lazy(
  () => import('../components/risk-distribution-chart').then((m) => ({ default: m.RiskDistributionChart }))
)
const TopCategoriesChart = React.lazy(
  () => import('../components/top-categories-chart').then((m) => ({ default: m.TopCategoriesChart }))
)

const timeRangeOptions = [
  { value: 'today', label: 'Today' },
  { value: 'week', label: 'This Week' },
  { value: 'month', label: 'This Month' },
  { value: 'custom', label: 'Custom' },
]

function getTimeRange(value: string): { start_time?: number; end_time?: number } {
  const now = Math.floor(Date.now() / 1000)
  const day = 86400
  switch (value) {
    case 'today':
      return { start_time: now - day, end_time: now }
    case 'week':
      return { start_time: now - day * 7, end_time: now }
    case 'month':
      return { start_time: now - day * 30, end_time: now }
    default:
      return {}
  }
}

export function SecurityDashboardPage() {
  const { t } = useTranslation()
  const [data, setData] = useState<DashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('today')
  const [autoRefresh, setAutoRefresh] = useState(false)

  const loadDashboard = () => {
    setLoading(true)
    const params = getTimeRange(timeRange)
    securityApi.getDashboard(params).then((res: any) => {
      if (res.success) {
        setData(res.data)
      }
      setLoading(false)
    })
  }

  useEffect(() => {
    loadDashboard()
  }, [timeRange])

  useEffect(() => {
    if (!autoRefresh) return
    const interval = setInterval(() => loadDashboard(), 30000)
    return () => clearInterval(interval)
  }, [autoRefresh, timeRange])

  const statItems = [
    {
      title: t('Total Detections'),
      value: data?.summary?.total_detections ?? 0,
      icon: Activity,
      tone: 'default' as const,
      description: t('All time detected content'),
    },
    {
      title: t('Interceptions'),
      value: data?.summary?.total_interceptions ?? 0,
      icon: ShieldCheck,
      tone: 'rose' as const,
      description: t('Blocked or masked content'),
    },
    {
      title: t('Alerts'),
      value: data?.summary?.total_alerts ?? 0,
      icon: AlertTriangle,
      tone: 'amber' as const,
      description: t('Flagged for review'),
    },
    {
      title: t("Today's Detections"),
      value: data?.summary?.today_detections ?? 0,
      icon: CalendarDays,
      tone: 'teal' as const,
      description: t('In the last 24 hours'),
    },
    {
      title: t("Today's Interceptions"),
      value: data?.summary?.today_interceptions ?? 0,
      icon: Shield,
      tone: 'rose' as const,
      description: t('In the last 24 hours'),
    },
  ]

  return (
    <SecurityPageLayout
      actions={
        <>
          <Button
            variant={autoRefresh ? 'default' : 'outline'}
            size="sm"
            onClick={() => setAutoRefresh((v) => !v)}
          >
            {autoRefresh ? t('Auto Refresh: ON') : t('Auto Refresh: OFF')}
          </Button>
          <Button variant="outline" size="sm" onClick={loadDashboard}>
            <RefreshCw className="mr-1.5 size-3.5" />
            {t('Refresh')}
          </Button>
        </>
      }
    >
      <div className="space-y-6">
        <Card className="border-border/60 bg-gradient-to-r from-primary/5 via-transparent to-transparent">
          <CardContent className="flex items-center justify-between gap-4 py-4">
            <div className="flex items-center gap-3">
              <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
                <ShieldCheck className="size-5" />
              </div>
              <div>
                <h3 className="text-sm font-medium">{t('Global Security Detection')}</h3>
                <p className="text-xs text-muted-foreground">
                  {t('Enable or disable AI content security detection globally.')}
                </p>
              </div>
            </div>
            <Switch checked={true} disabled={true} aria-readonly="true" />
          </CardContent>
        </Card>

        <div className="flex items-center gap-2">
          <Filter className="size-4 text-muted-foreground" />
          <Select value={timeRange} onValueChange={(v) => setTimeRange(v ?? 'today')}>
            <SelectTrigger className="w-40">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {timeRangeOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {t(opt.label)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <span className="text-xs text-muted-foreground">
            <CalendarClock className="mr-1 inline size-3" />
            {t('Data updates every 30s when auto-refresh is enabled')}
          </span>
        </div>

        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
          {statItems.map((item) => (
            <SecurityStatCard
              key={item.title}
              title={item.title}
              value={item.value}
              description={item.description}
              icon={item.icon}
              tone={item.tone}
              loading={loading}
            />
          ))}
        </div>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <Suspense fallback={<ChartSkeleton />}>
            <TopCategoriesChart data={data?.top_categories ?? []} />
          </Suspense>
          <Suspense fallback={<ChartSkeleton />}>
            <RiskDistributionChart data={data?.risk_distribution ?? { low: 0, medium: 0, high: 0, critical: 0 }} />
          </Suspense>
        </div>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <TopUsersTable data={data?.top_users ?? []} />
          <Card>
            <CardHeader className="flex flex-row items-center gap-2">
              <Bot className="size-4 text-muted-foreground" />
              <CardTitle>{t('Top Models')}</CardTitle>
            </CardHeader>
            <CardContent>
              {data?.top_models?.length ? (
                <ul className="space-y-2">
                  {data.top_models.map((item: any, idx: number) => (
                    <li
                      key={idx}
                      className="flex items-center justify-between rounded-lg border px-3 py-2 transition-colors hover:bg-muted/50"
                    >
                      <span className="text-sm">{item.model_name}</span>
                      <span className="text-sm font-medium tabular-nums">{item.count}</span>
                    </li>
                  ))}
                </ul>
              ) : (
                <EmptyState
                  icon={Bot}
                  title={t('No Data')}
                  description={t('No model detection data for the selected period.')}
                  className="min-h-[180px] rounded-lg border border-dashed"
                  bordered={false}
                />
              )}
            </CardContent>
          </Card>
        </div>
      </div>
    </SecurityPageLayout>
  )
}

function ChartSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-5 w-32" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-[240px] w-full" />
      </CardContent>
    </Card>
  )
}
