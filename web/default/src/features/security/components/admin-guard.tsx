import { useTranslation } from 'react-i18next'
import { EmptyState } from '@/components/empty-state'

interface AdminGuardProps {
  children: React.ReactNode
  role?: string
}

export function AdminGuard({ children, role = 'admin' }: AdminGuardProps) {
  const { t } = useTranslation()
  // 从全局状态获取当前用户信息
  const user = (window as any).__USER__ as { role?: string } | undefined

  if (user?.role !== role) {
    return (
      <EmptyState
        title={t('Access Denied')}
        description={t('You do not have permission to access this page.')}
      />
    )
  }

  return <>{children}</>
}
