import type { LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Skeleton } from '@/components/ui/skeleton'

type StatTone = 'default' | 'rose' | 'amber' | 'teal'

interface SecurityStatCardProps {
  title: string
  value: string | number
  description?: string
  icon: LucideIcon
  tone?: StatTone
  loading?: boolean
}

const TONE_CLASSES: Record<StatTone, string> = {
  default: 'bg-primary/10 text-primary',
  rose: 'bg-rose-500/10 text-rose-600 dark:text-rose-400',
  amber: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
  teal: 'bg-teal-500/10 text-teal-600 dark:text-teal-400',
}

export function SecurityStatCard({
  title,
  value,
  description,
  icon: Icon,
  tone = 'default',
  loading,
}: SecurityStatCardProps) {
  return (
    <div className="group flex flex-col justify-between gap-3 rounded-xl border bg-card p-4 transition-all hover:shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <div className={cn('flex size-8 items-center justify-center rounded-lg', TONE_CLASSES[tone])}>
            <Icon className="size-4" />
          </div>
          <span className="text-xs font-medium text-muted-foreground">{title}</span>
        </div>
      </div>

      {loading ? (
        <div className="flex flex-col gap-1.5">
          <Skeleton className="h-7 w-20" />
          {description && <Skeleton className="h-3 w-28" />}
        </div>
      ) : (
        <div className="flex flex-col gap-0.5">
          <div className="text-2xl font-semibold tracking-tight tabular-nums">{value}</div>
          {description && (
            <p className="text-xs text-muted-foreground">{description}</p>
          )}
        </div>
      )}
    </div>
  )
}
