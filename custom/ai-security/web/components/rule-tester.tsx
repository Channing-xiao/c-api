import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { aiSecurityApi, type AISecurityRuleTestResult } from '../api/ai-security'
import { actionOptions, riskLevelOptions } from '../constants'

interface RuleTesterProps {
  ruleId: number
  ruleName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RuleTester(props: RuleTesterProps) {
  const { t } = useTranslation()
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<AISecurityRuleTestResult | null>(null)

  const handleTest = async () => {
    if (!content.trim()) return
    setLoading(true)
    try {
      const res: any = await aiSecurityApi.testRule(props.ruleId, content)
      if (res.success) setResult(res.data)
    } finally {
      setLoading(false)
    }
  }

  const action = actionOptions.find((a) => a.value === result?.action)
  const risk = riskLevelOptions.find((r) => r.value === result?.risk_level)

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <DialogHeader>
          <DialogTitle>{t('aiSecurity.testRule')}: {props.ruleName}</DialogTitle>
        </DialogHeader>

        <div className='space-y-4 py-4'>
          <div className='space-y-2'>
            <Label htmlFor='test-content'>{t('aiSecurity.testContent')}</Label>
            <textarea
              id='test-content'
              value={content}
              onChange={(e) => setContent(e.target.value)}
              rows={4}
              className='border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:bg-input/30 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 w-full rounded-lg border bg-transparent px-2.5 py-1 text-base transition-colors outline-none focus-visible:ring-3 md:text-sm'
            />
          </div>

          <Button onClick={handleTest} disabled={loading || !content.trim()} className='w-full'>
            {loading ? t('aiSecurity.testing') : t('aiSecurity.runTest')}
          </Button>

          {result && (
            <div className='rounded-xl border bg-muted/30 p-4 space-y-3'>
              <div className='flex flex-wrap items-center gap-2'>
                <span className='text-sm font-medium'>{t('aiSecurity.result')}: </span>
                {result.detected ? (
                  <Badge variant='destructive'>{t('aiSecurity.detected')}</Badge>
                ) : (
                  <Badge variant='outline'>{t('aiSecurity.notDetected')}</Badge>
                )}
                {action && <Badge className={action.color}>{t(action.labelKey)}</Badge>}
                {risk && (
                  <span className={`rounded px-2 py-0.5 text-xs font-medium ${risk.color}`}>
                    {t(risk.labelKey)}
                  </span>
                )}
              </div>

              {result.processed_content && (
                <div className='space-y-1'>
                  <span className='text-xs font-medium text-muted-foreground'>
                    {t('aiSecurity.processedContent')}
                  </span>
                  <pre className='rounded bg-muted p-2 text-xs'>{result.processed_content}</pre>
                </div>
              )}

              {result.matches.length > 0 && (
                <div className='space-y-1'>
                  <span className='text-xs font-medium text-muted-foreground'>{t('aiSecurity.matches')}</span>
                  <ul className='space-y-1'>
                    {result.matches.map((m, idx) => (
                      <li key={idx} className='rounded bg-muted p-2 text-xs'>
                        {m.matched_text}
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={() => props.onOpenChange(false)}>
            {t('aiSecurity.close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
