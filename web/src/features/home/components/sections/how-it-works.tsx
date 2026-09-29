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
import { Code2, PenTool, Building2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

export function HowItWorks() {
  const { t } = useTranslation()

  const teams = [
    {
      icon: <Code2 className='size-5' strokeWidth={1.6} />,
      title: t('Developer teams'),
      desc: t(
        'Integrate multi-model capabilities quickly and cut vendor adaptation and API maintenance costs.'
      ),
      tags: [t('API access'), t('Model switching'), t('Call logs')],
    },
    {
      icon: <PenTool className='size-5' strokeWidth={1.6} />,
      title: t('Content teams'),
      desc: t(
        'Call image, video and speech models in one place to speed up marketing assets and short-video production.'
      ),
      tags: [t('Image generation'), t('Video generation'), t('Speech generation')],
    },
    {
      icon: <Building2 className='size-5' strokeWidth={1.6} />,
      title: t('Enterprise admins'),
      desc: t(
        'Review usage, cost, key permissions and call records centrally across teams with less overhead.'
      ),
      tags: [t('Cost management'), t('Permission control'), t('Usage stats')],
    },
  ]

  return (
    <section className='relative z-10 px-4 py-10 md:px-6 md:py-14'>
      {/* Gradient banner panel */}
      <div className='relative isolate mx-auto max-w-[1400px] overflow-hidden rounded-[32px] px-6 py-16 shadow-[0_40px_90px_-45px_rgba(49,46,129,0.55)] md:px-12 md:py-20'>
        <div
          aria-hidden
          className='absolute inset-0 -z-20 bg-[linear-gradient(120deg,#172554_0%,#1d4ed8_48%,#7c3aed_100%)]'
        />
        <div
          aria-hidden
          className='absolute inset-0 -z-10 bg-[linear-gradient(to_right,rgba(255,255,255,0.10)_1px,transparent_1px),linear-gradient(to_bottom,rgba(255,255,255,0.10)_1px,transparent_1px)] bg-[size:44px_44px] [mask-image:radial-gradient(ellipse_80%_70%_at_50%_30%,black_30%,transparent_100%)]'
        />
        <div
          aria-hidden
          className='absolute -top-24 -right-16 -z-10 size-80 rounded-full bg-violet-300/25 blur-3xl'
        />
        <div
          aria-hidden
          className='absolute -bottom-24 -left-20 -z-10 size-96 rounded-full bg-sky-400/20 blur-3xl'
        />

        <div className='mx-auto max-w-6xl'>
          <AnimateInView className='mb-12 text-center md:mb-16'>
            <p className='mb-3 text-xs font-medium tracking-widest uppercase text-indigo-100/80'>
              {t('Teams')}
            </p>
            <h2 className='text-2xl font-bold tracking-tight text-white md:text-3xl'>
              {t('Built for every AI-powered team')}
            </h2>
            <p className='mx-auto mt-3 max-w-2xl text-sm leading-relaxed text-indigo-50/85'>
              {t(
                'Whether it is API integration, content production or enterprise governance, one platform covers the workflow.'
              )}
            </p>
          </AnimateInView>

          <div className='grid gap-5 md:grid-cols-3'>
            {teams.map((team, i) => (
              <AnimateInView
                key={team.title}
                delay={i * 140}
                animation='fade-up'
                className='group border-white/15 bg-white/10 hover:border-white/30 hover:bg-white/15 relative flex flex-col rounded-2xl border p-6 backdrop-blur-sm transition-all duration-300 hover:shadow-[0_20px_45px_-25px_rgba(2,6,23,0.6)]'
              >
                <span className='mb-5 inline-flex size-11 items-center justify-center rounded-xl bg-white/15 text-white transition-colors group-hover:bg-white/25'>
                  {team.icon}
                </span>
                <h3 className='mb-2 text-base font-semibold text-white'>
                  {team.title}
                </h3>
                <p className='mb-5 text-sm leading-relaxed text-indigo-50/85'>
                  {team.desc}
                </p>
                <div className='mt-auto flex flex-wrap gap-1.5'>
                  {team.tags.map((tag) => (
                    <span
                      key={tag}
                      className='rounded-md border border-white/20 bg-white/10 px-2 py-0.5 text-[11px] text-indigo-50/90'
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </AnimateInView>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
