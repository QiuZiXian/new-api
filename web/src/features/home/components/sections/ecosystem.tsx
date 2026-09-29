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
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import { Button } from '@/components/ui/button'
import { Link } from '@tanstack/react-router'

import { ECOSYSTEM_APPS } from '../../constants'
import { ScrollingIcons } from '../scrolling-icons'

/**
 * Second screen — Application Ecosystem / Gateway showcase.
 * A central gateway card flanked by two reverse-scrolling app icon walls,
 * echoing the "app marketplace" aggregation feel of the reference sites while
 * keeping the existing blue-purple tech aesthetic.
 */
export function Ecosystem() {
  const { t } = useTranslation()
  const isAuthenticated = false

  return (
    <section className='relative z-10 overflow-hidden border-t border-border/40 px-6 py-24 md:py-32'>
      {/* Ambient radial glow */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-20 dark:opacity-[0.08]'
        style={{
          background: [
            'radial-gradient(ellipse 55% 45% at 50% 30%, oklch(0.72 0.18 250 / 70%) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 40% at 80% 80%, oklch(0.65 0.15 200 / 50%) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <AnimateInView className='mx-auto mb-12 max-w-2xl text-center md:mb-16'>
        <p className='text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase'>
          {t('Application Ecosystem')}
        </p>
        <h2 className='text-2xl leading-tight font-bold tracking-tight md:text-3xl'>
          {t('One gateway, connected to')}
          <br />
          <span className='from-blue-400 via-violet-400 to-purple-500 bg-gradient-to-r bg-clip-text text-transparent'>
            {t('the tools you already use')}
          </span>
        </h2>
      </AnimateInView>

      <div className='mx-auto grid max-w-6xl grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-5'>
        <ScrollingIcons
          icons={ECOSYSTEM_APPS}
          direction='up'
          className='-my-2 h-[420px]'
        />
        <div className='flex min-w-0 flex-col items-center justify-center gap-3'>
          <Button
            className='group h-[52px] w-full max-w-[280px] rounded-lg bg-gradient-to-r from-blue-600 to-blue-500 px-6 text-base font-medium text-white shadow-[0_10px_24px_-8px_rgba(37,99,235,0.55)] hover:from-blue-700 hover:to-blue-600'
            render={
              isAuthenticated ? <Link to='/dashboard' /> : <Link to='/sign-up' />
            }
          >
            {t('Register now')}
            <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
          </Button>
          <Button
            className='h-[52px] w-full max-w-[280px] rounded-lg border-blue-200 px-6 text-base font-medium text-blue-600 shadow-sm hover:border-blue-300 hover:bg-blue-50 dark:border-blue-500/40 dark:text-blue-400 dark:hover:bg-blue-500/10'
            variant='outline'
            render={
              isAuthenticated ? (
                <Link to='/dashboard' />
              ) : (
                <Link to='/sign-in' />
              )
            }
          >
            {t('Get Started')}
          </Button>
          <Button
            className='h-[52px] w-full max-w-[280px] rounded-lg border-border/50 px-6 text-base font-medium hover:bg-muted/50'
            variant='outline'
            render={<Link to='/pricing' />}
          >
            {t('View Pricing')}
          </Button>
        </div>
        <ScrollingIcons
          icons={ECOSYSTEM_APPS}
          direction='down'
          className='-my-2 h-[420px]'
        />
      </div>
    </section>
  )
}