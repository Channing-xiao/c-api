import { z } from 'zod'
import { useUrlFilters } from './use-url-filters'

const ruleFilterSchema = z.object({
  group_id: z.coerce.number().optional().default(0),
  type: z.coerce.number().optional().default(0),
  status: z.coerce.number().optional().default(0),
})

export type RuleFilters = z.infer<typeof ruleFilterSchema>

export function useRuleFilters() {
  return useUrlFilters(ruleFilterSchema)
}
