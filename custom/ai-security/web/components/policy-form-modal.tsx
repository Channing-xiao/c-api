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
import type { AISecurityGroup, AISecurityPolicy } from '../api/ai-security'
import { actionOptions, scopeOptions, statusOptions } from '../constants'

interface PolicyFormModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  initialData: AISecurityPolicy | null
  groups: AISecurityGroup[]
  onSubmit: (data: Partial<AISecurityPolicy>) => Promise<void>
}

export function PolicyFormModal(props: PolicyFormModalProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [form, setForm] = useState<Partial<AISecurityPolicy>>({
    user_id: 0,
    group_id: 0,
    scope: 3,
    default_action: 4,
    custom_response: '',
    whitelist_ips: '',
    priority: 0,
    status: 1,
  })

  useEffect(() => {
    if (props.open) {
      setForm(
        props.initialData
          ? {
              user_id: props.initialData.user_id,
              group_id: props.initialData.group_id,
              scope: props.initialData.scope,
              default_action: props.initialData.default_action,
              custom_response: props.initialData.custom_response,
              whitelist_ips: props.initialData.whitelist_ips,
              priority: props.initialData.priority,
              status: props.initialData.status,
            }
          : {
              user_id: 0,
              group_id: props.groups[0]?.id ?? 0,
              scope: 3,
              default_action: 4,
              custom_response: '',
              whitelist_ips: '',
              priority: 0,
              status: 1,
            }
      )
    }
  }, [props.open, props.initialData, props.groups])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.user_id || !form.group_id) return
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
              {props.initialData ? t('aiSecurity.editPolicy') : t('aiSecurity.createPolicy')}
            </DialogTitle>
          </DialogHeader>

          <div className='grid grid-cols-1 gap-4 py-4 sm:grid-cols-2'>
            <div className='space-y-2'>
              <Label htmlFor='policy-user'>{t('aiSecurity.userId')}</Label>
              <Input
                id='policy-user'
                type='number'
                value={form.user_id}
                onChange={(e) => setForm({ ...form, user_id: Number(e.target.value) })}
                required
              />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='policy-group'>{t('aiSecurity.group')}</Label>
              <Select
                value={String(form.group_id ?? 0)}
                onValueChange={(v) => setForm({ ...form, group_id: Number(v) })}
              >
                <SelectTrigger id='policy-group' className='w-full'>
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
              <Label htmlFor='policy-scope'>{t('aiSecurity.scope')}</Label>
              <Select
                value={String(form.scope ?? 3)}
                onValueChange={(v) => setForm({ ...form, scope: Number(v) })}
              >
                <SelectTrigger id='policy-scope' className='w-full'>
                  <SelectValue placeholder={t('aiSecurity.selectScope')} />
                </SelectTrigger>
                <SelectContent>
                  {scopeOptions.map((opt) => (
                    <SelectItem key={opt.value} value={String(opt.value)}>
                      {t(opt.labelKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2'>
              <Label htmlFor='policy-action'>{t('aiSecurity.defaultAction')}</Label>
              <Select
                value={String(form.default_action ?? 4)}
                onValueChange={(v) => setForm({ ...form, default_action: Number(v) })}
              >
                <SelectTrigger id='policy-action' className='w-full'>
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

            <div className='space-y-2'>
              <Label htmlFor='policy-priority'>{t('aiSecurity.priority')}</Label>
              <Input
                id='policy-priority'
                type='number'
                value={form.priority}
                onChange={(e) => setForm({ ...form, priority: Number(e.target.value) })}
              />
            </div>

            <div className='space-y-2'>
              <Label htmlFor='policy-status'>{t('aiSecurity.status')}</Label>
              <Select
                value={String(form.status ?? 1)}
                onValueChange={(v) => setForm({ ...form, status: Number(v) })}
              >
                <SelectTrigger id='policy-status' className='w-full'>
                  <SelectValue placeholder={t('aiSecurity.selectStatus')} />
                </SelectTrigger>
                <SelectContent>
                  {statusOptions.map((opt) => (
                    <SelectItem key={opt.value} value={String(opt.value)}>
                      {t(opt.labelKey)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-2 sm:col-span-2'>
              <Label htmlFor='policy-whitelist'>{t('aiSecurity.whitelistIps')}</Label>
              <Input
                id='policy-whitelist'
                value={form.whitelist_ips}
                onChange={(e) => setForm({ ...form, whitelist_ips: e.target.value })}
                placeholder={t('aiSecurity.whitelistIpsPlaceholder')}
              />
            </div>

            <div className='space-y-2 sm:col-span-2'>
              <Label htmlFor='policy-custom-response'>{t('aiSecurity.customResponse')}</Label>
              <textarea
                id='policy-custom-response'
                value={form.custom_response}
                onChange={(e) => setForm({ ...form, custom_response: e.target.value })}
                placeholder={t('aiSecurity.customResponsePlaceholder')}
                rows={3}
                className='border-input focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 dark:bg-input/30 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40 w-full rounded-lg border bg-transparent px-2.5 py-1 text-base transition-colors outline-none focus-visible:ring-3 md:text-sm'
              />
            </div>
          </div>

          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => props.onOpenChange(false)}>
              {t('aiSecurity.cancel')}
            </Button>
            <Button type='submit' disabled={loading || !form.user_id || !form.group_id}>
              {loading ? t('aiSecurity.saving') : t('aiSecurity.save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
