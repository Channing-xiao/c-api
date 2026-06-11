import { useTranslation } from 'react-i18next'
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
      <div className="flex items-center justify-between px-6 py-4">
        <h1 className="text-2xl font-bold">
          {title ?? t('Content Security')}
        </h1>
        {actions && <div className="flex items-center gap-2">{actions}</div>}
      </div>
      <SecurityTabs />
      <div className="flex-1 overflow-auto p-6">{children}</div>
    </div>
  )
}
