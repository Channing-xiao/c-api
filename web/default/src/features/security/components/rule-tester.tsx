import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { securityApi, type RuleTestResult } from '../api/security'

interface RuleTesterProps {
  ruleId: number
  ruleName: string
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function RuleTester({ ruleId, ruleName, open, onOpenChange }: RuleTesterProps) {
  const { t } = useTranslation()
  const [content, setContent] = useState('')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<RuleTestResult | null>(null)

  const handleTest = async () => {
    if (!content.trim()) return
    setLoading(true)
    try {
      const res: any = await securityApi.testRule(ruleId, content.trim())
      if (res.success) {
        setResult(res.data)
      }
    } finally {
      setLoading(false)
    }
  }

  const handleClose = () => {
    setContent('')
    setResult(null)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('Test Rule')}: {ruleName}</DialogTitle>
          <DialogDescription>
            {t('Enter content to test against this rule.')}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <Textarea
            placeholder={t('Enter test content...')}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={4}
          />

          {result && (
            <div className="rounded-lg border p-4 space-y-2">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">{t('Detected')}:</span>
                <Badge variant={result.detected ? 'destructive' : 'outline'}>
                  {result.detected ? t('Yes') : t('No')}
                </Badge>
              </div>
              {result.detected && (
                <>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{t('Action')}:</span>
                    <span className="text-sm">{result.action_name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium">{t('Risk Score')}:</span>
                    <span className="text-sm">{result.risk_score}</span>
                  </div>
                  {result.matches?.length > 0 && (
                    <div className="text-sm">
                      <span className="font-medium">{t('Matches')}:</span>{' '}
                      {result.matches.map((m) => m.matched_text).join(', ')}
                    </div>
                  )}
                  {result.processed_content && (
                    <div className="text-sm">
                      <span className="font-medium">{t('Processed')}:</span>{' '}
                      {result.processed_content}
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>{t('Close')}</Button>
          <Button onClick={handleTest} disabled={loading || !content.trim()}>
            {loading ? t('Testing...') : t('Test')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
