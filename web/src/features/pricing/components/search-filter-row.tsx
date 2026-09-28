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
import { Search } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'

import {
  getQuotaTypeLabels,
  QUOTA_TYPES,
  SORT_OPTIONS,
  type QuotaTypeOption,
  type SortOption,
} from '../constants'

interface SearchFilterRowProps {
  search: string
  onSearchChange: (value: string) => void
  sortBy: string
  onSortChange: (value: string) => void
  quotaType: string
  onQuotaTypeChange: (value: string) => void
  className?: string
}

/**
 * Reference-style toolbar: centered search plus two compact dropdowns
 * (price ordering and billing scale) shown to the right of it.
 */
export function SearchFilterRow(props: SearchFilterRowProps) {
  const { t } = useTranslation()
  const quotaLabels = getQuotaTypeLabels(t)
  const sortOptions: { value: SortOption; label: string }[] = [
    { value: SORT_OPTIONS.NAME, label: t('Default order') },
    { value: SORT_OPTIONS.PRICE_LOW, label: t('Price: Low to High') },
    { value: SORT_OPTIONS.PRICE_HIGH, label: t('Price: High to Low') },
  ]

  return (
    <div
      className={cn(
        'flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-center',
        props.className
      )}
    >
      <div className='relative w-full lg:max-w-md'>
        <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2' />
        <input
          type='text'
          value={props.search}
          onChange={(event) => props.onSearchChange(event.target.value)}
          placeholder={t('Enter keywords to search...')}
          aria-label={t('Search')}
          className='border-border/60 bg-background focus:border-blue-500/60 focus:ring-blue-500/15 h-10 w-full rounded-lg border pr-3 pl-10 text-sm shadow-xs outline-none transition-colors placeholder:text-muted-foreground/60 focus:ring-4'
        />
      </div>

      <div className='flex items-center gap-2.5'>
        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground hidden text-sm whitespace-nowrap sm:inline'>
            {t('Price range')}
          </span>
          <Select
            value={props.sortBy}
            onValueChange={(value) => props.onSortChange(value ?? props.sortBy)}
          >
            <SelectTrigger size='sm' className='min-w-28' aria-label={t('Price range')}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {sortOptions.map((option) => (
                <SelectItem key={option.value} value={option.value}>
                  {option.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className='flex items-center gap-1.5'>
          <span className='text-muted-foreground hidden text-sm whitespace-nowrap sm:inline'>
            {t('Usage scale')}
          </span>
          <Select
            value={props.quotaType}
            onValueChange={(value) =>
              props.onQuotaTypeChange(value ?? props.quotaType)
            }
          >
            <SelectTrigger size='sm' className='min-w-28' aria-label={t('Usage scale')}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {(
                [QUOTA_TYPES.ALL, QUOTA_TYPES.TOKEN, QUOTA_TYPES.REQUEST] as string[]
              ).map((option) => (
                <SelectItem key={option} value={option}>
                  {quotaLabels[option as QuotaTypeOption]}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
    </div>
  )
}
