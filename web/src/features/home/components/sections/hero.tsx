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
import {
  ArrowRight,
  BrainCircuit,
  Building2,
  Clapperboard,
  Image as ImageIcon,
  Layers,
  Video,
  Zap,
} from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

const SLIDE_INTERVAL = 7000
const SLIDE_IDS = ['glm-launch', 'visual-aggregation'] as const
const SLIDE_COUNT = SLIDE_IDS.length
const HERO_BACKGROUNDS = [
  '/home/hero-slide-1.jpg',
  '/home/hero-slide-2.jpg',
] as const

function FeatureCard(props: {
  icon: React.ReactNode
  title: string
  desc: string
}) {
  return (
    <div className='rounded-xl border border-blue-100/90 bg-white/80 p-3.5 shadow-[0_2px_10px_-4px_rgba(37,99,235,0.12)] backdrop-blur-xs dark:border-blue-400/10 dark:bg-white/5'>
      <div className='flex items-center gap-2'>
        <span className='flex size-7 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:bg-blue-400/10 dark:text-blue-400'>
          {props.icon}
        </span>
        <span className='text-slate-900 text-[13px] font-semibold dark:text-slate-100'>
          {props.title}
        </span>
      </div>
      <p className='text-slate-500 mt-2 text-xs leading-relaxed dark:text-slate-400'>
        {props.desc}
      </p>
    </div>
  )
}

function BulletItem(props: { children: React.ReactNode }) {
  return (
    <li className='text-slate-500 flex items-start gap-1.5 text-[11px] leading-relaxed dark:text-slate-400'>
      <span className='bg-blue-500 mt-[5px] size-1 shrink-0 rounded-full' />
      <span className='min-w-0'>{props.children}</span>
    </li>
  )
}

function SlideCTA(props: {
  isAuthenticated?: boolean
  label: string
  hint?: string
  to: string
}) {
  const { t } = useTranslation()
  const target = props.isAuthenticated ? '/dashboard' : props.to
  return (
    <div className='mt-6 flex flex-wrap items-center gap-3'>
      <Button
        className='group h-11 rounded-lg bg-gradient-to-r from-blue-600 to-blue-500 px-6 text-sm font-medium text-white shadow-[0_8px_20px_-6px_rgba(37,99,235,0.5)] hover:from-blue-700 hover:to-blue-600'
        render={<Link to={target} />}
      >
        {props.isAuthenticated ? t('Go to Dashboard') : props.label}
        <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
      </Button>
      {props.hint && (
        <span className='text-slate-500 text-xs dark:text-slate-400'>
          {props.hint}
        </span>
      )}
    </div>
  )
}

/** Slide 1 — new flagship model launch. */
function SlideGLM(props: { isAuthenticated?: boolean }) {
  const { t } = useTranslation()
  const features = [
    {
      icon: <BrainCircuit className='size-4' />,
      title: t('Smarter general intelligence'),
      desc: t('Logic, reasoning, long-form writing and coding all improved'),
    },
    {
      icon: <Layers className='size-4' />,
      title: t('Stronger complex task handling'),
      desc: t(
        'Stable performance on long documents and multi-turn professional scenarios'
      ),
    },
    {
      icon: <Zap className='size-4' />,
      title: t('Faster responses'),
      desc: t('Efficient invocation with less waiting for smoother workflows'),
    },
    {
      icon: <Building2 className='size-4' />,
      title: t('Enterprise-grade stability'),
      desc: t(
        'Built for business rollout, knowledge Q&A and assistant scenarios'
      ),
    },
  ]

  return (
    <div>
      <h1 className='text-slate-900 text-[clamp(1.9rem,3.4vw,2.75rem)] leading-[1.2] font-bold tracking-tight dark:text-slate-50'>
        <span className='text-blue-600 dark:text-blue-400'>GLM-5.3</span>
        {t(' is now live')}
      </h1>
      <p className='text-slate-600 mt-3 text-sm leading-relaxed md:text-base dark:text-slate-400'>
        {t(
          'A smarter, faster, and more stable new-generation large model platform'
        )}
      </p>

      <div className='mt-6 grid grid-cols-1 gap-3 sm:grid-cols-2'>
        {features.map((feature) => (
          <FeatureCard key={feature.title} {...feature} />
        ))}
      </div>

      <SlideCTA
        isAuthenticated={props.isAuthenticated}
        label={t('Try GLM-5.3 now')}
        hint={t('Sign in to unlock the new-generation model capabilities')}
        to='/pricing'
      />
    </div>
  )
}

