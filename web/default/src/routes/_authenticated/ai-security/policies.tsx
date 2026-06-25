import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { ROLE } from '@/lib/roles'
import { AISecurityPolicyPage } from '@custom/ai-security/web/pages/policy-page'

export const Route = createFileRoute('/_authenticated/ai-security/policies')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()
    if (!auth.user || auth.user.role < ROLE.ADMIN) {
      throw redirect({ to: '/403' })
    }
  },
  component: AISecurityPolicyPage,
})
