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
import { ArrowRight, Image, MessageSquareText, Search, Video } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'

import {
  HERO_CHIP_ARTWORK_URL,
  HERO_LANDSCAPE_ARTWORK_URL,
} from '../lib/artwork'

interface PricingHeroProps {
  total: number
  loading?: boolean
  search?: string
  onSearchChange?: (value: string) => void
}

const QUICK_LINKS = [
  { key: 'Chat & Reasoning', keyword: '', icon: MessageSquareText },
  { key: 'Image Creation', keyword: 'image', icon: Image },
  { key: 'Video Generation', keyword: 'video', icon: Video },
] as const

export function PricingHero(props: PricingHeroProps) {
  const { t } = useTranslation()

  return (
    <section className='relative isolate overflow-hidden border-b border-blue-100/70 bg-[linear-gradient(110deg,#f4f8ff_0%,#eaf2ff_45%,#f7fbff_100%)] dark:border-white/10 dark:bg-[linear-gradient(110deg,#0b1224_0%,#101d3a_55%,#0b1224_100%)]'>
      {/* Space for the fixed h-16 / md:h-20 header: keep this banner pinned
          directly below the nav bar instead of sliding underneath it. */}
      <div aria-hidden className='h-16 md:h-20' />

      <div className='pointer-events-none absolute -top-24 right-1/4 -z-10 size-72 rounded-full bg-blue-300/25 blur-3xl dark:bg-blue-500/10' />

      {/* Fixed desktop height matches the previous banner (header space +
          230px content = ~310px total). */}
      <div className='mx-auto grid max-w-[1400px] items-center gap-5 px-4 py-6 sm:px-6 lg:h-[230px] lg:grid-cols-[minmax(0,1fr)_minmax(0,0.95fr)] lg:py-0 xl:px-8'>
        <div className='landing-animate-fade-up min-w-0'>
          <p className='text-[10px] font-semibold tracking-[0.22em] text-blue-600 uppercase dark:text-blue-400'>
            {t('AI MODEL MARKETPLACE')}
          </p>
          <h1 className='text-foreground mt-2 text-[clamp(1.5rem,2.3vw,2.1rem)] leading-[1.15] font-bold tracking-tight'>
            {t('Find your AI model')}
            <br />
            <span className='bg-gradient-to-r from-blue-600 via-indigo-500 to-sky-500 bg-clip-text text-transparent dark:from-blue-400 dark:via-indigo-300 dark:to-sky-400'>
              {t('by provider')}
            </span>
          </h1>
          <p className='text-muted-foreground mt-1.5 text-xs leading-relaxed'>
            {t('Chat, image and video models, all connected with one API.')}
          </p>

          <div className='mt-3 flex max-w-xl items-center gap-2 rounded-full border border-blue-100 bg-white/90 p-1 pl-3.5 shadow-[0_6px_20px_-10px_rgba(37,99,235,0.5)] backdrop-blur-sm dark:border-white/10 dark:bg-white/5'>
            <Search className='text-muted-foreground size-3.5 shrink-0' />
            <input
              type='text'
              value={props.search ?? ''}
              onChange={(event) => props.onSearchChange?.(event.target.value)}
              placeholder={t('Search models, vendors or capabilities')}
              aria-label={t('Search')}
              className='placeholder:text-muted-foreground/70 min-w-0 flex-1 border-0 bg-transparent text-sm outline-none focus-visible:ring-0'
            />
            <button
              type='button'
              aria-label={t('Search')}
              onClick={() => props.onSearchChange?.(props.search ?? '')}
              className='flex size-7 shrink-0 items-center justify-center rounded-full bg-blue-600 text-white transition-colors hover:bg-blue-700'
            >
              <ArrowRight className='size-3.5' />
            </button>
          </div>

          <p className='text-muted-foreground mt-2 text-[11px]'>
            {props.loading ? (
              <Skeleton className='h-3.5 w-40' />
            ) : (
              t('{{count}} models available now', { count: props.total })
            )}
          </p>
        </div>

        {/* 让 chip 图向左溢出容器，避免与整幅背景图完全叠在一起；
            圆角改由背景图自己裁剪，容器不再 overflow-hidden。 */}
        <div className='relative hidden h-[190px] rounded-2xl border border-white/70 shadow-[0_16px_36px_-20px_rgba(37,99,235,0.55)] lg:block dark:border-white/10'>
          <img
            src={HERO_LANDSCAPE_ARTWORK_URL}
            alt=''
            referrerPolicy='no-referrer'
            className='absolute inset-0 h-full w-full rounded-2xl object-cover'
          />
          <div className='absolute inset-0 rounded-2xl bg-[linear-gradient(90deg,rgba(238,244,255,0.5)_0%,rgba(238,244,255,0.05)_55%,transparent_100%)] dark:bg-[linear-gradient(90deg,rgba(8,15,35,0.5)_0%,rgba(8,15,35,0.05)_55%,transparent_100%)]' />
          <img
            src={HERO_CHIP_ARTWORK_URL}
            alt=''
            referrerPolicy='no-referrer'
            className='absolute top-3 -left-8 h-[86px] w-[86px] rounded-xl object-cover shadow-lg ring-1 ring-white/70'
          />
          <div className='absolute right-3 bottom-3 left-3 flex flex-wrap gap-1.5'>
            {QUICK_LINKS.map((link) => {
              const Icon = link.icon
              return (
                <button
                  key={link.key}
                  type='button'
                  onClick={() => props.onSearchChange?.(link.keyword)}
                  className='inline-flex items-center gap-1 rounded-full border border-white/70 bg-white/85 px-2.5 py-1 text-[11px] font-medium text-slate-700 shadow-sm backdrop-blur-sm transition-colors hover:bg-white'
                >
                  <Icon className='size-3' />
                  {t(link.key)}
                </button>
              )
            })}
          </div>
        </div>
      </div>
    </section>
  )
}
