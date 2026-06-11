import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { securityApi, type SecurityGroup, type SecurityRule } from '../api/security'
import { SecurityPageLayout } from '../components/security-page-layout'
import { RuleFormModal } from '../components/rule-form-modal'
import { RuleTester } from '../components/rule-tester'
import { ConfirmDialog } from '@/components/confirm-dialog'

const ruleTypeMap: Record<number, string> = {
  1: 'Keyword',
  2: 'Regex',
  3: 'NER',
  4: 'AI',
}

const actionMap: Record<number, string> = {
  1: 'Pass',
  2: 'Alert',
  3: 'Mask',
  4: 'Block',
  5: 'Review',
}

export function SecurityRulePage() {
  const { t } = useTranslation()
  const [rules, setRules] = useState<SecurityRule[]>([])
  const [groups, setGroups] = useState<SecurityGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<SecurityRule | null>(null)
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set())
  const [testRuleId, setTestRuleId] = useState<number | null>(null)
  const [testRuleName, setTestRuleName] = useState('')
  const [testModalOpen, setTestModalOpen] = useState(false)
  const [batchConfirmOpen, setBatchConfirmOpen] = useState(false)
  const [batchAction, setBatchAction] = useState<'delete' | 'enable' | 'disable'>('delete')

  useEffect(() => {
    loadRules()
    securityApi.getGroups({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) setGroups(res.data.items)
    })
  }, [])

  const loadRules = () => {
    setLoading(true)
    securityApi.getRules({ page: 1, page_size: 100 }).then((res: any) => {
      if (res.success) {
        setRules(res.data.items)
      }
      setLoading(false)
    })
  }

  const handleDelete = async (id: number) => {
    if (!confirm(t('Are you sure?'))) return
    try {
      await securityApi.deleteRule(id)
      toast.success(t('Rule deleted'))
      loadRules()
    } catch {
      toast.error(t('Failed to delete rule'))
    }
  }

  const handleCreate = () => {
    setEditingRule(null)
    setModalOpen(true)
  }

  const handleEdit = (rule: SecurityRule) => {
    setEditingRule(rule)
    setModalOpen(true)
  }

  const handleSubmit = async (data: Partial<SecurityRule>) => {
    try {
      if (editingRule) {
        await securityApi.updateRule(editingRule.id, data)
        toast.success(t('Rule updated'))
      } else {
        await securityApi.createRule(data)
        toast.success(t('Rule created'))
      }
      loadRules()
    } catch {
      toast.error(t('Failed to save rule'))
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

  const openTest = (rule: SecurityRule) => {
    setTestRuleId(rule.id)
    setTestRuleName(rule.name)
    setTestModalOpen(true)
  }

  const handleBatchAction = async () => {
    const ids = Array.from(selectedIds)
    try {
      if (batchAction === 'delete') {
        await securityApi.batchDeleteRules(ids)
        toast.success(t('Rules deleted'))
      } else if (batchAction === 'enable') {
        await securityApi.batchUpdateRuleStatus(ids, 1)
        toast.success(t('Rules enabled'))
      } else if (batchAction === 'disable') {
        await securityApi.batchUpdateRuleStatus(ids, 0)
        toast.success(t('Rules disabled'))
      }
      setSelectedIds(new Set())
      setBatchConfirmOpen(false)
      loadRules()
    } catch {
      toast.error(t('Batch operation failed'))
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
      actions={<Button onClick={handleCreate}>{t('Create Rule')}</Button>}
    >
      <div className="space-y-4">
      {selectedIds.size > 0 && (
        <div className="flex items-center gap-2 rounded-lg border p-3">
          <span className="text-sm">{t('{{count}} selected', { count: selectedIds.size })}</span>
          <Button variant="outline" size="sm" onClick={() => { setBatchAction('enable'); setBatchConfirmOpen(true) }}>
            {t('Batch Enable')}
          </Button>
          <Button variant="outline" size="sm" onClick={() => { setBatchAction('disable'); setBatchConfirmOpen(true) }}>
            {t('Batch Disable')}
          </Button>
          <Button variant="destructive" size="sm" onClick={() => { setBatchAction('delete'); setBatchConfirmOpen(true) }}>
            {t('Batch Delete')}
          </Button>
        </div>
      )}
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10">
                  <Checkbox
                    checked={selectedIds.size > 0 && selectedIds.size === rules.length}
                    onCheckedChange={toggleSelectAll}
                  />
                </TableHead>
                <TableHead>{t('Name')}</TableHead>
                <TableHead>{t('Group')}</TableHead>
                <TableHead>{t('Type')}</TableHead>
                <TableHead>{t('Action')}</TableHead>
                <TableHead>{t('Risk Score')}</TableHead>
                <TableHead className="text-right">{t('Actions')}</TableHead>
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
                  <TableCell className="font-medium">{rule.name}</TableCell>
                  <TableCell>{rule.group_name}</TableCell>
                  <TableCell><Badge variant="outline">{ruleTypeMap[rule.type] ?? rule.type}</Badge></TableCell>
                  <TableCell><Badge>{actionMap[rule.action] ?? rule.action}</Badge></TableCell>
                  <TableCell>{rule.risk_score}</TableCell>
                  <TableCell className="text-right space-x-2">
                    <Button variant="outline" size="sm" onClick={() => openTest(rule)}>{t('Test')}</Button>
                    <Button variant="outline" size="sm" onClick={() => handleEdit(rule)}>{t('Edit')}</Button>
                    <Button variant="destructive" size="sm" onClick={() => handleDelete(rule.id)}>
                      {t('Delete')}
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <RuleFormModal
        open={modalOpen}
        onOpenChange={setModalOpen}
        initialData={editingRule}
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
            ? t('Confirm Batch Delete')
            : batchAction === 'enable'
            ? t('Confirm Batch Enable')
            : t('Confirm Batch Disable')
        }
        desc={t('This action will affect {{count}} rules.', { count: selectedIds.size })}
        handleConfirm={handleBatchAction}
        destructive={batchAction === 'delete'}
      />
      </div>
    </SecurityPageLayout>
  )
}
