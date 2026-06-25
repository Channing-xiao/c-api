import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { AISecurityGroup, AISecurityRule } from '../api/ai-security'
import { ruleTypeOptions, actionOptions } from '../constants'

interface RuleFormModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData: AISecurityRule | null
  groups: AISecurityGroup[]
  onSubmit: (data: Partial<AISecurityRule>) => Promise<void>
}

export function RuleFormModal(props: RuleFormModalProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState<Partial<AISecurityRule>>({
    group_id: 0,
    name: '',
    type: 1,
    content: '',
    extra_config: '',
    action: 1,
    priority: 0,
    risk_score: 50,
  })

  useEffect(() => {
    if (props.open) {
      setForm(
        props.initialData
          ? {
              group_id: props.initialData.group_id,
              name: props.initialData.name,
              type: props.initialData.type,
              content: props.initialData.content,
              extra_config: props.initialData.extra_config,
              action: props.initialData.action,
              priority: props.initialData.priority,
              risk_score: props.initialData.risk_score,
            }
          : {
              group_id: props.groups[0]?.id ?? 0,
              name: '',
              type: 1,
              content: '',
              extra_config: '',
              action: 1,
              priority: 0,
              risk_score: 50,
            }
      )
    }
  }, [props.open, props.initialData, props.groups])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name?.trim() || !form.content?.trim() || !form.group_id) return
    setLoading(true)
    try {
      await props.onSubmit(form)
      props.onOpenChange(false)
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className='sm:max-w-lg'>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>
              {props.initialData ? t('aiSecurity.editRule') : t('aiSecurity.createRule')}
            </DialogTitle>
          </DialogHeader>

          <div className='grid grid-cols-1 gap-4 py-4 sm:grid-cols-2'>
            <div className='space-y-2'>
              <Label htmlFor='rule-group'>{t('aiSecurity.group')}</Label>
              <Select
                value={String(form.group_id ?? 0)}
                onValueChange={(v) => setForm({ ...form, group_id: Number(v) })}
              >
                <SelectTrigger id='rule-group' className='w-full'>
                  <SelectValue placeholder={t('aiSecurity.selectGroup')} />
                </SelectTrigger>
                <SelectContent>
                  {props.groups.map((g) => (
                    <SelectItem key={g.id} value={String(g.id)}>
                      {g.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2'>
              <Label htmlFor='rule-name'>{t('aiSecurity.name')}</Label>
              <Input
                id='rule-name'
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder={t('aiSecurity.ruleNamePlaceholder')}
                required
              />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='rule-type'>{t('aiSecurity.type')}</Label>
              <Select
                value={String(form.type ?? 1)}
                onValueChange={(v) => setForm({ ...form, type: Number(v) })}
              >
                <SelectTrigger id='rule-type' className='w-full'>
                  <SelectValue placeholder={t('aiSecurity.selectType')} />
                </SelectTrigger>
                <SelectContent>
                  {ruleTypeOptions.map((opt) => (
                    <SelectItem key={opt.value} value={String(opt.value)}>
                      {t(opt.labelKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2'>
              <Label htmlFor='rule-action'>{t('aiSecurity.action')}</Label>
              <Select
                value={String(form.action ?? 1)}
                onValueChange={(v) => setForm({ ...form, action: Number(v) })}
              >
                <SelectTrigger id='rule-action' className='w-full'>
                  <SelectValue placeholder={t('aiSecurity.selectAction')} />
                </SelectTrigger>
                <SelectContent>
                  {actionOptions.map((opt) => (
                    <SelectItem key={opt.value} value={String(opt.value)}>
                      {t(opt.labelKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2 sm:col-span-2'>
              <Label htmlFor='rule-content'>{t('aiSecurity.content')}</Label>
              <textarea
                id='rule-content'
                value={form.content}
                onChange={(e) => setForm({ ...form, content: e.target.value })}
                placeholder={t('aiSecurity.ruleContentPlaceholder')}
                required
                rows={4}
                className='border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:bg-input/30 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 w-full rounded-lg border bg-transparent px-2.5 py-1 text-base transition-colors outline-none focus-visible:ring-3 md:text-sm'
              />
            </div>

            <div className='space-y-2 sm:col-span-2'>
              <Label htmlFor='rule-extra'>{t('aiSecurity.extraConfig')}</Label>
              <textarea
                id='rule-extra'
                value={form.extra_config}
                onChange={(e) => setForm({ ...form, extra_config: e.target.value })}
                placeholder={t('aiSecurity.ruleExtraPlaceholder')}
                rows={2}
                className='border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:bg-input/30 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 w-full rounded-lg border bg-transparent px-2.5 py-1 text-base transition-colors outline-none focus-visible:ring-3 md:text-sm'
              />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='rule-priority'>{t('aiSecurity.priority')}</Label>
              <Input
                id='rule-priority'
                type='number'
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
              />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='rule-risk'>{t('aiSecurity.riskScore')}</Label>
              <Input
                id='rule-risk'
                type='number'
                min={0}
                max={100}
                value={form.risk_score}
                onChange={(e) => setForm({ ...form, risk_score: Number(e.target.value) })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => props.onOpenChange(false)}>
              {t('aiSecurity.cancel')}
            </Button>
            <Button
              type='submit'
              disabled={loading || !form.name?.trim() || !form.content?.trim() || !form.group_id}
            >
              {loading ? t('aiSecurity.saving') : t('aiSecurity.save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
