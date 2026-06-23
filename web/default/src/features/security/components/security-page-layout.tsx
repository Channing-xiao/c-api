import { useTranslation } from 'react-i18next'
import { Shield } from 'lucide-react'
import { SecurityTabs } from './security-tabs'

interface SecurityPageLayoutProps {
  title?: string
  children: React.ReactNode
  actions?: React.ReactNode
}

export function SecurityPageLayout({
  title,
  children,
  actions,
}: SecurityPageLayoutProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col h-full">
      <div className="relative overflow-hidden border-b bg-card px-6 py-5">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 opacity-30 dark:opacity-20"
          style={{
            background: [
              'radial-gradient(ellipse 60% 120% at 90% 0%, color-mix(in oklch, var(--primary) 10%, transparent) 0%, transparent 55%)',
              'radial-gradient(ellipse 40% 80% at 10% 100%, color-mix(in oklch, var(--primary) 6%, transparent) 0%, transparent 50%)',
            ].join(', '),
          }}
        />
        <div className="relative flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Shield className="size-5" />
            </div>
            <div>
              <h1 className="text-xl font-semibold tracking-tight">
                {title ?? t('Content Security')}
              </h1>
              <p className="text-xs text-muted-foreground">
                {t('AI content security detection and audit')}
              </p>
            </div>
          </div>
          {actions && <div className="flex items-center gap-2">{actions}</div>}
        </div>
      </div>
      <SecurityTabs />
      <div className="flex-1 overflow-auto bg-muted/20 p-6">{children}</div>
    </div>
  )
}
