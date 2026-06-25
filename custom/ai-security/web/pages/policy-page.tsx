import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Shield, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { aiSecurityApi, type AISecurityGroup, type AISecurityPolicy } from '../api/ai-security'
import { AISecurityLayout } from '../components/ai-security-layout'
import { PolicyFormModal } from '../components/policy-form-modal'
import { actionOptions, scopeOptions } from '../constants'

export function AISecurityPolicyPage() {
  const { t } = useTranslation()
  const [policies, setPolicies] = useState<AISecurityPolicy[]>([])
  const [groups, setGroups] = useState<AISecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingPolicy, setEditingPolicy] = useState<AISecurityPolicy | null>(null)

  useEffect(() => {
    loadPolicies()
    aiSecurityApi.getGroups({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) setGroups(res.data.items)
    })
  }, [])

  const loadPolicies = () => {
    setLoading(true)
    aiSecurityApi
      .getPolicies({ page: 1, page_size: 100 })
      .then((res: any) => {
        if (res.success) setPolicies(res.data.items)
      })
      .finally(() => setLoading(false))
  }

  const groupName = (id: number) => groups.find((g) => g.id === id)?.name ?? '-'

  const getLabel = (options: { value: number; labelKey: string }[], value: number) => options.find((o) => o.value === value)?.labelKey ?? String(value)

  const handleDelete = async (id: number) => {
    if (!confirm(t('aiSecurity.confirmDeletePolicy'))) return
    try {
      await aiSecurityApi.deletePolicy(id)
      toast.success(t('aiSecurity.policyDeleted'))
      loadPolicies()
    } catch {
      toast.error(t('aiSecurity.policyDeleteFailed'))
    }
  }

  const handleCreate = () => {
    setEditingPolicy(null)
    setModalOpen(true)
  }

  const handleEdit = (policy: AISecurityPolicy) => {
    setEditingPolicy(policy)
    setModalOpen(true)
  }

  const handleSubmit = async (data: Partial<AISecurityPolicy>) => {
    try {
      if (editingPolicy) {
        await aiSecurityApi.updatePolicy(editingPolicy.id, data)
        toast.success(t('aiSecurity.policyUpdated'))
      } else {
        await aiSecurityApi.createPolicy(data)
        toast.success(t('aiSecurity.policyCreated'))
      }
      loadPolicies()
    } catch {
      toast.error(t('aiSecurity.policySaveFailed'))
    }
  }

  if (loading) {
    return (
      <AISecurityLayout>
        <div className='space-y-4'>
          <Skeleton className='h-8 w-48' />
          <Skeleton className='h-32 w-full' />
        </div>
      </AISecurityLayout>
    )
  }

  return (
    <AISecurityLayout
      actions={
        <Button onClick={handleCreate}>
          <Plus className='mr-1.5 size-4' />
          {t('aiSecurity.createPolicy')}
        </Button>
      }
    >
      <div className='space-y-4'>
        <Card>
          <CardContent className='p-0'>
            {policies.length === 0 ? (
              <EmptyState
                icon={Shield}
                title={t('aiSecurity.noPolicies')}
                description={t('aiSecurity.noPoliciesDesc')}
                action={
                  <Button onClick={handleCreate}>
                    <Plus className='mr-1.5 size-4' />
                    {t('aiSecurity.createPolicy')}
                  </Button>
                }
                className='min-h-[260px]'
                bordered={false}
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('aiSecurity.user')}</TableHead>
                    <TableHead>{t('aiSecurity.group')}</TableHead>
                    <TableHead>{t('aiSecurity.scope')}</TableHead>
                    <TableHead>{t('aiSecurity.defaultAction')}</TableHead>
                    <TableHead>{t('aiSecurity.priority')}</TableHead>
                    <TableHead>{t('aiSecurity.status')}</TableHead>
                    <TableHead className='text-right'>{t('aiSecurity.actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {policies.map((policy) => (
                    <TableRow key={policy.id}>
                      <TableCell className='font-medium'>{policy.user_name ?? policy.user_id}</TableCell>
                      <TableCell className='text-muted-foreground'>{groupName(policy.group_id)}</TableCell>
                      <TableCell>
                        <Badge variant='outline'>{t(getLabel(scopeOptions, policy.scope))}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge>{t(getLabel(actionOptions, policy.default_action))}</Badge>
                      </TableCell>
                      <TableCell>{policy.priority}</TableCell>
                      <TableCell>
                        {policy.status === 1 ? t('aiSecurity.enabled') : t('aiSecurity.disabled')}
                      </TableCell>
                      <TableCell className='text-right space-x-2'>
                        <Button variant='outline' size='sm' onClick={() => handleEdit(policy)}>
                          {t('aiSecurity.edit')}
                        </Button>
                        <Button variant='destructive' size='sm' onClick={() => handleDelete(policy.id)}>
                          {t('aiSecurity.delete')}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
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
    </AISecurityLayout>
  )
}
