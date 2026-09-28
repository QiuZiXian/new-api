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
import { KeyRound, Route, Wallet, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const capabilities = [
    {
      icon: <KeyRound className='size-5' strokeWidth={1.6} />,
      title: t('Unified access'),
      desc: t(
        'One API key calls text, image, video and multimodal models, cutting repeated vendor integration costs.'
      ),
    },
    {
      icon: <Wallet className='size-5' strokeWidth={1.6} />,
      title: t('Predictable cost'),
      desc: t(
        'Pay-as-you-go and per-request billing with quota limits and usage stats keep every call accountable.'
      ),
    },
    {
      icon: <ShieldCheck className='size-5' strokeWidth={1.6} />,
      title: t('Permission control'),
      desc: t(
        'Key-level quotas, expiry, model restrictions and IP allowlists fit per-project and per-team governance.'
      ),
    },
    {
      icon: <Route className='size-5' strokeWidth={1.6} />,
      title: t('Stable routing'),
      desc: t(
        'Unified scheduling across multiple channels reduces the impact of a single model or account outage.'
      ),
    },
  ]

  return (
    <section className='relative z-10 px-6 py-20 md:py-28'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-12 text-center md:mb-16'>
          <p className='text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase'>
            {t('Enterprise Capabilities')}
          </p>
          <h2 className='text-2xl font-bold tracking-tight md:text-3xl'>
            {t('Designed for enterprise-grade AI integration')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-3 max-w-2xl text-sm leading-relaxed'>
            {t(
              'From model calls to usage management, teams get unified access, unified billing and unified governance.'
            )}
          </p>
        </AnimateInView>

        <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4'>
          {capabilities.map((capability, i) => (
            <AnimateInView
              key={capability.title}
              delay={i * 120}
              animation='fade-up'
              className='group border-border/50 bg-card hover:border-blue-500/30 hover:shadow-[0_16px_40px_-20px_rgba(37,99,235,0.35)] relative flex flex-col rounded-2xl border p-6 transition-all duration-300'
            >
              <span className='mb-5 inline-flex size-11 items-center justify-center rounded-xl bg-blue-500/10 text-blue-600 transition-colors group-hover:bg-blue-500/15 dark:bg-blue-400/10 dark:text-blue-400'>
                {capability.icon}
              </span>
              <h3 className='mb-2 text-base font-semibold'>
                {capability.title}
              </h3>
              <p className='text-muted-foreground text-sm leading-relaxed'>
                {capability.desc}
              </p>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
