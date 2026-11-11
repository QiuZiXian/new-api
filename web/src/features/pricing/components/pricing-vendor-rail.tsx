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
import { Grid2X2 } from 'lucide-react'
import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'

import { getLobeIcon } from '@/lib/lobe-icon'
import { cn } from '@/lib/utils'

import { FILTER_ALL } from '../constants'
import type { PricingModel, PricingVendor } from '../types'
import { ProviderLogo } from './provider-logo'

interface PricingVendorRailProps {
  vendors: PricingVendor[]
  models: PricingModel[]
  value: string
  onChange: (value: string) => void
  className?: string
}

export function PricingVendorRail(props: PricingVendorRailProps) {
  const { t } = useTranslation()

  const vendors = useMemo(() => {
    const counts = new Map<string, number>()
    for (const model of props.models) {
      if (!model.vendor_name) continue
      counts.set(model.vendor_name, (counts.get(model.vendor_name) ?? 0) + 1)
    }
    return props.vendors
      .map((vendor) => ({ ...vendor, count: counts.get(vendor.name) ?? 0 }))
      .filter((vendor) => vendor.count > 0)
      .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
  }, [props.models, props.vendors])

  return (
    <aside
      className={cn(
        'border-border/60 bg-card/60 h-fit rounded-xl border p-2.5',
        props.className
      )}
    >
      <div className='flex items-center justify-between gap-2 px-1.5 pt-1 pb-2.5'>
        <h2 className='text-foreground text-sm font-semibold'>
          {t('Model Vendors')}
        </h2>
        <span className='bg-muted text-muted-foreground rounded-md px-1.5 py-0.5 text-[11px] tabular-nums'>
          {vendors.length}
        </span>
      </div>

      <div className='space-y-0.5'>
        <button
          type='button'
          onClick={() => props.onChange(FILTER_ALL)}
          className={cn(
            'flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors',
            props.value === FILTER_ALL
              ? 'bg-blue-600 font-medium text-white shadow-sm'
              : 'text-muted-foreground hover:bg-muted hover:text-foreground'
          )}
        >
          <Grid2X2 className='size-4 shrink-0' />
          <span className='min-w-0 flex-1 truncate'>{t('All Vendors')}</span>
          <span
            className={cn(
              'shrink-0 text-xs tabular-nums',
              props.value === FILTER_ALL
                ? 'text-white/85'
                : 'text-muted-foreground'
            )}
          >
            {props.models.length}
          </span>
        </button>

        {vendors.map((vendor) => {
          const active = props.value === vendor.name
          return (
            <button
              key={vendor.id ?? vendor.name}
              type='button'
              onClick={() => props.onChange(vendor.name)}
              className={cn(
                'flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-left text-sm transition-colors',
                active
                  ? 'bg-blue-50 font-medium text-blue-700 dark:bg-blue-500/15 dark:text-blue-300'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground'
              )}
            >
              <span className='flex size-4 shrink-0 items-center justify-center overflow-hidden'>
                {vendor.icon ? (
                  getLobeIcon(vendor.icon, 16)
                ) : (
                  <ProviderLogo name={vendor.name} size={16} />
                )}
              </span>
              <span className='min-w-0 flex-1 truncate'>{vendor.name}</span>
              <span className='text-muted-foreground shrink-0 text-xs tabular-nums'>
                {vendor.count}
              </span>
            </button>
          )
        })}
      </div>
    </aside>
  )
}
