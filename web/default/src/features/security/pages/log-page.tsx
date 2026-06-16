import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Download,
  FileSpreadsheet,
  Filter,
  RefreshCw,
  ScrollText,
} from 'lucide-react'
import { toast } from 'sonner'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { securityApi, type SecurityHitLog } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
import { LogDetailDrawer } from '../components/log-detail-drawer'
import { useSecurityOptions } from '../constants'

const POLLING_TIMEOUT_MS = 30 * 60 * 1000 // 30 minutes

export function SecurityLogPage() {
  const { t } = useTranslation()
  const { getLabel, actions, getRiskLevel } = useSecurityOptions()
  const [logs, setLogs] = useState<SecurityHitLog[]>([])
  const [loading, setLoading] = useState(true)
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [selectedLog, setSelectedLog] = useState<SecurityHitLog | null>(null)
  const [modelName, setModelName] = useState('')
  const [startTime, setStartTime] = useState('')
  const [endTime, setEndTime] = useState('')
  const [autoRefresh, setAutoRefresh] = useState(false)
  const pollingStartTime = useRef<number | null>(null)

  useEffect(() => {
    loadLogs()
    let interval: ReturnType<typeof setInterval>

    const handleVisibility = () => {
      if (document.hidden) {
        clearInterval(interval)
      } else if (autoRefresh) {
        interval = setInterval(() => loadLogs(), 5000)
      }
    }

    if (autoRefresh) {
      if (pollingStartTime.current == null) {
        pollingStartTime.current = Date.now()
      }
      interval = setInterval(() => {
        if (pollingStartTime.current && Date.now() - pollingStartTime.current > POLLING_TIMEOUT_MS) {
          setAutoRefresh(false)
          pollingStartTime.current = null
          toast.info(t('Auto-refresh paused after 30 minutes'))
          clearInterval(interval)
          return
        }
        loadLogs()
      }, 5000)
    } else {
      pollingStartTime.current = null
    }

    document.addEventListener('visibilitychange', handleVisibility)
    return () => {
      clearInterval(interval)
      document.removeEventListener('visibilitychange', handleVisibility)
    }
  }, [autoRefresh])

  const loadLogs = () => {
    const params: Record<string, any> = { page: 1, page_size: 100 }
    if (modelName) params.model_name = modelName
    if (startTime) params.start_time = Math.floor(new Date(startTime).getTime() / 1000)
    if (endTime) params.end_time = Math.floor(new Date(endTime).getTime() / 1000)

    securityApi.getLogs(params).then((res: any) => {
      if (res.success) {
        setLogs(res.data.items)
      }
      setLoading(false)
    })
  }

  const handleExport = async (format: 'csv' | 'excel') => {
    try {
      const params: Record<string, any> = { format }
      if (modelName) params.model_name = modelName
      if (startTime) params.start_time = Math.floor(new Date(startTime).getTime() / 1000)
      if (endTime) params.end_time = Math.floor(new Date(endTime).getTime() / 1000)
      const res = await securityApi.exportLogs(params)
      const blob = new Blob([res.data], {
        type:
          format === 'excel'
            ? 'application/vnd.ms-excel'
            : 'text/csv;charset=utf-8;',
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `security_logs_${new Date().toISOString().slice(0, 10)}.${format === 'excel' ? 'xls' : 'csv'}`
      document.body.appendChild(a)
      a.click()
      a.remove()
      URL.revokeObjectURL(url)
    } catch {
      toast.error(t('Export failed'))
    }
  }

  const openDrawer = (log: SecurityHitLog) => {
    setSelectedLog(log)
    setDrawerOpen(true)
  }

  const resetFilters = () => {
    setModelName('')
    setStartTime('')
    setEndTime('')
    securityApi.getLogs({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) setLogs(res.data.items)
    })
  }

  if (loading) {
    return (
      <SecurityPageLayout>
        <div className="space-y-4">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-32 w-full" />
        </div>
      </SecurityPageLayout>
    )
  }

  return (
    <SecurityPageLayout
      actions={
        <>
          <Button
            variant={autoRefresh ? 'default' : 'outline'}
            size="sm"
            onClick={() => setAutoRefresh((v) => !v)}
          >
            <RefreshCw className="mr-1.5 size-3.5" />
            {autoRefresh ? t('Auto Refresh: ON') : t('Auto Refresh: OFF')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport('csv')}>
            <Download className="mr-1.5 size-3.5" />
            {t('Export CSV')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport('excel')}>
            <FileSpreadsheet className="mr-1.5 size-3.5" />
            {t('Export Excel')}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <Card className="border-border/60">
          <CardContent className="flex flex-wrap items-center gap-3 py-4">
            <Filter className="size-4 text-muted-foreground" />
            <Input
              placeholder={t('Model name')}
              value={modelName}
              onChange={(e) => setModelName(e.target.value)}
              className="w-48"
            />
            <Input
              type="datetime-local"
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
              className="w-56"
            />
            <Input
              type="datetime-local"
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
              className="w-56"
            />
            <Button size="sm" onClick={loadLogs}>{t('Apply')}</Button>
            <Button variant="outline" size="sm" onClick={resetFilters}>
              {t('Reset')}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-0">
            {logs.length === 0 ? (
              <EmptyState
                icon={ScrollText}
                title={t('No Logs')}
                description={t('No audit logs match the current filters.')}
                action={
                  <Button variant="outline" size="sm" onClick={resetFilters}>
                    {t('Reset Filters')}
                  </Button>
                }
                className="min-h-[260px]"
                bordered={false}
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('Time')}</TableHead>
                    <TableHead>{t('User')}</TableHead>
                    <TableHead>{t('Model')}</TableHead>
                    <TableHead>{t('Action')}</TableHead>
                    <TableHead>{t('Risk')}</TableHead>
                    <TableHead>{t('Score')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {logs.map((log) => (
                    <TableRow
                      key={log.id}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => openDrawer(log)}
                    >
                      <TableCell className="tabular-nums">{new Date(log.created_at * 1000).toLocaleString()}</TableCell>
                      <TableCell>{log.user_name}</TableCell>
                      <TableCell className="text-muted-foreground">{log.model_name}</TableCell>
                      <TableCell><Badge>{getLabel(actions, log.action)}</Badge></TableCell>
                      <TableCell>
                        {(() => {
                          const risk = getRiskLevel(log.risk_level)
                          return risk ? (
                            <span className={`px-2 py-1 rounded-md text-xs font-medium ${risk.color}`}>
                              {t(risk.labelKey)}
                            </span>
                          ) : (
                            log.risk_level
                          )
                        })()}
                      </TableCell>
                      <TableCell className="tabular-nums">{log.risk_score}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <LogDetailDrawer
          log={selectedLog}
          open={drawerOpen}
          onOpenChange={setDrawerOpen}
        />
      </div>
    </SecurityPageLayout>
  )
}
