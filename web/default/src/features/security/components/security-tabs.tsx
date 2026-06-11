import { useTranslation } from 'react-i18next'
import { useNavigate, useLocation } from '@tanstack/react-router'
import { cn } from '@/lib/utils'

const tabs = [
  { key: 'dashboard', label: 'Dashboard', path: '/security' },
  { key: 'groups', label: 'Groups', path: '/security/groups' },
  { key: 'rules', label: 'Rules', path: '/security/rules' },
  { key: 'policies', label: 'Policies', path: '/security/policies' },
  { key: 'logs', label: 'Audit Logs', path: '/security/logs' },
]

export function SecurityTabs() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()

  const activeTab =
    tabs.find((tab) => location.pathname.startsWith(tab.path))?.key ?? 'dashboard'

  return (
    <div className="border-b">
      <nav className="flex space-x-1 px-6">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => navigate({ to: tab.path })}
            className={cn(
              'relative px-4 py-3 text-sm font-medium transition-colors',
              activeTab === tab.key
                ? 'text-foreground'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {t(tab.label)}
            {activeTab === tab.key && (
              <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-primary" />
            )}
          </button>
        ))}
      </nav>
    </div>
  )
}
