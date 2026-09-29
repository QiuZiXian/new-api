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

import { Skeleton } from '@/components/ui/skeleton'

interface PricingHeroProps {
  total: number
  loading?: boolean
}

export function PricingHero(props: PricingHeroProps) {
  const { t } = useTranslation()

  return (
    <section className='relative isolate overflow-hidden'>
      {/* Banner artwork from design assets */}
      <div aria-hidden className='absolute inset-0 -z-20'>
        <img
          src='/home/models-banner.png'
          alt=''
          className='absolute inset-0 h-full w-full object-cover object-[75%_center] sm:object-[80%_center]'
        />
        {/* Readability overlay: darken behind copy, fade out over artwork */}
        <div className='absolute inset-0 bg-[linear-gradient(90deg,rgba(238,244,255,0.9)_0%,rgba(238,244,255,0.5)_40%,rgba(238,244,255,0)_70%)] dark:bg-[linear-gradient(90deg,rgba(8,15,35,0.92)_0%,rgba(8,15,35,0.55)_40%,rgba(8,15,35,0.12)_70%)]' />
        <div className='absolute inset-x-0 bottom-0 h-20 bg-[linear-gradient(180deg,transparent,var(--background))]' />
      </div>

      <div className='mx-auto flex max-w-[1800px] flex-col items-start gap-6 px-4 pt-24 pb-12 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:pt-32 sm:pb-14 xl:px-8'>
        <div className='landing-animate-fade-up min-w-0 opacity-0'>
          <h1 className='text-slate-900 text-[clamp(1.75rem,3vw,2.5rem)] leading-tight font-bold tracking-tight dark:text-slate-50'>
            {t('Model Square')}
          </h1>
          <p className='text-slate-600 mt-2 text-sm md:text-base dark:text-slate-400'>
            {t('Latest & most comprehensive | Model resource hub')}
          </p>
          <p className='text-muted-foreground mt-2 text-xs md:text-sm'>
            {props.loading ? (
              <Skeleton className='h-4 w-40' />
            ) : (
              t('{{count}} models available now', { count: props.total })
            )}
          </p>
        </div>
      </div>
    </section>
  )
}
