/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'

import { getPricing } from '@/features/pricing/api'

/**
 * Model options for the per-model quota breakdown, sourced from the model
 * square (`/api/pricing`). Shares the `['pricing']` query cache with the
 * model square page so opening the drawer does not refetch.
 */
export function useModelSquareOptions(enabled = true) {
  const { data, isLoading } = useQuery({
    queryKey: ['pricing'],
    queryFn: getPricing,
    staleTime: 5 * 60 * 1000,
    enabled,
  })

  const options = useMemo(() => {
    const names = new Set<string>()
    for (const model of data?.data ?? []) {
      const name = model.model_name?.trim()
      if (name) names.add(name)
    }
    return [...names]
      .sort((a, b) => a.localeCompare(b))
      .map((name) => ({ value: name, label: name }))
  }, [data])

  return { options, isLoading }
}
