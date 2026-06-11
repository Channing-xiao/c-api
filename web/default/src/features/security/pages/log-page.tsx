import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { securityApi, type SecurityHitLog } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
import { LogDetailDrawer } from '../components/log-detail-drawer'

const POLLING_TIMEOUT_MS = 30 * 60 * 1000 // 30 minutes

const actionMap: Record<number, string> = {
  1: 'Pass',
  2: 'Alert',
  3: 'Mask',
  4: 'Block',
  5: 'Review',
}

const riskLevelMap: Record<number, { label: string; color: string }> = {
  1: { label: 'Low', color: 'bg-green-100 text-green-800' },
  2: { label: 'Medium', color: 'bg-yellow-100 text-yellow-800' },
  3: { label: 'High', color: 'bg-orange-100 text-orange-800' },
  4: { label: 'Critical', color: 'bg-red-100 text-red-800' },
}

export function SecurityLogPage() {
  const { t } = useTranslation()
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

  if (loading) {
    return (
      <SecurityPageLayout>
        <div className="space-y-4 p-6">
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
            {autoRefresh ? t('Auto Refresh: ON') : t('Auto Refresh: OFF')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport('csv')}>
            {t('Export CSV')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => handleExport('excel')}>
            {t('Export Excel')}
          </Button>
        </>
      }
    >
      <div className="space-y-4">
        <div className="flex items-center gap-2 flex-wrap">
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
          <Button variant="outline" size="sm" onClick={() => {
            setModelName('')
            setStartTime('')
            setEndTime('')
            securityApi.getLogs({ page: 1, page_size: 100 }).then((res: any) => {
              if (res.success) setLogs(res.data.items)
            })
          }}>
            {t('Reset')}
          </Button>
        </div>
      <Card>
        <CardContent className="p-0">
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
                  <TableCell>{new Date(log.created_at * 1000).toLocaleString()}</TableCell>
                  <TableCell>{log.user_name}</TableCell>
                  <TableCell>{log.model_name}</TableCell>
                  <TableCell><Badge>{actionMap[log.action] ?? log.action}</Badge></TableCell>
                  <TableCell>
                    {(() => {
                      const risk = riskLevelMap[log.risk_level]
                      return risk ? <span className={`px-2 py-1 rounded text-xs font-medium ${risk.color}`}>{risk.label}</span> : log.risk_level
                    })()}
                  </TableCell>
                  <TableCell>{log.risk_score}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
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
