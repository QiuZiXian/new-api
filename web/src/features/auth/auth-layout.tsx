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
import { KeyRound, BarChart3, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
}

function BrandFeature(props: {
  icon: React.ReactNode
  title: string
  desc: string
}) {
  return (
    <div className='flex items-start gap-3'>
      <span className='flex size-9 shrink-0 items-center justify-center rounded-lg bg-white/15 text-white backdrop-blur-xs'>
        {props.icon}
      </span>
      <div className='min-w-0'>
        <p className='text-sm font-semibold text-white'>{props.title}</p>
        <p className='mt-0.5 text-xs leading-relaxed text-blue-100/85'>
          {props.desc}
        </p>
      </div>
    </div>
  )
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='grid min-h-svh max-w-none lg:grid-cols-2'>
      {/* Left: brand panel (desktop only) */}
      <aside className='relative hidden overflow-hidden bg-[linear-gradient(160deg,#1d4ed8_0%,#2563eb_45%,#0ea5e9_100%)] lg:flex lg:flex-col'>
        {/* Decorative grid + glows */}
        <div
          aria-hidden
          className='absolute inset-0 bg-[linear-gradient(to_right,rgba(255,255,255,0.08)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.08)_1px,transparent_1px)] bg-[size:3rem_3rem] [mask-image:radial-gradient(ellipse_75%_65%_at_45%_40%,black_25%,transparent_100%)]'
        />
        <div
          aria-hidden
          className='absolute -top-24 -right-16 size-80 rounded-full bg-cyan-300/25 blur-3xl'
        />
        <div
          aria-hidden
          className='absolute bottom-0 -left-20 size-96 rounded-full bg-blue-300/20 blur-3xl'
        />
        <svg
          aria-hidden
          viewBox='0 0 480 300'
          className='absolute inset-x-0 bottom-0 w-full opacity-40'
          fill='none'
        >
          <g stroke='rgba(255,255,255,0.5)' strokeWidth='1'>
            <path d='M-10 240h150l30 30h130' />
            <path d='M490 210h-130l-36 36h-110' />
          </g>
          <g fill='rgba(255,255,255,0.7)'>
            <circle cx='170' cy='240' r='2.5' />
            <circle cx='214' cy='270' r='2.5' />
            <circle cx='330' cy='246' r='2.5' />
          </g>
        </svg>

        {/* Brand */}
        <Link
          to='/'
          className='relative z-10 flex items-center gap-2.5 px-10 pt-9 transition-opacity hover:opacity-85'
        >
          <div className='relative size-9'>
            {loading ? (
              <Skeleton className='absolute inset-0 rounded-full' />
            ) : (
              <img
                src={logo}
                alt={t('Logo')}
                className='size-9 rounded-full bg-white object-cover'
              />
            )}
          </div>
          {loading ? (
            <Skeleton className='h-6 w-28' />
          ) : (
            <span className='text-lg font-semibold text-white'>
              {systemName}
            </span>
          )}
        </Link>

        {/* Value proposition */}
        <div className='relative z-10 flex flex-1 flex-col justify-center gap-8 px-10 py-12'>
          <div>
            <h2 className='text-3xl leading-snug font-bold tracking-tight text-white'>
              {t('Welcome back')}
            </h2>
            <p className='mt-3 max-w-md text-sm leading-relaxed text-blue-100/90'>
              {t(
                'Sign in to manage your API keys, model usage, billing and enterprise AI access.'
              )}
            </p>
          </div>

          <div className='space-y-5'>
            <BrandFeature
              icon={<KeyRound className='size-4' />}
              title={t('Unified model access')}
              desc={t('One key for text, image, video and multimodal models.')}
            />
            <BrandFeature
              icon={<ShieldCheck className='size-4' />}
              title={t('Enterprise-grade access control')}
              desc={t('Quotas, expiry, model limits and IP allowlists.')}
            />
            <BrandFeature
              icon={<BarChart3 className='size-4' />}
              title={t('Usage & cost visibility')}
              desc={t(
                'Track tokens, requests, costs and task history in real time.'
              )}
            />
          </div>
        </div>

        <p className='relative z-10 px-10 pb-8 text-xs text-blue-100/70'>
          © {new Date().getFullYear()} {loading ? <Skeleton className='inline-block h-3 w-24' /> : systemName}
        </p>
      </aside>

      {/* Right: form area */}
      <main className='relative flex flex-col'>
        {/* Mobile top bar with logo */}
        <div className='flex items-center gap-2 px-4 pt-5 lg:hidden'>
          <Link
            to='/'
            className='flex items-center gap-2 transition-opacity hover:opacity-80'
          >
            <div className='relative size-8'>
              {loading ? (
                <Skeleton className='absolute inset-0 rounded-full' />
              ) : (
                <img
                  src={logo}
                  alt={t('Logo')}
                  className='size-8 rounded-full object-cover'
                />
              )}
            </div>
            {loading ? (
              <Skeleton className='h-5 w-20' />
            ) : (
              <span className='text-base font-medium'>{systemName}</span>
            )}
          </Link>
        </div>

        <div className='container flex flex-1 items-center justify-center py-10 lg:py-0'>
          <div className='glass-1 relative mx-auto flex w-full flex-col justify-center space-y-2 border-border/50 px-6 py-8 shadow-[0_20px_60px_-25px_rgba(15,23,42,0.18)] sm:w-[440px] sm:rounded-2xl sm:p-8 dark:shadow-[0_20px_60px_-25px_rgba(0,0,0,0.6)]'>
            {children}
          </div>
        </div>
      </main>
    </div>
  )
}
