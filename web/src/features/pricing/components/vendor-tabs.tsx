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
import { useTranslation } from 'react-i18next'

import { cn } from '@/lib/utils'

import { FILTER_ALL } from '../constants'

interface VendorTabsProps {
  vendors: string[]
  value: string
  onChange: (vendor: string) => void
  className?: string
}

/** Horizontal vendor selector with underline-style active state. */
export function VendorTabs(props: VendorTabsProps) {
  const { t } = useTranslation()
  const tabs = [
    { value: FILTER_ALL, label: t('All') },
    ...props.vendors.map((vendor) => ({ value: vendor, label: vendor })),
  ]

  return (
    <div
      role='tablist'
      aria-label={t('Vendors')}
      className={cn(
        'hover-scrollbar overflow-x-auto',
        '[scrollbar-width:none] [&::-webkit-scrollbar]:hidden',
        props.className
      )}
    >
      <div className='border-border/60 flex min-w-max items-center justify-start gap-1 border-b sm:justify-center'>
        {tabs.map((tab) => {
          const isActive = tab.value === props.value
          return (
            <button
              key={tab.value}
              type='button'
              role='tab'
              aria-selected={isActive}
              onClick={() => props.onChange(tab.value)}
              className={cn(
                'relative shrink-0 px-3.5 py-2.5 text-sm transition-colors',
                isActive
                  ? 'text-blue-600 font-medium dark:text-blue-400'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {tab.label}
              <span
                aria-hidden
                className={cn(
                  'absolute inset-x-2 -bottom-px h-0.5 rounded-full transition-all duration-200',
                  isActive
                    ? 'bg-blue-600 dark:bg-blue-400'
                    : 'bg-transparent'
                )}
              />
            </button>
          )
        })}
      </div>
    </div>
  )
}
