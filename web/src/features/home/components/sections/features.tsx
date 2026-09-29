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
    <section
      aria-labelledby='features-heading'
      className='relative z-10 px-4 py-8 md:px-6 md:py-10'
    >
      {/* Deep gradient capability banner */}
      <div className='relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[24px] px-6 py-12 shadow-[0_40px_90px_-45px_rgba(30,64,175,0.55)] ring-1 ring-white/10 md:rounded-[32px] md:px-12 md:py-18'>
        <div
          aria-hidden
          className='absolute inset-0 -z-20 bg-[linear-gradient(135deg,#0b2a6b_0%,#1d4ed8_42%,#4f46e5_74%,#06b6d4_100%)]'
        />
        <div
          aria-hidden
          className='absolute inset-0 -z-10 bg-[linear-gradient(to_right,rgba(255,255,255,0.10)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.10)_1px,transparent_1px)] bg-[size:44px_44px] [mask-image:radial-gradient(ellipse_80%_70%_at_50%_30%,black_30%,transparent_100%)]'
        />
        <div
          aria-hidden
          className='absolute -top-24 -left-16 -z-10 size-80 rounded-full bg-cyan-300/25 blur-2xl'
        />
        <div
          aria-hidden
          className='absolute -right-20 -bottom-24 -z-10 size-96 rounded-full bg-fuchsia-400/20 blur-2xl'
        />
        {/* Top highlight arc */}
        <div
          aria-hidden
          className='absolute inset-x-0 top-0 -z-10 h-px bg-gradient-to-r from-transparent via-white/50 to-transparent'
        />

        <div className='mx-auto max-w-6xl'>
          <AnimateInView className='mb-12 text-center md:mb-14'>
            <span className='inline-flex items-center rounded-full border border-white/20 bg-white/10 px-3 py-1 text-[11px] font-medium tracking-[0.18em] text-blue-100 uppercase backdrop-blur-sm'>
              {t('Enterprise Capabilities')}
            </span>
            <h2
              id='features-heading'
              className='mt-5 text-[clamp(1.75rem,3vw,2.5rem)] leading-[1.18] font-bold tracking-tight text-white'
            >
              {t('Designed for enterprise-grade AI integration')}
            </h2>
            <p className='mx-auto mt-4 max-w-2xl text-[15px] leading-relaxed text-blue-50/85'>
              {t(
                'From model calls to usage management, teams get unified access, unified billing and unified governance.'
              )}
            </p>
          </AnimateInView>

          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4'>
            {capabilities.map((capability, i) => (
              <AnimateInView
                key={capability.title}
                delay={i * 110}
                animation='fade-up'
                className='group relative flex flex-col overflow-hidden rounded-[20px] border border-white/15 bg-white/10 p-6 backdrop-blur-sm transition-all duration-300 motion-safe:hover:-translate-y-1 hover:border-white/30 hover:bg-white/15 hover:shadow-[0_24px_50px_-28px_rgba(2,6,23,0.7)]'
              >
                {/* Hover highlight bar */}
                <span
                  aria-hidden
                  className='absolute inset-x-6 top-0 h-px bg-gradient-to-r from-transparent via-white/70 to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100'
                />
                {/* Index */}
                <span className='absolute top-5 right-5 text-[13px] font-semibold tabular-nums text-white/25 transition-colors group-hover:text-white/50'>
                  {String(i + 1).padStart(2, '0')}
                </span>

                <span className='mb-5 inline-flex size-12 items-center justify-center rounded-[14px] border border-white/15 bg-white/12 text-white transition-colors duration-300 group-hover:border-white/30 group-hover:bg-white/20'>
                  {capability.icon}
                </span>
                <h3 className='mb-2 text-[17px] font-semibold text-white'>
                  {capability.title}
                </h3>
                <p className='text-sm leading-relaxed text-blue-50/80'>
                  {capability.desc}
                </p>
              </AnimateInView>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
