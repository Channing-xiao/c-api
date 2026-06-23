import { z } from 'zod'
import { useUrlFilters } from './use-url-filters'

const policyFilterSchema = z.object({
  user_id: z.coerce.number().optional().default(0),
  status: z.coerce.number().optional().default(0),
})

export type PolicyFilters = z.infer<typeof policyFilterSchema>

export function usePolicyFilters() {
  return useUrlFilters(policyFilterSchema)
}
