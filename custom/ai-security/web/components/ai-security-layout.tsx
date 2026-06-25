import { Link, useLocation } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { LayoutDashboard, FolderOpen, FileText, Shield, ScrollText } from 'lucide-react'
import { cn } from '@/lib/utils'
import '../i18n/register'

interface AISecurityLayoutProps {
  children: React.ReactNode
  actions?: React.ReactNode
}

const tabs = [
  { path: '/ai-security', labelKey: 'aiSecurity.tabs.dashboard', icon: LayoutDashboard },
  { path: '/ai-security/groups', labelKey: 'aiSecurity.tabs.groups', icon: FolderOpen },
  { path: '/ai-security/rules', labelKey: 'aiSecurity.tabs.rules', icon: FileText },
  { path: '/ai-security/policies', labelKey: 'aiSecurity.tabs.policies', icon: Shield },
  { path: '/ai-security/logs', labelKey: 'aiSecurity.tabs.logs', icon: ScrollText },
]

export function AISecurityLayout(props: AISecurityLayoutProps) {
  const { t } = useTranslation()
  const location = useLocation()

  return (
    <div className='space-y-4'>
      <div className='flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between'>
        <div>
          <h1 className='text-2xl font-semibold tracking-tight'>{t('aiSecurity.title')}</h1>
          <p className='text-muted-foreground text-sm'>{t('aiSecurity.subtitle')}</p>
        </div>
        {props.actions}
      </div>

      <nav className='border-b'>
        <div className='flex gap-1'>
          {tabs.map((tab) => {
            const Icon = tab.icon
            const active =
              tab.path === '/ai-security'
                ? location.pathname === '/ai-security' || location.pathname === '/ai-security/dashboard'
                : location.pathname.startsWith(tab.path)
            return (
              <Link
                key={tab.path}
                to={tab.path}
                className={cn(
                  'flex items-center gap-2 border-b-2 px-3 py-2 text-sm font-medium transition-colors',
                  active
                    ? 'border-primary text-primary'
                    : 'border-transparent text-muted-foreground hover:text-foreground'
                )}
              >
                <Icon className='size-4' />
                {t(tab.labelKey)}
              </Link>
            )
          })}
        </div>
      </nav>

      <div>{props.children}</div>
    </div>
  )
}
