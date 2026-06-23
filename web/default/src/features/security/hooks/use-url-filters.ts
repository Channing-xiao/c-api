import { useMemo } from 'react'
import { useSearch } from '@tanstack/react-router'
import { z } from 'zod'

export function useUrlFilters<T extends z.ZodTypeAny>(schema: T) {
  type FilterType = z.infer<T>
  const search = useSearch({ from: '__root__' }) as Record<string, unknown>

  const filters = useMemo(() => {
    const parsed = schema.safeParse(search)
    if (parsed.success) {
      return parsed.data as FilterType
    }
    return {} as FilterType
  }, [search, schema])

  return filters
}
