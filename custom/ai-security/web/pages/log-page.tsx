import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Download, ScrollText } from 'lucide-react'
import { toast } from 'sonner'
import {
  aiSecurityApi,
  type AISecurityGroup,
  type AISecurityHitLog,
  type AISecurityRule,
} from '../api/ai-security'
import { AISecurityLayout } from '../components/ai-security-layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { EmptyState } from '@/components/empty-state'
import { LogDetailDrawer } from '../components/log-detail-drawer'
import { actionOptions, riskLevelOptions } from '../constants'

export function AISecurityLogPage() {
  const { t } = useTranslation()
  const [logs, setLogs] = useState<AISecurityHitLog[]>([])
  const [rules, setRules] = useState<AISecurityRule[]>([])
  const [groups, setGroups] = useState<AISecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [actionFilter, setActionFilter] = useState('0')
  const [riskFilter, setRiskFilter] = useState('0')
  const [ruleFilter, setRuleFilter] = useState('0')
  const [groupFilter, setGroupFilter] = useState('0')
  const [modelFilter, setModelFilter] = useState('')
  const [userFilter, setUserFilter] = useState('')
  const [selectedLog, setSelectedLog] = useState<AISecurityHitLog | null>(null)
  const [drawerOpen, setDrawerOpen] = useState(false)

  useEffect(() => {
    aiSecurityApi.getRules({ page: 1, page_size: 200 }).then((res: any) => {
      if (res.success) setRules(res.data.items)
    })
    aiSecurityApi.getGroups({ page: 1, page_size: 200 }).then((res: any) => {
      if (res.success) setGroups(res.data.items)
    })
  }, [])

  // 离散下拉（动作/风险/规则/分组）变更即重载；模型/用户文本框通过回车或查询按钮触发
  useEffect(() => {
    loadLogs()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [actionFilter, riskFilter, ruleFilter, groupFilter])

  const loadLogs = () => {
    setLoading(true)
    aiSecurityApi
      .getLogs({
        page: 1,
        page_size: 100,
        action: Number(actionFilter) || undefined,
        risk_level: Number(riskFilter) || undefined,
        rule_id: Number(ruleFilter) || undefined,
        group_id: Number(groupFilter) || undefined,
        model_name: modelFilter.trim() || undefined,
        user_id: Number(userFilter) || undefined,
      })
      .then((res: any) => {
        if (res.success) setLogs(res.data.items)
      })
      .finally(() => setLoading(false))
  }

  const currentFilters = () => ({
    action: Number(actionFilter) || undefined,
    risk_level: Number(riskFilter) || undefined,
    rule_id: Number(ruleFilter) || undefined,
    group_id: Number(groupFilter) || undefined,
    model_name: modelFilter.trim() || undefined,
    user_id: Number(userFilter) || undefined,
  })

  const handleExport = async () => {
    try {
      const res = await aiSecurityApi.exportLogs({ format: 'csv', ...currentFilters() })
      const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' })
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'ai_security_logs.csv'
      a.click()
      window.URL.revokeObjectURL(url)
    } catch {
      toast.error(t('aiSecurity.exportFailed'))
    }
  }

  const openDetail = (log: AISecurityHitLog) => {
    setSelectedLog(log)
    setDrawerOpen(true)
  }

  const getLabel = (options: { value: number; labelKey: string }[], value: number) =>
    options.find((o) => o.value === value)?.labelKey ?? String(value)

  const ruleName = (id?: number) => (id ? rules.find((r) => r.id === id)?.name ?? `#${id}` : '-')
  const groupName = (id?: number) => (id ? groups.find((g) => g.id === id)?.name ?? `#${id}` : '-')

  const formatTime = (ts: number) => new Date(ts * 1000).toLocaleString()

  const onTextKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') loadLogs()
  }

  return (
    <AISecurityLayout
      actions={
        <Button variant='outline' onClick={handleExport}>
          <Download className='mr-1.5 size-4' />
          {t('aiSecurity.export')}
        </Button>
      }
    >
      <div className='space-y-4'>
        <div className='flex flex-wrap items-center gap-3'>
          <Select value={actionFilter} onValueChange={setActionFilter}>
            <SelectTrigger className='w-36'>
              <SelectValue placeholder={t('aiSecurity.action')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='0'>{t('aiSecurity.allActions')}</SelectItem>
              {actionOptions.map((opt) => (
                <SelectItem key={opt.value} value={String(opt.value)}>
                  {t(opt.labelKey)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <Select value={riskFilter} onValueChange={setRiskFilter}>
            <SelectTrigger className='w-36'>
              <SelectValue placeholder={t('aiSecurity.riskLevel')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='0'>{t('aiSecurity.allRiskLevels')}</SelectItem>
              {riskLevelOptions.map((opt) => (
                <SelectItem key={opt.value} value={String(opt.value)}>
                  {t(opt.labelKey)}
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

          <Input
            className='w-40'
            placeholder={t('aiSecurity.filterByModel')}
            value={modelFilter}
            onChange={(e) => setModelFilter(e.target.value)}
            onKeyDown={onTextKeyDown}
          />
          <Input
            className='w-36'
            type='number'
            placeholder={t('aiSecurity.filterByUser')}
            value={userFilter}
            onChange={(e) => setUserFilter(e.target.value)}
            onKeyDown={onTextKeyDown}
          />
          <Button variant='outline' onClick={loadLogs}>
            {t('aiSecurity.apply')}
          </Button>
        </div>

        <Card>
          <CardContent className='p-0'>
            {loading ? (
              <div className='space-y-3 p-6'>
                <Skeleton className='h-8 w-full' />
                <Skeleton className='h-8 w-full' />
              </div>
            ) : logs.length === 0 ? (
              <EmptyState
                icon={ScrollText}
                title={t('aiSecurity.noLogs')}
                description={t('aiSecurity.noLogsDesc')}
                className='min-h-[260px]'
                bordered={false}
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('aiSecurity.time')}</TableHead>
                    <TableHead>{t('aiSecurity.user')}</TableHead>
                    <TableHead>{t('aiSecurity.model')}</TableHead>
                    <TableHead>{t('aiSecurity.action')}</TableHead>
                    <TableHead>{t('aiSecurity.rule')}</TableHead>
                    <TableHead>{t('aiSecurity.group')}</TableHead>
                    <TableHead>{t('aiSecurity.riskLevel')}</TableHead>
                    <TableHead>{t('aiSecurity.score')}</TableHead>
                    <TableHead className='text-right'>{t('aiSecurity.actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => {
                    const risk = riskLevelOptions.find((r) => r.value === log.risk_level)
                    return (
                      <TableRow key={log.id}>
                        <TableCell className='text-muted-foreground text-sm'>{formatTime(log.created_at)}</TableCell>
                        <TableCell>{log.user_name ?? log.user_id}</TableCell>
                        <TableCell className='text-muted-foreground'>{log.model_name}</TableCell>
                        <TableCell>
                          <Badge>{t(getLabel(actionOptions, log.action))}</Badge>
                        </TableCell>
                        <TableCell className='text-muted-foreground'>{ruleName(log.rule_id)}</TableCell>
                        <TableCell className='text-muted-foreground'>{groupName(log.group_id)}</TableCell>
                        <TableCell>
                          {risk && (
                            <span className={`rounded px-2 py-0.5 text-xs font-medium ${risk.color}`}>
                              {t(risk.labelKey)}
                            </span>
                          )}
                        </TableCell>
                        <TableCell>{log.risk_score}</TableCell>
                        <TableCell className='text-right'>
                          <Button variant='outline' size='sm' onClick={() => openDetail(log)}>
                            {t('aiSecurity.detail')}
                          </Button>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <LogDetailDrawer log={selectedLog} open={drawerOpen} onOpenChange={setDrawerOpen} />
      </div>
    </AISecurityLayout>
  )
}