/** Slide 2 — aggregated visual generation models. */
function SlideAggregate(props: { isAuthenticated?: boolean }) {
  const { t } = useTranslation()
  const cards = [
    {
      icon: <Clapperboard className='size-4' />,
      name: 'Seedance2.5',
      bullets: [
        t('Cinematic quality with strong long-take storytelling'),
        t('Ideal for premium ads, short films and series-style footage'),
        t('Consistent and stable frames'),
      ],
    },
    {
      icon: <Video className='size-4' />,
      name: 'MiniMax-H3',
      bullets: [
        t('High prompt adherence with synced audio & video'),
        t('Commercial marketing videos and CG animation'),
        t('Fast high-quality ad output with great value'),
      ],
    },
    {
      icon: <ImageIcon className='size-4' />,
      name: 'Wan3.0',
      bullets: [
        t('Standout image-to-video with locked first-frame composition'),
        t('Turn static posters into dynamic video'),
        t('Batch-produce marketing assets'),
      ],
    },
  ]

  return (
    <div>
      <h1 className='text-slate-900 text-[clamp(1.9rem,3.4vw,2.75rem)] leading-[1.2] font-bold tracking-tight dark:text-slate-50'>
        {t('Aggregate')}{' '}
        <span className='text-blue-600 dark:text-blue-400'>
          Seedance2.5、MiniMax-H3、Wan3.0
        </span>
      </h1>
      <p className='text-slate-600 mt-3 text-sm leading-relaxed md:text-base dark:text-slate-400'>
        {t(
          'Top AI visual generation models in one place — image creation, text-to-video and image-to-video without switching tools.'
        )}
      </p>

      <div className='mt-6 grid grid-cols-1 gap-3 sm:grid-cols-3'>
        {cards.map((card) => (
          <div
            key={card.name}
            className='rounded-xl border border-blue-100/90 bg-white/80 p-3.5 shadow-[0_2px_10px_-4px_rgba(37,99,235,0.12)] backdrop-blur-xs dark:border-blue-400/10 dark:bg-white/5'
          >
            <div className='flex items-center gap-2'>
              <span className='flex size-7 shrink-0 items-center justify-center rounded-lg bg-blue-500/10 text-blue-600 dark:bg-blue-400/10 dark:text-blue-400'>
                {card.icon}
              </span>
              <span className='text-slate-900 text-[13px] font-semibold dark:text-slate-100'>
                {card.name}
              </span>
            </div>
            <ul className='mt-2.5 space-y-1.5'>
              {card.bullets.map((bullet) => (
                <BulletItem key={bullet}>{bullet}</BulletItem>
              ))}
            </ul>
          </div>
        ))}
      </div>

      <div className='border-blue-100/90 mt-5 flex flex-col gap-3 rounded-xl border bg-white/75 px-4 py-3.5 backdrop-blur-xs sm:flex-row sm:items-center dark:border-blue-400/10 dark:bg-white/5'>
        <p className='text-slate-600 line-clamp-2 flex-1 text-xs leading-relaxed dark:text-slate-400'>
          {t(
            'Enter a prompt to quickly generate HD images and videos — marketing posters, short-video covers, ad bumpers and batch content. For creators, operators and enterprises, it lowers the production barrier from idea to final cut.'
          )}
        </p>
        <Button
          className='group h-10 shrink-0 rounded-lg bg-gradient-to-r from-blue-600 to-blue-500 px-5 text-sm font-medium text-white shadow-[0_8px_20px_-6px_rgba(37,99,235,0.5)] hover:from-blue-700 hover:to-blue-600'
          render={
            <Link to={props.isAuthenticated ? '/dashboard' : '/sign-in'} />
          }
        >
          {props.isAuthenticated ? t('Go to Dashboard') : t('Sign in to experience')}
          <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
        </Button>
      </div>
    </div>
  )
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const reduced = useReducedMotion()
  const [active, setActive] = useState(0)
  const pausedRef = useRef(false)

  useEffect(() => {
    if (reduced) return
    const id = window.setInterval(() => {
      if (!pausedRef.current) {
        setActive((current) => (current + 1) % SLIDE_COUNT)
      }
    }, SLIDE_INTERVAL)
    return () => window.clearInterval(id)
  }, [reduced])

  const slideTransition = {
    duration: reduced ? 0 : 0.45,
    ease: 'easeOut' as const,
  }

  return (
    <section
      className='relative isolate overflow-hidden'
      onMouseEnter={() => (pausedRef.current = true)}
      onMouseLeave={() => (pausedRef.current = false)}
    >
      {/* Rotating background artwork from design assets */}
      <div aria-hidden className='absolute inset-0 -z-20'>
        <AnimatePresence mode='wait' initial={false}>
          <motion.img
            key={active}
            src={HERO_BACKGROUNDS[active]}
            alt=''
            initial={reduced ? { opacity: 1 } : { opacity: 0, scale: 1.03 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={reduced ? { opacity: 0 } : { opacity: 0 }}
            transition={{ duration: reduced ? 0 : 0.6, ease: 'easeOut' }}
            className='absolute inset-0 h-full w-full object-cover object-[72%_center] sm:object-[78%_center]'
          />
        </AnimatePresence>
        {/* Readability overlay: darken behind copy, fade out over artwork */}
        <div className='absolute inset-0 bg-[linear-gradient(90deg,rgba(238,244,255,0.92)_0%,rgba(238,244,255,0.55)_38%,rgba(238,244,255,0)_65%)] dark:bg-[linear-gradient(90deg,rgba(8,15,35,0.94)_0%,rgba(8,15,35,0.6)_38%,rgba(8,15,35,0.12)_65%)]' />
        <div className='absolute inset-x-0 bottom-0 h-24 bg-[linear-gradient(180deg,transparent,var(--background))]' />
      </div>

      <div className='mx-auto max-w-7xl px-6 pt-24 pb-16 sm:pt-32 lg:pb-24'>
        {/* Left: rotating banner copy */}
        <div className='landing-animate-fade-up min-w-0 max-w-2xl opacity-0'>
          <AnimatePresence mode='wait' initial={false}>
            <motion.div
              key={active}
              initial={reduced ? { opacity: 1 } : { opacity: 0, y: 18 }}
              animate={{ opacity: 1, y: 0 }}
              exit={reduced ? { opacity: 0 } : { opacity: 0, y: -18 }}
              transition={slideTransition}
            >
              {active === 0 ? (
                <SlideGLM isAuthenticated={props.isAuthenticated} />
              ) : (
                <SlideAggregate isAuthenticated={props.isAuthenticated} />
              )}
            </motion.div>
          </AnimatePresence>

          {/* Slide dots */}
          <div className='mt-8 flex items-center gap-2'>
            {SLIDE_IDS.map((slideId, index) => (
              <button
                key={slideId}
                type='button'
                aria-label={t('Slide {{index}}', { index: index + 1 })}
                aria-current={active === index}
                onClick={() => setActive(index)}
                className={cn(
                  'h-2 rounded-full transition-all duration-300',
                  active === index
                    ? 'bg-blue-600 w-6 dark:bg-blue-400'
                    : 'bg-blue-300/70 hover:bg-blue-400 w-2 dark:bg-slate-600 dark:hover:bg-slate-500'
                )}
              />
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
