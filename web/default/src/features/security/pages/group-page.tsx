import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { FolderOpen, Info, Plus } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { EmptyState } from '@/components/empty-state'
import { securityApi, type SecurityGroup } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
import { GroupFormModal } from '../components/group-form-modal'

export function SecurityGroupPage() {
  const { t } = useTranslation()
  const [groups, setGroups] = useState<SecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingGroup, setEditingGroup] = useState<SecurityGroup | null>(null)
  const [migrationDismissed, setMigrationDismissed] = useState(() => {
    return localStorage.getItem('security_migration_dismissed') === '1'
  })

  useEffect(() => {
    loadGroups()
  }, [])

  const loadGroups = () => {
    setLoading(true)
    securityApi.getGroups({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) {
        setGroups(res.data.items)
      }
      setLoading(false)
    })
  }

  const dismissMigration = () => {
    localStorage.setItem('security_migration_dismissed', '1')
    setMigrationDismissed(true)
  }

  const handleDelete = async (id: number) => {
    if (!confirm(t('Are you sure?'))) return
    try {
      await securityApi.deleteGroup(id)
      toast.success(t('Group deleted'))
      loadGroups()
    } catch {
      toast.error(t('Failed to delete group'))
    }
  }

  const handleCreate = () => {
    setEditingGroup(null)
    setModalOpen(true)
  }

  const handleEdit = (group: SecurityGroup) => {
    setEditingGroup(group)
    setModalOpen(true)
  }

  const handleSubmit = async (data: Partial<SecurityGroup>) => {
    try {
      if (editingGroup) {
        await securityApi.updateGroup(editingGroup.id, data)
        toast.success(t('Group updated'))
      } else {
        await securityApi.createGroup(data)
        toast.success(t('Group created'))
      }
      loadGroups()
    } catch {
      toast.error(t('Failed to save group'))
    }
  }

  const handleToggleStatus = async (group: SecurityGroup) => {
    const newStatus = group.status === 1 ? 0 : 1
    try {
      await securityApi.updateGroupStatus(group.id, newStatus)
      toast.success(t(newStatus === 1 ? 'Group enabled' : 'Group disabled'))
      loadGroups()
    } catch {
      toast.error(t('Failed to update group status'))
    }
  }

  if (loading) {
    return (
      <SecurityPageLayout>
        <div className="space-y-4">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-32 w-full" />
        </div>
      </SecurityPageLayout>
    )
  }

  return (
    <SecurityPageLayout
      actions={
        <Button onClick={handleCreate}>
          <Plus className="mr-1.5 size-4" />
          {t('Create Group')}
        </Button>
      }
    >
      <div className="space-y-4">
        {!migrationDismissed && (
          <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/30 p-4">
            <div className="flex items-start gap-3">
              <div className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <Info className="size-4" />
              </div>
              <div className="text-sm">
                <span className="font-medium">{t('System Migration')}</span>
                <p className="text-muted-foreground">
                  {t('The legacy Sensitive Words system has been consolidated into the new AI Content Security module.')}
                </p>
              </div>
            </div>
            <Button variant="ghost" size="sm" onClick={dismissMigration}>
              {t('Dismiss')}
            </Button>
          </div>
        )}

        <Card>
          <CardContent className="p-0">
            {groups.length === 0 ? (
              <EmptyState
                icon={FolderOpen}
                title={t('No Groups')}
                description={t('Create your first rule group to organize security rules.')}
                action={
                  <Button onClick={handleCreate}>
                    <Plus className="mr-1.5 size-4" />
                    {t('Create Group')}
                  </Button>
                }
                className="min-h-[260px]"
                bordered={false}
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('Name')}</TableHead>
                    <TableHead>{t('Description')}</TableHead>
                    <TableHead>{t('Status')}</TableHead>
                    <TableHead className="text-right">{t('Actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {groups.map((group) => (
                    <TableRow key={group.id}>
                      <TableCell className="font-medium">{group.name}</TableCell>
                      <TableCell className="text-muted-foreground">{group.description}</TableCell>
                      <TableCell>
                        <div className='flex items-center gap-2'>
                          <Switch
                            checked={group.status === 1}
                            onCheckedChange={() => handleToggleStatus(group)}
                            size='sm'
                          />
                          <span className='text-sm'>
                            {group.status === 1 ? t('Enabled') : t('Disabled')}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right space-x-2">
                        <Button variant="outline" size="sm" onClick={() => handleEdit(group)}>{t('Edit')}</Button>
                        <Button variant="destructive" size="sm" onClick={() => handleDelete(group.id)}>
                          {t('Delete')}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <GroupFormModal
          open={modalOpen}
          onOpenChange={setModalOpen}
          initialData={editingGroup}
          groups={groups}
          onSubmit={handleSubmit}
        />
      </div>
    </SecurityPageLayout>
  )
}
