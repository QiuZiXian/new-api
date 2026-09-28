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
    <section className='border-border/40 relative z-10 border-t px-6 py-20 md:py-28'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-12 text-center md:mb-16'>
          <p className='text-muted-foreground mb-3 text-xs font-medium tracking-widest uppercase'>
            {t('Teams')}
          </p>
          <h2 className='text-2xl font-bold tracking-tight md:text-3xl'>
            {t('Built for every AI-powered team')}
          </h2>
          <p className='text-muted-foreground mx-auto mt-3 max-w-2xl text-sm leading-relaxed'>
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
              className='border-border/50 bg-card hover:border-blue-500/30 hover:shadow-[0_16px_40px_-20px_rgba(37,99,235,0.35)] relative flex flex-col rounded-2xl border p-6 transition-all duration-300'
            >
              <span className='mb-5 inline-flex size-11 items-center justify-center rounded-xl bg-blue-500/10 text-blue-600 dark:bg-blue-400/10 dark:text-blue-400'>
                {team.icon}
              </span>
              <h3 className='mb-2 text-base font-semibold'>{team.title}</h3>
              <p className='text-muted-foreground mb-5 text-sm leading-relaxed'>
                {team.desc}
              </p>
              <div className='mt-auto flex flex-wrap gap-1.5'>
                {team.tags.map((tag) => (
                  <span
                    key={tag}
                    className='border-border/60 bg-muted/40 text-muted-foreground rounded-md border px-2 py-0.5 text-[11px]'
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
