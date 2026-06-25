import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { FileText, Plus, Trash2, Copy } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { EmptyState } from '@/components/empty-state'
import { aiSecurityApi, type AISecurityGroup, type AISecurityRule } from '../api/ai-security'
import { AISecurityLayout } from '../components/ai-security-layout'
import { RuleFormModal } from '../components/rule-form-modal'
import { RuleTester } from '../components/rule-tester'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { ruleTypeOptions, actionOptions } from '../constants'

export function AISecurityRulePage() {
  const { t } = useTranslation()
  const [rules, setRules] = useState<AISecurityRule[]>([])
  const [groups, setGroups] = useState<AISecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AISecurityRule | null>(null)
  const [formInitial, setFormInitial] = useState<AISecurityRule | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [testRuleId, setTestRuleId] = useState<number | null>(null)
  const [testRuleName, setTestRuleName] = useState('')
  const [testModalOpen, setTestModalOpen] = useState(false)
  const [batchConfirmOpen, setBatchConfirmOpen] = useState(false)
  const [batchAction, setBatchAction] = useState<'delete' | 'enable' | 'disable'>('delete')

  useEffect(() => {
    loadRules()
    aiSecurityApi.getGroups({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) setGroups(res.data.items)
    })
  }, [])

  const loadRules = () => {
    setLoading(true)
    aiSecurityApi
      .getRules({ page: 1, page_size: 100 })
      .then((res: any) => {
        if (res.success) setRules(res.data.items)
      })
      .finally(() => setLoading(false))
  }

  const groupName = (id: number) => groups.find((g) => g.id === id)?.name ?? '-'

  const getLabel = (options: { value: number; labelKey: string }[], value: number) =>
    options.find((o) => o.value === value)?.labelKey ?? String(value)

  const handleDelete = async (id: number) => {
    if (!confirm(t('aiSecurity.confirmDeleteRule'))) return
    try {
      await aiSecurityApi.deleteRule(id)
      toast.success(t('aiSecurity.ruleDeleted'))
      loadRules()
    } catch {
      toast.error(t('aiSecurity.ruleDeleteFailed'))
    }
  }

  const handleCreate = () => {
    setEditingRule(null)
    setFormInitial(null)
    setModalOpen(true)
  }

  const handleEdit = (rule: AISecurityRule) => {
    setEditingRule(rule)
    setFormInitial(rule)
    setModalOpen(true)
  }

  // 复制：以现有规则预填表单，但保持"新建"语义（提交走 createRule）
  const handleCopy = (rule: AISecurityRule) => {
    setEditingRule(null)
    setFormInitial({ ...rule, name: rule.name + t('aiSecurity.copySuffix') })
    setModalOpen(true)
  }

  const handleSubmit = async (data: Partial<AISecurityRule>) => {
    try {
      if (editingRule) {
        await aiSecurityApi.updateRule(editingRule.id, data)
        toast.success(t('aiSecurity.ruleUpdated'))
      } else {
        await aiSecurityApi.createRule(data)
        toast.success(t('aiSecurity.ruleCreated'))
      }
      loadRules()
    } catch {
      toast.error(t('aiSecurity.ruleSaveFailed'))
    }
  }

  const toggleSelect = (id: number) => {
    const next = new Set(selectedIds)
    if (next.has(id)) {
      next.delete(id)
    } else {
      next.add(id)
    }
    setSelectedIds(next)
  }

  const toggleSelectAll = () => {
    if (selectedIds.size === rules.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(rules.map((r) => r.id)))
    }
  }

  const openTest = (rule: AISecurityRule) => {
    setTestRuleId(rule.id)
    setTestRuleName(rule.name)
    setTestModalOpen(true)
  }

  const handleBatchAction = async () => {
    const ids = Array.from(selectedIds)
    try {
      if (batchAction === 'delete') {
        await aiSecurityApi.batchDeleteRules(ids)
        toast.success(t('aiSecurity.rulesDeleted'))
      } else if (batchAction === 'enable') {
        await aiSecurityApi.batchUpdateRuleStatus(ids, 1)
        toast.success(t('aiSecurity.rulesEnabled'))
      } else if (batchAction === 'disable') {
        await aiSecurityApi.batchUpdateRuleStatus(ids, 0)
        toast.success(t('aiSecurity.rulesDisabled'))
      }
      setSelectedIds(new Set())
      setBatchConfirmOpen(false)
      loadRules()
    } catch {
      toast.error(t('aiSecurity.batchOperationFailed'))
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
          {t('aiSecurity.createRule')}
        </Button>
      }
    >
      <div className='space-y-4'>
        {selectedIds.size > 0 && (
          <div className='flex flex-wrap items-center gap-2 rounded-xl border bg-muted/30 p-3'>
            <span className='text-sm font-medium'>
              {t('aiSecurity.selectedCount', { count: selectedIds.size })}
            </span>
            <div className='ml-auto flex items-center gap-2'>
              <Button
                variant='outline'
                size='sm'
                onClick={() => {
                  setBatchAction('enable')
                  setBatchConfirmOpen(true)
                }}
              >
                {t('aiSecurity.batchEnable')}
              </Button>
              <Button
                variant='outline'
                size='sm'
                onClick={() => {
                  setBatchAction('disable')
                  setBatchConfirmOpen(true)
                }}
              >
                {t('aiSecurity.batchDisable')}
              </Button>
              <Button
                variant='destructive'
                size='sm'
                onClick={() => {
                  setBatchAction('delete')
                  setBatchConfirmOpen(true)
                }}
              >
                <Trash2 className='mr-1.5 size-3.5' />
                {t('aiSecurity.batchDelete')}
              </Button>
            </div>
          </div>
        )}

        <Card>
          <CardContent className='p-0'>
            {rules.length === 0 ? (
              <EmptyState
                icon={FileText}
                title={t('aiSecurity.noRules')}
                description={t('aiSecurity.noRulesDesc')}
                action={
                  <Button onClick={handleCreate}>
                    <Plus className='mr-1.5 size-4' />
                    {t('aiSecurity.createRule')}
                  </Button>
                }
                className='min-h-[260px]'
                bordered={false}
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className='w-10'>
                      <Checkbox
                        checked={selectedIds.size > 0 && selectedIds.size === rules.length}
                        onCheckedChange={toggleSelectAll}
                      />
                    </TableHead>
                    <TableHead>{t('aiSecurity.name')}</TableHead>
                    <TableHead>{t('aiSecurity.group')}</TableHead>
                    <TableHead>{t('aiSecurity.type')}</TableHead>
                    <TableHead>{t('aiSecurity.action')}</TableHead>
                    <TableHead>{t('aiSecurity.riskScore')}</TableHead>
                    <TableHead className='text-right'>{t('aiSecurity.actions')}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((rule) => (
                    <TableRow key={rule.id}>
                      <TableCell>
                        <Checkbox
                          checked={selectedIds.has(rule.id)}
                          onCheckedChange={() => toggleSelect(rule.id)}
                        />
                      </TableCell>
                      <TableCell className='font-medium'>{rule.name}</TableCell>
                      <TableCell className='text-muted-foreground'>{groupName(rule.group_id)}</TableCell>
                      <TableCell>
                        <Badge variant='outline'>{t(getLabel(ruleTypeOptions, rule.type))}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge>{t(getLabel(actionOptions, rule.action))}</Badge>
                      </TableCell>
                      <TableCell>{rule.risk_score}</TableCell>
                      <TableCell className='text-right space-x-2'>
                        <Button variant='outline' size='sm' onClick={() => openTest(rule)}>
                          {t('aiSecurity.test')}
                        </Button>
                        <Button variant='outline' size='sm' onClick={() => handleEdit(rule)}>
                          {t('aiSecurity.edit')}
                        </Button>
                        <Button variant='outline' size='sm' onClick={() => handleCopy(rule)}>
                          <Copy className='mr-1.5 size-3.5' />
                          {t('aiSecurity.copy')}
                        </Button>
                        <Button variant='destructive' size='sm' onClick={() => handleDelete(rule.id)}>
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

        <RuleFormModal
          open={modalOpen}
          onOpenChange={setModalOpen}
          initialData={formInitial}
          groups={groups}
          onSubmit={handleSubmit}
        />

        {testRuleId != null && (
          <RuleTester
            ruleId={testRuleId}
            ruleName={testRuleName}
            open={testModalOpen}
            onOpenChange={setTestModalOpen}
          />
        )}

        <ConfirmDialog
          open={batchConfirmOpen}
          onOpenChange={setBatchConfirmOpen}
          title={
            batchAction === 'delete'
              ? t('aiSecurity.confirmBatchDelete')
              : batchAction === 'enable'
              ? t('aiSecurity.confirmBatchEnable')
              : t('aiSecurity.confirmBatchDisable')
          }
          desc={t('aiSecurity.batchAffectCount', { count: selectedIds.size })}
          handleConfirm={handleBatchAction}
          destructive={batchAction === 'delete'}
        />
      </div>
    </AISecurityLayout>
  )
}
