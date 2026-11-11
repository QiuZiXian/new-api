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
import { Filter, RotateCcw } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import { FILTER_ALL } from '../constants'

interface PricingTagBarProps {
  tags: string[]
  value: string
  onChange: (value: string) => void
  onOpenFilters: () => void
  onClearFilters: () => void
  hasActiveFilters: boolean
  activeFilterCount: number
  className?: string
}

export function PricingTagBar(props: PricingTagBarProps) {
  const { t } = useTranslation()

  return (
    <div
      className={cn(
        'border-border/60 bg-card/60 flex flex-wrap items-center gap-2 rounded-xl border px-3 py-2.5',
        props.className
      )}
    >
      <div className='hover-scrollbar flex min-w-0 flex-1 items-center gap-1.5 overflow-x-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden'>
        <button
          type='button'
          onClick={() => props.onChange(FILTER_ALL)}
          className={cn(
            'shrink-0 rounded-full px-3 py-1.5 text-xs font-medium transition-colors',
            props.value === FILTER_ALL
              ? 'bg-blue-600 text-white shadow-sm'
              : 'bg-muted text-muted-foreground hover:text-foreground'
          )}
        >
          {t('All Models')}
        </button>
        {props.tags.map((tag) => {
          const active = props.value.toLowerCase() === tag.toLowerCase()
          return (
            <button
              key={tag}
              type='button'
              onClick={() => props.onChange(active ? FILTER_ALL : tag)}
              className={cn(
                'shrink-0 rounded-full border px-3 py-1.5 text-xs transition-colors',
                active
                  ? 'border-blue-500 bg-blue-50 font-medium text-blue-700 dark:border-blue-400/60 dark:bg-blue-500/15 dark:text-blue-300'
                  : 'border-border/70 text-muted-foreground hover:border-blue-400/60 hover:text-foreground'
              )}
            >
              {tag}
            </button>
          )
        })}
      </div>

      <div className='ml-auto flex shrink-0 items-center gap-1.5 border-l pl-2.5'>
        <button
          type='button'
          onClick={props.onOpenFilters}
          className='border-border/70 text-muted-foreground hover:text-foreground inline-flex items-center gap-1.5 rounded-md border px-2 py-1.5 text-xs transition-colors'
        >
          <Filter className='size-3.5' />
          {t('Filter')}
          {props.activeFilterCount > 0 && (
            <span className='rounded bg-blue-600 px-1 text-[10px] leading-4 text-white tabular-nums'>
              {props.activeFilterCount}
            </span>
          )}
        </button>
        <button
          type='button'
          onClick={props.onClearFilters}
          disabled={!props.hasActiveFilters}
          className='text-muted-foreground hover:text-foreground inline-flex items-center gap-1 px-1.5 py-1.5 text-xs transition-colors disabled:cursor-not-allowed disabled:opacity-40'
        >
          <RotateCcw className='size-3.5' />
          {t('Clear filters')}
        </button>
      </div>
    </div>
  )
}
