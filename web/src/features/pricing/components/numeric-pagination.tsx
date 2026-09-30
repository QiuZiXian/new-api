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
import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { ReactNode } from 'react'

import { cn } from '@/lib/utils'

interface NumericPaginationProps {
  page: number
  totalPages: number
  onPageChange: (page: number) => void
  className?: string
}

/** Build a compact page list: 1 … (c-1) c (c+1) … total. */
function buildPageList(page: number, totalPages: number): (number | '…')[] {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1)
  }
  const pages = new Set<number>([1, totalPages, page - 1, page, page + 1])
  const sorted = [...pages].filter((p) => p >= 1 && p <= totalPages).sort((a, b) => a - b)
  const list: (number | '…')[] = []
  let previous = 0
  for (const current of sorted) {
    if (current - previous > 1) list.push('…')
    list.push(current)
    previous = current
  }
  return list
}

export function NumericPagination(props: NumericPaginationProps) {
  const { page, totalPages } = props
  if (totalPages <= 1) return null
  const pages = buildPageList(page, totalPages)

  const buttonBase =
    'inline-flex size-8 items-center justify-center rounded-md border text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-40'

  // Build items with data-dependent keys; ellipses are keyed by the page
  // number they follow so multiple gaps stay unique.
  const items: ReactNode[] = []
  let lastPage = 0
  for (const item of pages) {
    if (item === '…') {
      items.push(
        <span
          key={`ellipsis-after-${lastPage}`}
          className='text-muted-foreground inline-flex size-8 items-center justify-center text-sm'
        >
          …
        </span>
      )
    } else {
      items.push(
        <button
          key={item}
          type='button'
          onClick={() => props.onPageChange(item)}
          aria-current={item === page ? 'page' : undefined}
          className={cn(
            buttonBase,
            item === page
              ? 'border-blue-600 bg-blue-600 font-medium text-white dark:border-blue-500 dark:bg-blue-500'
              : 'border-border/60 text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          {item}
        </button>
      )
      lastPage = item
    }
  }

  return (
    <nav
      aria-label='Pagination'
      className={cn('flex items-center justify-center gap-1.5', props.className)}
    >
      <button
        type='button'
        className={cn(buttonBase, 'border-border/60 text-muted-foreground hover:bg-muted')}
        disabled={page <= 1}
        onClick={() => props.onPageChange(page - 1)}
        aria-label='Previous page'
      >
        <ChevronLeft className='size-4' />
      </button>
      {items}
      <button
        type='button'
        className={cn(buttonBase, 'border-border/60 text-muted-foreground hover:bg-muted')}
        disabled={page >= totalPages}
        onClick={() => props.onPageChange(page + 1)}
        aria-label='Next page'
      >
        <ChevronRight className='size-4' />
      </button>
    </nav>
  )
}
