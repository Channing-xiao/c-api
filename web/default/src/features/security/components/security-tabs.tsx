import { useTranslation } from 'react-i18next'
import { useNavigate, useLocation } from '@tanstack/react-router'
import {
  LayoutDashboard,
  FolderOpen,
  FileText,
  ShieldCheck,
  ScrollText,
} from 'lucide-react'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

const tabs = [
  { key: 'dashboard', label: 'Dashboard', path: '/security', icon: LayoutDashboard },
  { key: 'groups', label: 'Groups', path: '/security/groups', icon: FolderOpen },
  { key: 'rules', label: 'Rules', path: '/security/rules', icon: FileText },
  { key: 'policies', label: 'Policies', path: '/security/policies', icon: ShieldCheck },
  { key: 'logs', label: 'Audit Logs', path: '/security/logs', icon: ScrollText },
]

export function SecurityTabs() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()

  const activeTab =
    tabs.find((tab) => location.pathname.startsWith(tab.path))?.key ?? 'dashboard'

  return (
    <div className="border-b bg-card px-6">
      <Tabs
        value={activeTab}
        onValueChange={(value) => {
          const tab = tabs.find((t) => t.key === value)
          if (tab) navigate({ to: tab.path })
        }}
      >
        <TabsList variant="line" className="h-11 bg-transparent">
          {tabs.map((tab) => {
            const Icon = tab.icon
            return (
              <TabsTrigger
                key={tab.key}
                value={tab.key}
                className="gap-2 px-4 py-2"
              >
                <Icon className="size-4" />
                {t(tab.label)}
              </TabsTrigger>
            )
          })}
        </TabsList>
      </Tabs>
    </div>
  )
}
