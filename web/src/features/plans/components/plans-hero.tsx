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
import { Clock3, Headphones, ShieldCheck, TrendingUp } from 'lucide-react'
import { useTranslation } from 'react-i18next'

const HIGHLIGHTS = [
  {
    icon: TrendingUp,
    title: 'Flexible scaling',
    description: 'Choose capacity on demand',
  },
  {
    icon: Clock3,
    title: 'Instant activation',
    description: 'Credited immediately after purchase',
  },
  {
    icon: ShieldCheck,
    title: 'Secure and reliable',
    description: 'Enterprise-grade data protection',
  },
  {
    icon: Headphones,
    title: 'Professional support',
    description: '24/7 service',
  },
] as const

export function PlansHero() {
  const { t } = useTranslation()

  return (
    <section className='relative isolate overflow-hidden'>
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(70%_90%_at_50%_-10%,rgba(59,130,246,0.16),transparent_65%)] dark:bg-[radial-gradient(70%_90%_at_50%_-10%,rgba(59,130,246,0.22),transparent_65%)]'
      />
      <div className='mx-auto w-full max-w-5xl px-4 pt-24 pb-6 text-center sm:pt-28'>
        <h1 className='text-3xl leading-tight font-bold tracking-tight sm:text-[2.75rem]'>
          {t('Simple and transparent')}{' '}
          <span className='bg-gradient-to-r from-sky-500 via-blue-500 to-blue-600 bg-clip-text text-transparent'>
            {t('Pay as you go')}
          </span>
        </h1>
        <p className='text-muted-foreground mx-auto mt-3 max-w-2xl text-sm'>
          {t(
            'Choose the plan that fits you — no hidden fees, no minimum spend'
          )}
        </p>
      </div>

      <div className='mx-auto w-full max-w-5xl px-4 pt-2 pb-8'>
        <div className='grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-4'>
          {HIGHLIGHTS.map((item) => {
            const Icon = item.icon
            return (
              <div
                key={item.title}
                className='flex flex-col items-center text-center'
              >
                <div className='flex size-10 items-center justify-center rounded-xl bg-sky-50 text-sky-600 dark:bg-sky-500/10 dark:text-sky-400'>
                  <Icon className='size-5' />
                </div>
                <p className='mt-2.5 text-[13px] font-medium'>
                  {t(item.title)}
                </p>
                <p className='text-muted-foreground mt-0.5 text-[11px]'>
                  {t(item.description)}
                </p>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
