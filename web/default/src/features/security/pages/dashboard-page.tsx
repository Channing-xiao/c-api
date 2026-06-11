import React, { Suspense, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
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
import { securityApi, type DashboardData } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
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
            {t('Refresh')}
          </Button>
        </>
      }
    >
      <div className="space-y-6">
        <div className="flex items-center justify-between rounded-lg border p-4">
          <div>
            <h3 className="text-sm font-medium">{t('Global Security Detection')}</h3>
            <p className="text-xs text-muted-foreground">
              {t('Enable or disable AI content security detection globally.')}
            </p>
          </div>
          <Switch checked={true} disabled={true} aria-readonly="true" />
        </div>

        <div className="flex items-center gap-2">
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
        </div>

        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            {Array.from({ length: 5 }).map((_, i) => (
              <Card key={i}>
                <CardHeader className="pb-2">
                  <Skeleton className="h-4 w-24" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-8 w-16" />
                </CardContent>
              </Card>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('Total Detections')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{data?.summary?.total_detections ?? 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('Interceptions')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">{data?.summary?.total_interceptions ?? 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t('Alerts')}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-yellow-600">{data?.summary?.total_alerts ?? 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t("Today's Detections")}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{data?.summary?.today_detections ?? 0}</div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                  {t("Today's Interceptions")}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">{data?.summary?.today_interceptions ?? 0}</div>
              </CardContent>
            </Card>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Suspense fallback={<ChartSkeleton />}>
            <TopCategoriesChart data={data?.top_categories ?? []} />
          </Suspense>
          <Suspense fallback={<ChartSkeleton />}>
            <RiskDistributionChart data={data?.risk_distribution ?? { low: 0, medium: 0, high: 0, critical: 0 }} />
          </Suspense>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <TopUsersTable data={data?.top_users ?? []} />
          <Card>
            <CardHeader>
              <CardTitle>{t('Top Models')}</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-2">
                {data?.top_models?.map((item: any, idx: number) => (
                  <li key={idx} className="flex justify-between">
                    <span>{item.model_name}</span>
                    <span className="font-medium">{item.count}</span>
                  </li>
                )) ?? <li className="text-muted-foreground">{t('No data')}</li>}
              </ul>
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
