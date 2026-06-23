import { z } from 'zod'
import { useUrlFilters } from './use-url-filters'

const logFilterSchema = z.object({
  start_time: z.coerce.number().optional().default(0),
  end_time: z.coerce.number().optional().default(0),
  model_name: z.string().optional().default(''),
  user_id: z.coerce.number().optional().default(0),
  action: z.coerce.number().optional().default(0),
  risk_level: z.coerce.number().optional().default(0),
  content_type: z.coerce.number().optional().default(0),
})

export type LogFilters = z.infer<typeof logFilterSchema>

export function useLogFilters() {
  return useUrlFilters(logFilterSchema)
}
