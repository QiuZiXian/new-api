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

/** Isometric cube stack illustration for the banner. */
function CubeArt() {
  return (
    <svg
      aria-hidden
      viewBox='0 0 300 240'
      fill='none'
      xmlns='http://www.w3.org/2000/svg'
      className='h-auto w-full max-w-[300px]'
      role='presentation'
    >
      <defs>
        <linearGradient id='cube-top' x1='0' y1='0' x2='1' y2='1'>
          <stop offset='0%' stopColor='#dbeafe' />
          <stop offset='100%' stopColor='#93c5fd' />
        </linearGradient>
        <linearGradient id='cube-left' x1='0' y1='0' x2='0' y2='1'>
          <stop offset='0%' stopColor='#3b82f6' />
          <stop offset='100%' stopColor='#1d4ed8' />
        </linearGradient>
        <linearGradient id='cube-right' x1='0' y1='0' x2='0' y2='1'>
          <stop offset='0%' stopColor='#60a5fa' />
          <stop offset='100%' stopColor='#2563eb' />
        </linearGradient>
        <radialGradient id='cube-glow' cx='50%' cy='50%' r='50%'>
          <stop offset='0%' stopColor='#93c5fd' stopOpacity='0.5' />
          <stop offset='100%' stopColor='#93c5fd' stopOpacity='0' />
        </radialGradient>
      </defs>
      <ellipse cx='150' cy='130' rx='140' ry='100' fill='url(#cube-glow)' />
      {/* Base platform */}
      <g>
        <polygon points='150,168 250,124 150,80 50,124' fill='url(#cube-top)' />
        <polygon points='50,124 150,168 150,196 50,152' fill='url(#cube-left)' opacity='0.85' />
        <polygon points='250,124 150,168 150,196 250,152' fill='url(#cube-right)' opacity='0.9' />
      </g>
      {/* Stacked cubes */}
      <g>
        <polygon points='150,124 206,100 150,76 94,100' fill='url(#cube-top)' />
        <polygon points='94,100 150,124 150,168 94,144' fill='url(#cube-left)' />
        <polygon points='206,100 150,124 150,168 206,144' fill='url(#cube-right)' />
      </g>
      <g>
        <polygon points='150,88 190,71 150,54 110,71' fill='url(#cube-top)' />
        <polygon points='110,71 150,88 150,124 110,107' fill='url(#cube-left)' />
        <polygon points='190,71 150,88 150,124 190,107' fill='url(#cube-right)' />
      </g>
      {/* Accent ring + sparks */}
      <ellipse cx='150' cy='124' rx='104' ry='46' fill='none' stroke='#60a5fa' strokeOpacity='0.5' strokeDasharray='4 8' />
      <circle cx='262' cy='92' r='4' fill='#60a5fa' opacity='0.85' />
      <circle cx='42' cy='150' r='3' fill='#3b82f6' opacity='0.7' />
      <circle cx='196' cy='40' r='3.5' fill='#93c5fd' opacity='0.9' />
      <circle cx='84' cy='56' r='2.5' fill='#22d3ee' opacity='0.8' />
    </svg>
  )
}

export function PricingHero(props: PricingHeroProps) {
  const { t } = useTranslation()

  return (
    <section className='relative isolate overflow-hidden'>
      {/* Tech-blue gradient banner */}
      <div
        aria-hidden
        className='absolute inset-0 -z-20 bg-[linear-gradient(180deg,#e8f1ff_0%,#e2edff_60%,var(--background)_100%)] dark:bg-[linear-gradient(180deg,#0a1730_0%,#0c1c3a_60%,var(--background)_100%)]'
      />
      {/* Circuit traces */}
      <svg
        aria-hidden
        viewBox='0 0 1440 260'
        preserveAspectRatio='xMidYMin slice'
        className='absolute inset-x-0 top-0 -z-10 h-full w-full opacity-40 dark:opacity-20'
      >
        <g stroke='#93c5fd' strokeWidth='1' fill='none'>
          <path d='M-20 60h260l36 36h200' />
          <path d='M-20 130h180l46 46h240' opacity='0.7' />
          <path d='M1460 50h-240l-40 40h-220' />
          <path d='M1460 170h-200l-56 56h-160' opacity='0.7' />
        </g>
        <g fill='#60a5fa'>
          <circle cx='476' cy='96' r='3' />
          <circle cx='446' cy='176' r='3' />
          <circle cx='960' cy='90' r='3' />
          <circle cx='1044' cy='226' r='3' />
        </g>
      </svg>
      <div
        aria-hidden
        className='absolute -top-16 right-[16%] -z-10 size-72 rounded-full bg-blue-400/25 blur-3xl dark:bg-blue-500/10'
      />

      <div className='mx-auto flex max-w-[1800px] flex-col items-start gap-6 px-4 pt-24 pb-12 sm:flex-row sm:items-center sm:justify-between sm:px-6 sm:pt-28 sm:pb-14 xl:px-8'>
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
        <div className='landing-animate-fade-up hidden shrink-0 opacity-0 md:block'>
          <CubeArt />
        </div>
      </div>
    </section>
  )
}
