import { useTranslation } from 'react-i18next'
import { FolderOpen, Info, Plus, Copy } from 'lucide-react'
import { toast } from 'sonner'
import { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { EmptyState } from '@/components/empty-state'
import { aiSecurityApi, type AISecurityGroup } from '../api/ai-security'
import { AISecurityLayout } from '../components/ai-security-layout'
import { GroupFormModal } from '../components/group-form-modal'

export function AISecurityGroupPage() {
  const { t } = useTranslation()
  const [groups, setGroups] = useState<AISecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingGroup, setEditingGroup] = useState<AISecurityGroup | null>(null)

  useEffect(() => {
    loadGroups()
  }, [])

  const loadGroups = () => {
    setLoading(true)
    aiSecurityApi
      .getGroups({ page: 1, page_size: 200 })
      .then((res: any) => {
        if (res.success) setGroups(res.data.items)
      })
      .finally(() => setLoading(false))
  }

  const tree = useMemo(() => {
    const map = new Map<number, AISecurityGroup[]>()
    groups.forEach((g) => {
      const list = map.get(g.parent_id) ?? []
      list.push(g)
      map.set(g.parent_id, list)
    })
    const build = (parentId: number, depth: number): AISecurityGroup[] => {
      const list = (map.get(parentId) ?? []).sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
      return list.flatMap((g) => [g, ...build(g.id, depth + 1)])
    }
    return build(0, 0)
  }, [groups])

  const parentName = (parentId: number) =>
    parentId ? groups.find((g) => g.id === parentId)?.name ?? `#${parentId}` : t('aiSecurity.none')

  const handleDelete = async (id: number) => {
    if (!confirm(t('aiSecurity.confirmDeleteGroup'))) return
    try {
      await aiSecurityApi.deleteGroup(id)
      toast.success(t('aiSecurity.groupDeleted'))
      loadGroups()
    } catch {
      toast.error(t('aiSecurity.groupDeleteFailed'))
    }
  }

  const handleCopy = async (id: number) => {
    try {
      await aiSecurityApi.copyGroup(id)
      toast.success(t('aiSecurity.groupCopied'))
      loadGroups()
    } catch {
      toast.error(t('aiSecurity.groupCopyFailed'))
    }
  }

  const handleCreate = () => {
    setEditingGroup(null)
    setModalOpen(true)
  }

  const handleEdit = (group: AISecurityGroup) => {
    setEditingGroup(group)
    setModalOpen(true)
  }

  const handleSubmit = async (data: Partial<AISecurityGroup>) => {
    try {
      if (editingGroup) {
        await aiSecurityApi.updateGroup(editingGroup.id, data)
        toast.success(t('aiSecurity.groupUpdated'))
      } else {
        await aiSecurityApi.createGroup(data)
        toast.success(t('aiSecurity.groupCreated'))
      }
      loadGroups()
    } catch {
      toast.error(t('aiSecurity.groupSaveFailed'))
    }
  }

  const handleToggleStatus = async (group: AISecurityGroup) => {
    const newStatus = group.status === 1 ? 0 : 1
    try {
      await aiSecurityApi.updateGroupStatus(group.id, newStatus)
      toast.success(t(newStatus === 1 ? 'aiSecurity.groupEnabled' : 'aiSecurity.groupDisabled'))
      loadGroups()
    } catch {
      toast.error(t('aiSecurity.groupStatusUpdateFailed'))
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
          {t('aiSecurity.createGroup')}
        </Button>
      }
    >
      <div className='space-y-4'>
        <div className='flex items-center justify-between gap-4 rounded-xl border bg-muted/30 p-4'>
          <div className='flex items-start gap-3'>
            <div className='mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Info className='size-4' />
            </div>
            <div className='text-sm'>
              <span className='font-medium'>{t('aiSecurity.groupTipTitle')}</span>
              <p className='text-muted-foreground'>{t('aiSecurity.groupTipDesc')}</p>
            </div>
          </div>
        </div>

        <Card>
          <CardContent className='p-0'>
            {groups.length === 0 ? (
              <EmptyState
                icon={FolderOpen}
                title={t('aiSecurity.noGroups')}
                description={t('aiSecurity.noGroupsDesc')}
                action={
                  <Button onClick={handleCreate}>
                    <Plus className='mr-1.5 size-4' />
                    {t('aiSecurity.createGroup')}
                  </Button>
                }
                className='min-h-[260px]'
                bordered={false}
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t('aiSecurity.name')}</TableHead>
                    <TableHead>{t('aiSecurity.parentGroup')}</TableHead>
                    <TableHead>{t('aiSecurity.description')}</TableHead>
                    <TableHead>{t('aiSecurity.status')}</TableHead>
                    <TableHead className='text-right'>{t('aiSecurity.actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {tree.map((group) => {
                    const depth = group.path.split('/').filter(Boolean).length - 1
                    return (
                      <TableRow key={group.id}>
                        <TableCell className='font-medium'>
                          <span style={{ paddingLeft: `${depth * 24}px` }}></span>
                          {group.name}
                        </TableCell>
                        <TableCell className='text-muted-foreground'>{parentName(group.parent_id)}</TableCell>
                        <TableCell className='text-muted-foreground'>{group.description}</TableCell>
                        <TableCell>
                          <div className='flex items-center gap-2'>
                            <Switch
                              checked={group.status === 1}
                              onCheckedChange={() => handleToggleStatus(group)}
                              size='sm'
                            />
                            <span className='text-sm'>
                              {group.status === 1 ? t('aiSecurity.enabled') : t('aiSecurity.disabled')}
                            </span>
                          </div>
                        </TableCell>
                        <TableCell className='text-right space-x-2'>
                          <Button variant='outline' size='sm' onClick={() => handleCopy(group.id)}>
                            <Copy className='mr-1.5 size-3.5' />
                            {t('aiSecurity.copy')}
                          </Button>
                          <Button variant='outline' size='sm' onClick={() => handleEdit(group)}>
                            {t('aiSecurity.edit')}
                          </Button>
                          <Button variant='destructive' size='sm' onClick={() => handleDelete(group.id)}>
                            {t('aiSecurity.delete')}
                          </Button>
                        </TableCell>
                      </TableRow>
                    )
                  })}
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
    </AISecurityLayout>
  )
}
