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
    <section className='relative z-10 px-4 py-10 md:px-6 md:py-14'>
      {/* Gradient banner panel */}
      <div className='relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[32px] px-6 py-16 shadow-[0_40px_90px_-45px_rgba(30,64,175,0.55)] md:px-12 md:py-20'>
        <div
          aria-hidden
          className='absolute inset-0 -z-20 bg-[linear-gradient(120deg,#1e3a8a_0%,#2563eb_38%,#4f46e5_70%,#0ea5e9_100%)]'
        />
        <div
          aria-hidden
          className='absolute inset-0 -z-10 bg-[linear-gradient(to_right,rgba(255,255,255,0.10)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.10)_1px,transparent_1px)] bg-[size:44px_44px] [mask-image:radial-gradient(ellipse_80%_70%_at_50%_30%,black_30%,transparent_100%)]'
        />
        <div
          aria-hidden
          className='absolute -top-24 -left-16 -z-10 size-80 rounded-full bg-cyan-300/25 blur-3xl'
        />
        <div
          aria-hidden
          className='absolute -right-20 -bottom-24 -z-10 size-96 rounded-full bg-fuchsia-400/20 blur-3xl'
        />

          <div className='mx-auto max-w-6xl'>
            <AnimateInView className='mb-12 text-center md:mb-16'>
              <p className='mb-3 text-xs font-medium tracking-widest uppercase text-blue-100/80'>
                {t('Enterprise Capabilities')}
              </p>
              <h2 className='text-2xl font-bold tracking-tight text-white md:text-3xl'>
                {t('Designed for enterprise-grade AI integration')}
              </h2>
              <p className='mx-auto mt-3 max-w-2xl text-sm leading-relaxed text-blue-50/85'>
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
                className='group border-white/15 bg-white/10 hover:border-white/30 hover:bg-white/15 relative flex flex-col rounded-2xl border p-6 backdrop-blur-sm transition-all duration-300 hover:shadow-[0_20px_45px_-25px_rgba(2,6,23,0.6)]'
              >
                <span className='mb-5 inline-flex size-11 items-center justify-center rounded-xl bg-white/15 text-white transition-colors group-hover:bg-white/25'>
                  {capability.icon}
                </span>
                <h3 className='mb-2 text-base font-semibold text-white'>
                  {capability.title}
                </h3>
                <p className='text-sm leading-relaxed text-blue-50/85'>
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
