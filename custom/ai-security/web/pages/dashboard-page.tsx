import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { BarChart3, ShieldAlert, ShieldCheck, ShieldX } from 'lucide-react'
import {
  aiSecurityApi,
  type AISecurityDashboardData,
  type AISecurityGroup,
  type AISecurityRule,
} from '../api/ai-security'
import { AISecurityLayout } from '../components/ai-security-layout'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { riskLevelOptions } from '../constants'

type TimeRange = 'today' | 'week' | 'month' | 'custom'

export function AISecurityDashboardPage() {
  const { t } = useTranslation()
  const [data, setData] = useState<AISecurityDashboardData | null>(null)
  const [loading, setLoading] = useState(true)
  const [range, setRange] = useState<TimeRange>('week')
  const [customStart, setCustomStart] = useState('')
  const [customEnd, setCustomEnd] = useState('')
  const [groups, setGroups] = useState<AISecurityGroup[]>([])
  const [rules, setRules] = useState<AISecurityRule[]>([])
  const [groupFilter, setGroupFilter] = useState('0')
  const [ruleFilter, setRuleFilter] = useState('0')
  const [userFilter, setUserFilter] = useState('')

  useEffect(() => {
    aiSecurityApi.getGroups({ page: 1, page_size: 200 }).then((res: any) => {
      if (res.success) setGroups(res.data.items)
    })
    aiSecurityApi.getRules({ page: 1, page_size: 200 }).then((res: any) => {
      if (res.success) setRules(res.data.items)
    })
  }, [])

  const computeRange = (): { start: number; end: number } => {
    const now = new Date()
    const endSec = Math.floor(now.getTime() / 1000)
    if (range === 'custom') {
      const start = customStart ? Math.floor(new Date(customStart).getTime() / 1000) : 0
      const end = customEnd ? Math.floor(new Date(customEnd).getTime() / 1000) + 86399 : endSec
      return { start, end }
    }
    if (range === 'today') {
      const midnight = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      return { start: Math.floor(midnight.getTime() / 1000), end: endSec }
    }
    const days = range === 'month' ? 30 : 7
    return { start: endSec - days * 24 * 3600, end: endSec }
  }

  const loadDashboard = () => {
    setLoading(true)
    const { start, end } = computeRange()
    aiSecurityApi
      .getDashboard({
        start_time: start,
        end_time: end,
        group_id: Number(groupFilter) || undefined,
        rule_id: Number(ruleFilter) || undefined,
        user_id: Number(userFilter) || undefined,
      })
      .then((res: any) => {
        if (res.success) setData(res.data)
      })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    loadDashboard()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [range, customStart, customEnd, groupFilter, ruleFilter])

  const summary = data?.summary ?? {
    total_detections: 0,
    total_interceptions: 0,
    total_alerts: 0,
    today_detections: 0,
    today_interceptions: 0,
  }

  const riskTotal = useMemo(() => {
    const d = data?.risk_distribution
    if (!d) return 0
    return (d.low ?? 0) + (d.medium ?? 0) + (d.high ?? 0) + (d.critical ?? 0)
  }, [data])

  const ranges: { value: TimeRange; labelKey: string }[] = [
    { value: 'today', labelKey: 'aiSecurity.timeRanges.today' },
    { value: 'week', labelKey: 'aiSecurity.timeRanges.week' },
    { value: 'month', labelKey: 'aiSecurity.timeRanges.month' },
    { value: 'custom', labelKey: 'aiSecurity.timeRanges.custom' },
  ]

  return (
    <AISecurityLayout>
      <div className='space-y-4'>
        <div className='flex flex-wrap items-center gap-3'>
          <Select value={range} onValueChange={(v) => setRange(v as TimeRange)}>
            <SelectTrigger className='w-32'>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {ranges.map((r) => (
                <SelectItem key={r.value} value={r.value}>
                  {t(r.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {range === 'custom' && (
            <>
              <Input
                className='w-40'
                type='date'
                value={customStart}
                onChange={(e) => setCustomStart(e.target.value)}
                aria-label={t('aiSecurity.startTime')}
              />
              <Input
                className='w-40'
                type='date'
                value={customEnd}
                onChange={(e) => setCustomEnd(e.target.value)}
                aria-label={t('aiSecurity.endTime')}
              />
            </>
          )}

          <Select value={groupFilter} onValueChange={setGroupFilter}>
            <SelectTrigger className='w-40'>
              <SelectValue placeholder={t('aiSecurity.group')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='0'>{t('aiSecurity.allGroups')}</SelectItem>
              {groups.map((g) => (
                <SelectItem key={g.id} value={String(g.id)}>
                  {g.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={ruleFilter} onValueChange={setRuleFilter}>
            <SelectTrigger className='w-40'>
              <SelectValue placeholder={t('aiSecurity.rule')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='0'>{t('aiSecurity.allRules')}</SelectItem>
              {rules.map((r) => (
                <SelectItem key={r.id} value={String(r.id)}>
                  {r.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Input
            className='w-36'
            type='number'
            placeholder={t('aiSecurity.filterByUser')}
            value={userFilter}
            onChange={(e) => setUserFilter(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') loadDashboard()
            }}
          />
          <Button variant='outline' onClick={loadDashboard}>
            {t('aiSecurity.apply')}
          </Button>
        </div>

        {loading ? (
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4'>
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className='h-28 w-full' />
            ))}
          </div>
        ) : (
          <>
            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4'>
              <StatCard
                icon={ShieldCheck}
                title={t('aiSecurity.totalDetections')}
                value={summary.total_detections}
                sub={t('aiSecurity.todayDetections', { count: summary.today_detections })}
              />
              <StatCard
                icon={ShieldX}
                title={t('aiSecurity.totalInterceptions')}
                value={summary.total_interceptions}
                sub={t('aiSecurity.todayInterceptions', { count: summary.today_interceptions })}
              />
              <StatCard
                icon={ShieldAlert}
                title={t('aiSecurity.totalAlerts')}
                value={summary.total_alerts}
              />
              <StatCard icon={BarChart3} title={t('aiSecurity.riskDistribution')} value={riskTotal} />
            </div>

            <Card>
              <CardHeader>
                <CardTitle className='text-base'>{t('aiSecurity.riskDistribution')}</CardTitle>
              </CardHeader>
              <CardContent>
                <div className='grid grid-cols-2 gap-3 sm:grid-cols-4'>
                  {riskLevelOptions.map((opt) => {
                    const dist = data?.risk_distribution
                    const byValue: Record<number, number> = {
                      1: dist?.low ?? 0,
                      2: dist?.medium ?? 0,
                      3: dist?.high ?? 0,
                      4: dist?.critical ?? 0,
                    }
                    const count = byValue[opt.value] ?? 0
                    return (
                      <div key={opt.value} className='rounded-lg border p-3'>
                        <span className={`rounded px-2 py-0.5 text-xs font-medium ${opt.color}`}>
                          {t(opt.labelKey)}
                        </span>
                        <p className='mt-2 text-2xl font-semibold'>{count}</p>
                      </div>
                    )
                  })}
                </div>
              </CardContent>
            </Card>

            <div className='grid grid-cols-1 gap-4 lg:grid-cols-2'>
              <Card>
                <CardHeader>
                  <CardTitle className='text-base'>{t('aiSecurity.topCategories')}</CardTitle>
                </CardHeader>
                <CardContent className='p-0'>
                  <SimpleTable
                    headers={[t('aiSecurity.category'), t('aiSecurity.count')]}
                    rows={data?.top_categories?.map((item) => [item.category, item.count]) ?? []}
                    emptyText={t('aiSecurity.noData')}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className='text-base'>{t('aiSecurity.topModels')}</CardTitle>
                </CardHeader>
                <CardContent className='p-0'>
                  <SimpleTable
                    headers={[t('aiSecurity.model'), t('aiSecurity.count')]}
                    rows={data?.top_models?.map((item) => [item.model_name, item.count]) ?? []}
                    emptyText={t('aiSecurity.noData')}
                  />
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader>
                <CardTitle className='text-base'>{t('aiSecurity.topUsers')}</CardTitle>
              </CardHeader>
              <CardContent className='p-0'>
                <SimpleTable
                  headers={[t('aiSecurity.userId'), t('aiSecurity.userName'), t('aiSecurity.count')]}
                  rows={
                    data?.top_users?.map((item) => [item.user_id, item.user_name ?? '-', item.count]) ?? []
                  }
                  emptyText={t('aiSecurity.noData')}
                />
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </AISecurityLayout>
  )
}

function StatCard(props: { icon: React.ElementType; title: string; value: number; sub?: string }) {
  const Icon = props.icon
  return (
    <Card>
      <CardContent className='flex items-center gap-4 p-6'>
        <div className='flex size-12 items-center justify-center rounded-lg bg-primary/10 text-primary'>
          <Icon className='size-6' />
        </div>
        <div>
          <p className='text-muted-foreground text-sm'>{props.title}</p>
          <p className='text-2xl font-semibold'>{props.value}</p>
          {props.sub && <p className='text-muted-foreground text-xs'>{props.sub}</p>}
        </div>
      </CardContent>
    </Card>
  )
}

function SimpleTable(props: { headers: string[]; rows: (string | number)[][]; emptyText: string }) {
  if (props.rows.length === 0) {
    return <div className='p-6 text-sm text-muted-foreground'>{props.emptyText}</div>
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {props.headers.map((h, i) => (
            <TableHead key={i}>{h}</TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {props.rows.map((row, i) => (
          <TableRow key={i}>
            {row.map((cell, j) => (
              <TableCell key={j}>{cell}</TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
