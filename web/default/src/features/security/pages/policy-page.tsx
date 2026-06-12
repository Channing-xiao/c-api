import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { securityApi, type SecurityGroup, type SecurityPolicy } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
import { PolicyFormModal } from '../components/policy-form-modal'
import { useSecurityOptions } from '../constants'

export function SecurityPolicyPage() {
  const { t } = useTranslation()
  const { getLabel, scopes, actions } = useSecurityOptions()
  const [policies, setPolicies] = useState<SecurityPolicy[]>([])
  const [groups, setGroups] = useState<SecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingPolicy, setEditingPolicy] = useState<SecurityPolicy | null>(null)

  useEffect(() => {
    loadPolicies()
    securityApi.getGroups({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) setGroups(res.data.items)
    })
  }, [])

  const loadPolicies = () => {
    setLoading(true)
    securityApi.getPolicies({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) {
        setPolicies(res.data.items)
      }
      setLoading(false)
    })
  }

  const handleDelete = async (id: number) => {
    if (!confirm(t('Are you sure?'))) return
    try {
      await securityApi.deletePolicy(id)
      toast.success(t('Policy deleted'))
      loadPolicies()
    } catch {
      toast.error(t('Failed to delete policy'))
    }
  }

  const handleCreate = () => {
    setEditingPolicy(null)
    setModalOpen(true)
  }

  const handleEdit = (policy: SecurityPolicy) => {
    setEditingPolicy(policy)
    setModalOpen(true)
  }

  const handleSubmit = async (data: Partial<SecurityPolicy>) => {
    try {
      if (editingPolicy) {
        await securityApi.updatePolicy(editingPolicy.id, data)
        toast.success(t('Policy updated'))
      } else {
        await securityApi.createPolicy(data)
        toast.success(t('Policy created'))
      }
      loadPolicies()
    } catch {
      toast.error(t('Failed to save policy'))
    }
  }

  if (loading) {
    return (
      <SecurityPageLayout>
        <div className="space-y-4 p-6">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-32 w-full" />
        </div>
      </SecurityPageLayout>
    )
  }

  return (
    <SecurityPageLayout
      actions={<Button onClick={handleCreate}>{t('Create Policy')}</Button>}
    >
      <div className="space-y-4">
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('User')}</TableHead>
                <TableHead>{t('Group')}</TableHead>
                <TableHead>{t('Scope')}</TableHead>
                <TableHead>{t('Default Action')}</TableHead>
                <TableHead>{t('Priority')}</TableHead>
                <TableHead className="text-right">{t('Actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {policies.map((policy) => (
                <TableRow key={policy.id}>
                  <TableCell className="font-medium">{policy.user_name}</TableCell>
                  <TableCell>{policy.group_name}</TableCell>
                  <TableCell><Badge variant="outline">{getLabel(scopes, policy.scope)}</Badge></TableCell>
                  <TableCell><Badge>{getLabel(actions, policy.default_action)}</Badge></TableCell>
                  <TableCell>{policy.priority ?? 0}</TableCell>
                  <TableCell className="text-right space-x-2">
                    <Button variant="outline" size="sm" onClick={() => handleEdit(policy)}>{t('Edit')}</Button>
                    <Button variant="destructive" size="sm" onClick={() => handleDelete(policy.id)}>
                      {t('Delete')}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <PolicyFormModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        initialData={editingPolicy}
        groups={groups}
        onSubmit={handleSubmit}
      />
      </div>
    </SecurityPageLayout>
  )
}
