import { useTranslation } from 'react-i18next'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Badge } from '@/components/ui/badge'
import type { AISecurityHitLog } from '../api/ai-security'
import { actionOptions, riskLevelOptions } from '../constants'

interface LogDetailDrawerProps {
  log: AISecurityHitLog | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function LogDetailDrawer(props: LogDetailDrawerProps) {
  const { t } = useTranslation()
  const log = props.log

  const getLabel = (options: { value: number; labelKey: string }[], value: number) => options.find((o) => o.value === value)?.labelKey ?? String(value)

  const formatTime = (ts: number) => new Date(ts * 1000).toLocaleString()

  return (
    <Sheet open={props.open} onOpenChange={props.onOpenChange}>
      <SheetContent className='sm:max-w-md'>
        <SheetHeader>
          <SheetTitle>{t('aiSecurity.logDetail')}</SheetTitle>
        </SheetHeader>

        {log && (
          <div className='mt-6 space-y-4 text-sm'>
            <Row label={t('aiSecurity.time')} value={formatTime(log.created_at)} />
            <Row label={t('aiSecurity.userId')} value={log.user_id} />
            <Row label={t('aiSecurity.model')} value={log.model_name} />
            <Row label={t('aiSecurity.ip')} value={log.ip} />
            <Row label={t('aiSecurity.contentHash')} value={log.original_content_hash} />
            <Row
              label={t('aiSecurity.action')}
              value={
                <Badge>{t(getLabel(actionOptions, log.action))}</Badge>
              }
            />
            <Row
              label={t('aiSecurity.riskLevel')}
              value={
                <span
                  className={`rounded px-2 py-0.5 text-xs font-medium ${
                    riskLevelOptions.find((r) => r.value === log.risk_level)?.color ?? ''
                  }`}
                >
                  {t(getLabel(riskLevelOptions, log.risk_level))}
                </span>
              }
            />
            <div className='space-y-1'>
              <span className='text-muted-foreground'>{t('aiSecurity.matchedText')}</span>
              <pre className='rounded bg-muted p-2 text-xs'>{log.matched_text || '-'}</pre>
            </div>
            {log.processed_content && (
              <div className='space-y-1'>
                <span className='text-muted-foreground'>{t('aiSecurity.processedContent')}</span>
                <pre className='rounded bg-muted p-2 text-xs'>{log.processed_content}</pre>
              </div>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}

function Row(props: { label: string; value: React.ReactNode }) {
  return (
    <div className='flex items-start justify-between gap-4'>
      <span className='text-muted-foreground'>{props.label}</span>
      <span className='font-medium'>{props.value}</span>
    </div>
  )
}
