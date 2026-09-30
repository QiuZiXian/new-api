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
import { KeyRound, Send, Settings2, type LucideIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

import { QUICK_START_STEPS } from '../constants'

const STEP_ICONS: Record<number, LucideIcon> = {
  1: KeyRound,
  2: Settings2,
  3: Send,
}

/**
 * Three-step "Quick Start" strip shown directly under the hero. Each card
 * is a numbered tile with a lucide icon, title and one-line description.
 * Content is sourced from translations to keep this file free of copy.
 */
export function QuickStart() {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 px-4 py-12 md:px-6 md:py-16'>
      <AnimateInView>
        <div className='mx-auto max-w-[1400px]'>
          <header className='mb-8 max-w-2xl'>
            <p className='text-blue-600 dark:text-blue-400 mb-2 text-sm font-semibold tracking-wider uppercase'>
              {t('docs.quickStart.eyebrow')}
            </p>
            <h2 className='text-foreground text-2xl font-bold tracking-tight sm:text-3xl'>
              {t('docs.quickStart.title')}
            </h2>
            <p className='text-muted-foreground mt-3 text-base text-balance'>
              {t('docs.quickStart.subtitle')}
            </p>
          </header>

          <ol className='grid grid-cols-1 gap-4 md:grid-cols-3'>
            {QUICK_START_STEPS.map((step) => {
              const Icon = STEP_ICONS[step.number]
              return (
                <li
                  key={step.number}
                  className='group relative flex flex-col gap-4 rounded-2xl border border-blue-100/80 bg-white/70 p-6 shadow-[0_2px_18px_-12px_rgba(37,99,235,0.25)] backdrop-blur-xs dark:border-blue-400/10 dark:bg-white/5'
                >
                  <div className='flex items-center gap-3'>
                    <span className='bg-blue-600/10 text-blue-600 dark:bg-blue-400/15 dark:text-blue-300 inline-flex size-9 items-center justify-center rounded-xl text-sm font-semibold'>
                      {step.number}
                    </span>
                    <span className='bg-blue-600/10 text-blue-600 dark:bg-blue-400/15 dark:text-blue-300 inline-flex size-9 items-center justify-center rounded-xl transition-colors group-hover:bg-blue-600/15'>
                      <Icon className='size-5' strokeWidth={1.6} />
                    </span>
                  </div>

                  <h3 className='text-foreground text-base leading-snug font-semibold tracking-tight sm:text-lg'>
                    {t(step.titleKey)}
                  </h3>
                  <p className='text-muted-foreground text-sm leading-relaxed text-balance'>
                    {t(step.descriptionKey)}
                  </p>
                </li>
              )
            })}
          </ol>
        </div>
      </AnimateInView>
    </section>
  )
}