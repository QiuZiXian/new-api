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
import { useTranslation } from 'react-i18next'

import { Skeleton } from '@/components/ui/skeleton'
import { useSystemConfig } from '@/hooks/use-system-config'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()

  return (
    <div className='relative grid h-svh max-w-none'>
      {/* Ambient blue-purple glow + grid backdrop */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 overflow-hidden'
      >
        <div
          className='absolute inset-0 opacity-25 dark:opacity-[0.14]'
          style={{
            background: [
              'radial-gradient(ellipse 55% 50% at 22% 20%, oklch(0.72 0.18 250 / 80%) 0%, transparent 70%)',
              'radial-gradient(ellipse 45% 40% at 82% 18%, oklch(0.65 0.15 200 / 60%) 0%, transparent 70%)',
              'radial-gradient(ellipse 40% 36% at 45% 82%, oklch(0.70 0.12 280 / 40%) 0%, transparent 70%)',
            ].join(', '),
          }}
        />
        <div className='absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_70%_60%_at_50%_40%,black_15%,transparent_100%)] bg-[size:4rem_4rem] opacity-[0.07]' />
      </div>

      <Link
        to='/'
        className='absolute top-4 left-4 z-10 flex items-center gap-2 transition-opacity hover:opacity-80 sm:top-8 sm:left-8'
      >
        <div className='relative h-8 w-8'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-full' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='h-8 w-8 rounded-full object-cover'
            />
          )}
        </div>
        {loading ? (
          <Skeleton className='h-6 w-24' />
        ) : (
          <h1 className='text-xl font-medium'>{systemName}</h1>
        )}
      </Link>
      <div className='container flex items-center pt-16 sm:pt-0'>
        <div className='glass-1 relative mx-auto flex w-full flex-col justify-center space-y-2 border-border/50 px-6 py-8 shadow-[0_20px_60px_-25px_rgba(15,23,42,0.18)] sm:w-[480px] sm:rounded-2xl sm:p-8 dark:shadow-[0_20px_60px_-25px_rgba(0,0,0,0.6)]'>
          {children}
        </div>
      </div>
    </div>
  )
}
