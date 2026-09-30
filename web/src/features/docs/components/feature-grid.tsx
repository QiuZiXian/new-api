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
import {
  Blocks,
  Cable,
  CircuitBoard,
  Languages,
  ShieldCheck,
  Workflow,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

import { FEATURE_GRID_ITEMS } from '../constants'
import type { FeatureGridIconId } from '../types'

const FEATURE_ICONS: Record<FeatureGridIconId, LucideIcon> = {
  multiModel: Blocks,
  openAICompat: Cable,
  billing: CircuitBoard,
  routing: Workflow,
  formatConvert: Languages,
  permission: ShieldCheck,
}

/**
 * Six headline capability tiles shown at the bottom of the docs page.
 * Icons come from `lucide-react`; copy is read from translations.
 */
export function FeatureGrid() {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 px-4 py-12 md:px-6 md:py-16'>
      <AnimateInView>
        <div className='mx-auto max-w-[1400px]'>
          <header className='mb-8 max-w-2xl'>
            <p className='text-blue-600 dark:text-blue-400 mb-1 text-sm font-semibold tracking-wider uppercase'>
              {t('docs.features.eyebrow')}
            </p>
            <h2 className='text-foreground text-2xl font-bold tracking-tight sm:text-3xl'>
              {t('docs.features.title')}
            </h2>
            <p className='text-muted-foreground mt-3 text-base text-balance'>
              {t('docs.features.subtitle')}
            </p>
          </header>

          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3'>
            {FEATURE_GRID_ITEMS.map((item) => {
              const Icon = FEATURE_ICONS[item.id]
              return (
                <article
                  key={item.id}
                  className='group border-border/60 bg-card hover:border-accent-foreground/30 flex flex-col gap-3 rounded-2xl border p-5 transition-colors'
                >
                  <span className='bg-blue-600/10 text-blue-600 dark:bg-blue-400/15 dark:text-blue-300 inline-flex size-10 items-center justify-center rounded-xl transition-colors group-hover:bg-blue-600/15'>
                    <Icon className='size-5' strokeWidth={1.6} />
                  </span>
                  <h3 className='text-foreground text-base font-semibold tracking-tight'>
                    {t(item.titleKey)}
                  </h3>
                  <p className='text-muted-foreground text-sm leading-relaxed text-balance'>
                    {t(item.descriptionKey)}
                  </p>
                </article>
              )
            })}
          </div>
        </div>
      </AnimateInView>
    </section>
  )
}