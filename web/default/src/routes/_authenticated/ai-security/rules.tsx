import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { ROLE } from '@/lib/roles'
import { AISecurityRulePage } from '@custom/ai-security/web/pages/rule-page'

export const Route = createFileRoute('/_authenticated/ai-security/rules')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()
    if (!auth.user || auth.user.role < ROLE.ADMIN) {
      throw redirect({ to: '/403' })
    }
  },
  component: AISecurityRulePage,
})
