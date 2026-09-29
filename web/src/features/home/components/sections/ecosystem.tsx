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
import { Rocket, Tags, UserPlus } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import { Button } from '@/components/ui/button'
import { Link } from '@tanstack/react-router'

/**
 * Second screen — gateway headline with the three primary entry actions.
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

      <AnimateInView className='mx-auto mb-10 max-w-2xl text-center md:mb-12'>
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

      <AnimateInView
        className='flex flex-col items-center justify-center gap-4 sm:flex-row sm:gap-5'
        animation='fade-up'
      >
        <Button
          className='group h-[52px] w-full rounded-xl bg-gradient-to-r from-blue-600 to-blue-500 px-7 text-base font-medium text-white shadow-[0_12px_28px_-10px_rgba(37,99,235,0.65)] transition-all duration-300 hover:-translate-y-0.5 hover:from-blue-700 hover:to-blue-600 hover:shadow-[0_18px_36px_-12px_rgba(37,99,235,0.75)] sm:w-auto'
          render={
            isAuthenticated ? <Link to='/dashboard' /> : <Link to='/sign-up' />
          }
        >
          <UserPlus className='mr-1.5 size-4' />
          {t('Register now')}
        </Button>
        <Button
          className='h-[52px] w-full rounded-xl border-blue-200 bg-white px-7 text-base font-medium text-blue-600 shadow-[0_10px_24px_-12px_rgba(37,99,235,0.35)] transition-all duration-300 hover:-translate-y-0.5 hover:border-blue-300 hover:bg-blue-50 hover:shadow-[0_16px_30px_-12px_rgba(37,99,235,0.45)] sm:w-auto dark:border-blue-500/40 dark:bg-blue-500/10 dark:text-blue-300 dark:hover:bg-blue-500/20'
          variant='outline'
          render={
            isAuthenticated ? <Link to='/dashboard' /> : <Link to='/sign-in' />
          }
        >
          <Rocket className='mr-1.5 size-4' />
          {t('Get Started')}
        </Button>
        <Button
          className='h-[52px] w-full rounded-xl border-border/60 bg-white/70 px-7 text-base font-medium shadow-[0_10px_24px_-14px_rgba(15,23,42,0.25)] backdrop-blur-sm transition-all duration-300 hover:-translate-y-0.5 hover:border-border hover:bg-white hover:shadow-[0_16px_30px_-14px_rgba(15,23,42,0.3)] sm:w-auto dark:bg-white/5'
          variant='outline'
          render={<Link to='/pricing' />}
        >
          <Tags className='mr-1.5 size-4' />
          {t('View Pricing')}
        </Button>
      </AnimateInView>
    </section>
  )
}