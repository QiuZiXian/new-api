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

import { HERO_ARTWORK_URL } from '../lib/artwork'

interface PricingHeroProps {
  total: number
  loading?: boolean
  search?: string
  onSearchChange?: (value: string) => void
}

const QUICK_LINKS = [
  {
    key: 'Chat & Reasoning',
    keyword: '',
    icon: MessageSquareText,
    description: 'Chat, reasoning and coding models',
  },
  {
    key: 'Image Creation',
    keyword: 'image',
    icon: Image,
    description: 'Image generation and editing',
  },
  {
    key: 'Video Generation',
    keyword: 'video',
    icon: Video,
    description: 'Text to video, one API call',
  },
] as const

export function PricingHero(props: PricingHeroProps) {
  const { t } = useTranslation()

  return (
    <section className='relative isolate overflow-hidden border-b border-blue-100/70 bg-[linear-gradient(110deg,#f4f8ff_0%,#eaf2ff_45%,#f7fbff_100%)] dark:border-white/10 dark:bg-[linear-gradient(110deg,#0b1224_0%,#101d3a_55%,#0b1224_100%)]'>
      {/* Space for the fixed h-16 / md:h-20 header: keep this banner pinned
          directly below the nav bar instead of sliding underneath it. */}
      <div aria-hidden className='h-16 md:h-20' />

      <div
        aria-hidden
        className='pointer-events-none absolute inset-y-0 right-0 -z-10 hidden w-1/2 lg:block'
      >
        <img
          src={HERO_ARTWORK_URL}
          alt=''
          referrerPolicy='no-referrer'
          className='h-full w-full object-cover object-center opacity-30'
        />
        <div className='absolute inset-0 bg-[linear-gradient(90deg,#f4f8ff_0%,rgba(244,248,255,0.55)_45%,transparent_100%)] dark:bg-[linear-gradient(90deg,#0b1224_0%,rgba(11,18,36,0.55)_45%,transparent_100%)]' />
      </div>
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

        <div className='hidden gap-2.5 lg:grid lg:grid-cols-3'>
          {QUICK_LINKS.map((link) => {
            const Icon = link.icon
            return (
              <button
                key={link.key}
                type='button'
                onClick={() => props.onSearchChange?.(link.keyword)}
                className='group border-border/60 bg-background/80 hover:border-blue-400/60 hover:bg-background flex min-h-[72px] flex-col items-start gap-1 rounded-xl border p-2.5 text-left shadow-[0_4px_16px_-12px_rgba(15,23,42,0.5)] backdrop-blur-sm transition-all hover:-translate-y-0.5 dark:border-white/10 dark:bg-white/5'
              >
                <span className='flex items-center gap-1.5 text-xs font-semibold'>
                  <span className='flex size-5 items-center justify-center rounded-md bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300'>
                    <Icon className='size-3' />
                  </span>
                  {t(link.key)}
                </span>
                <span className='text-muted-foreground line-clamp-2 text-[10px] leading-snug'>
                  {t(link.description)}
                </span>
              </button>
            )
          })}
          <div className='flex items-center justify-between gap-3 rounded-xl border border-blue-200/70 bg-[linear-gradient(100deg,#dbeafe,#eff6ff)] px-3 py-2.5 lg:col-span-3 dark:border-white/10 dark:bg-[linear-gradient(100deg,rgba(30,64,175,0.25),rgba(15,23,42,0.25))]'>
            <div className='min-w-0'>
              <p className='text-foreground text-xs font-semibold'>
                {t('One API for every model')}
              </p>
              <p className='text-muted-foreground truncate text-[10px]'>
                {t('Switch providers without changing your integration.')}
              </p>
            </div>
            <span className='shrink-0 text-[10px] font-medium text-blue-600 dark:text-blue-300'>
              POST /v1/chat/completions
            </span>
          </div>
        </div>
      </div>
    </section>
  )
}
