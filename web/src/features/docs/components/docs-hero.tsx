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
import { Link } from '@tanstack/react-router'
import { ArrowRight, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'

import { HERO_BADGE_KEY } from '../constants'

/**
 * Top-of-page hero for the public docs. Mirrors the visual language used by
 * the home page hero (slate-on-blue accent, gradient banner background) but
 * stays static — no carousel, no auto-rotating slides.
 */
export function DocsHero() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const isAuthenticated = Boolean(auth.user)

  const primaryTarget = isAuthenticated ? '/dashboard' : '/sign-in'
  const primaryLabel = isAuthenticated
    ? t('Go to Dashboard')
    : t('Open the Console')

  return (
    <section className='relative isolate overflow-hidden'>
      {/* Banner artwork + readability gradient; reuses the same dual-layer
          overlay trick as the home hero so copy stays legible in both themes. */}
      <div
        aria-hidden
        className='absolute inset-0 -z-20 bg-[linear-gradient(120deg,#172554_0%,#1d4ed8_48%,#7c3aed_100%)]'
      />
      <div
        aria-hidden
        className='absolute inset-0 -z-10 bg-[linear-gradient(180deg,rgba(8,15,35,0.55),rgba(8,15,35,0.85))]'
      />
      <div
        aria-hidden
        className='absolute inset-x-0 bottom-0 h-24 bg-[linear-gradient(180deg,transparent,var(--background))]'
      />

      <div className='mx-auto flex max-w-[1400px] flex-col gap-10 px-6 pt-24 pb-16 sm:pt-32 lg:flex-row lg:items-center lg:pb-24'>
        <div className='landing-animate-fade-up max-w-2xl opacity-0'>
          <span className='border-white/20 bg-white/10 text-white inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium backdrop-blur-xs'>
            <Sparkles className='size-3.5' strokeWidth={1.8} />
            {t(HERO_BADGE_KEY)}
          </span>

          <h1 className='mt-5 text-[clamp(2rem,4vw,3.25rem)] leading-[1.15] font-bold tracking-tight text-white'>
            {t('docs.hero.title')}
          </h1>

          <p className='mt-4 max-w-xl text-base leading-relaxed text-balance text-blue-50/85 sm:text-lg'>
            {t('docs.hero.subtitle')}
          </p>

          <div className='mt-7 flex flex-wrap items-center gap-3'>
            <Button
              className='group h-11 rounded-lg bg-white px-6 text-sm font-semibold text-slate-900 shadow-[0_8px_24px_-8px_rgba(255,255,255,0.45)] hover:bg-blue-50'
              render={<Link to={primaryTarget} />}
            >
              {primaryLabel}
              <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
            </Button>
            <Button
              className='h-11 rounded-lg border-white/30 bg-white/0 px-6 text-sm font-medium text-white backdrop-blur-xs hover:bg-white/10'
              variant='outline'
              render={<Link to='/playground' />}
            >
              {t('Try it in the Playground')}
            </Button>
          </div>

          <p className='mt-6 text-xs text-blue-50/65'>
            {t('docs.hero.footnote')}
          </p>
        </div>

        {/* Right side: an "endpoint at a glance" card that previews the
            URL shape users will paste into their clients. */}
        <div className='landing-animate-fade-up w-full max-w-xl opacity-0 lg:ml-auto'>
          <div className='rounded-2xl border border-white/15 bg-slate-950/60 p-5 shadow-[0_30px_60px_-30px_rgba(15,23,42,0.6)] backdrop-blur-md'>
            <div className='text-blue-100/80 mb-3 flex items-center justify-between text-xs font-medium tracking-wider uppercase'>
              <span>{t('docs.hero.endpointPreview.title')}</span>
              <span className='rounded-full bg-emerald-400/15 px-2 py-0.5 text-[10px] font-medium text-emerald-200'>
                {t('docs.hero.endpointPreview.status')}
              </span>
            </div>
            <dl className='space-y-3 text-sm'>
              <EndpointRow
                label={t('docs.hero.endpointPreview.baseUrl')}
                value='/v1'
              />
              <EndpointRow
                label={t('docs.hero.endpointPreview.chatCompletions')}
                value='POST /v1/chat/completions'
              />
              <EndpointRow
                label={t('docs.hero.endpointPreview.completions')}
                value='POST /v1/completions'
              />
              <EndpointRow
                label={t('docs.hero.endpointPreview.embeddings')}
                value='POST /v1/embeddings'
              />
              <EndpointRow
                label={t('docs.hero.endpointPreview.images')}
                value='POST /v1/images/generations'
              />
            </dl>
          </div>
        </div>
      </div>
    </section>
  )
}

function EndpointRow({ label, value }: { label: string; value: string }) {
  return (
    <div className='flex items-center justify-between gap-4 rounded-lg border border-white/10 bg-white/5 px-3 py-2'>
      <dt className='text-xs text-blue-100/70'>{label}</dt>
      <dd className='truncate font-mono text-[13px] text-white'>{value}</dd>
    </div>
  )
}