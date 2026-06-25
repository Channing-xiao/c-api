import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { ROLE } from '@/lib/roles'
import { AISecurityDashboardPage } from '@custom/ai-security/web/pages/dashboard-page'

export const Route = createFileRoute('/_authenticated/ai-security/dashboard')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()
    if (!auth.user || auth.user.role < ROLE.ADMIN) {
      throw redirect({ to: '/403' })
    }
  },
  component: AISecurityDashboardPage,
})
