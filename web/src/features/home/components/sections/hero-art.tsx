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
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'

interface HeroArtProps {
  variant: 'rings' | 'disc'
}

/** Soft glow disc reused behind every scene. */
function SceneGlow() {
  return (
    <>
      <defs>
        <radialGradient id='hero-glow' cx='50%' cy='45%' r='55%'>
          <stop offset='0%' stopColor='#93c5fd' stopOpacity='0.55' />
          <stop offset='60%' stopColor='#bfdbfe' stopOpacity='0.22' />
          <stop offset='100%' stopColor='#bfdbfe' stopOpacity='0' />
        </radialGradient>
        <linearGradient id='hero-tower' x1='0' y1='0' x2='0' y2='1'>
          <stop offset='0%' stopColor='#60a5fa' />
          <stop offset='55%' stopColor='#3b82f6' />
          <stop offset='100%' stopColor='#1d4ed8' />
        </linearGradient>
        <linearGradient id='hero-chip' x1='0' y1='0' x2='1' y2='1'>
          <stop offset='0%' stopColor='#e0f2fe' />
          <stop offset='100%' stopColor='#93c5fd' />
        </linearGradient>
        <linearGradient id='hero-chip-deep' x1='0' y1='0' x2='1' y2='1'>
          <stop offset='0%' stopColor='#93c5fd' />
          <stop offset='100%' stopColor='#3b82f6' />
        </linearGradient>
        <linearGradient id='hero-base' x1='0' y1='0' x2='0' y2='1'>
          <stop offset='0%' stopColor='#dbeafe' />
          <stop offset='100%' stopColor='#eff6ff' />
        </linearGradient>
        <linearGradient id='hero-bar' x1='0' y1='0' x2='0' y2='1'>
          <stop offset='0%' stopColor='#3b82f6' />
          <stop offset='100%' stopColor='#93c5fd' />
        </linearGradient>
        <linearGradient id='hero-screen' x1='0' y1='0' x2='1' y2='1'>
          <stop offset='0%' stopColor='#f0f9ff' />
          <stop offset='100%' stopColor='#dbeafe' />
        </linearGradient>
      </defs>
      <ellipse cx='280' cy='215' rx='250' ry='185' fill='url(#hero-glow)' />
    </>
  )
}

/** Scene A — glowing core tower with orbiting capability chips. */
function RingsScene() {
  return (
    <g>
      <ellipse cx='280' cy='330' rx='210' ry='42' fill='url(#hero-base)' />
      <ellipse
        cx='280'
        cy='322'
        rx='196'
        ry='34'
        fill='none'
        stroke='#93c5fd'
        strokeOpacity='0.6'
      />
      <ellipse
        cx='280'
        cy='300'
        rx='176'
        ry='64'
        fill='none'
        stroke='#60a5fa'
        strokeOpacity='0.5'
        strokeWidth='1.5'
        strokeDasharray='3 8'
      />
      <ellipse
        cx='280'
        cy='286'
        rx='128'
        ry='46'
        fill='none'
        stroke='#3b82f6'
        strokeOpacity='0.55'
        strokeWidth='2'
      />
      {/* Beam + tower */}
      <polygon points='280,60 252,196 308,196' fill='#bfdbfe' opacity='0.55' />
      <rect
        x='252'
        y='176'
        width='56'
        height='128'
        rx='16'
        fill='url(#hero-tower)'
      />
      <rect
        x='262'
        y='192'
        width='36'
        height='12'
        rx='6'
        fill='#eff6ff'
        opacity='0.85'
      />
      <rect
        x='262'
        y='212'
        width='36'
        height='8'
        rx='4'
        fill='#dbeafe'
        opacity='0.7'
      />
      <ellipse cx='280' cy='308' rx='44' ry='12' fill='#1d4ed8' opacity='0.9' />
      {/* Orbiting chips */}
      <g transform='rotate(-8 130 250)'>
        <rect
          x='96'
          y='226'
          width='64'
          height='48'
          rx='12'
          fill='url(#hero-chip)'
          stroke='#bfdbfe'
        />
        <rect x='110' y='240' width='36' height='6' rx='3' fill='#3b82f6' opacity='0.75' />
        <rect x='110' y='252' width='24' height='6' rx='3' fill='#60a5fa' opacity='0.6' />
      </g>
      <g transform='rotate(6 428 232)'>
        <rect
          x='400'
          y='204'
          width='60'
          height='60'
          rx='14'
          fill='url(#hero-chip-deep)'
        />
        <circle cx='430' cy='234' r='14' fill='#eff6ff' opacity='0.9' />
        <path
          d='M424 234h12M430 228v12'
          stroke='#3b82f6'
          strokeWidth='2.5'
          strokeLinecap='round'
        />
      </g>
      <g transform='rotate(-5 452 320)'>
        <rect
          x='424'
          y='296'
          width='52'
          height='40'
          rx='10'
          fill='url(#hero-chip)'
          stroke='#93c5fd'
        />
        <rect x='436' y='308' width='28' height='5' rx='2.5' fill='#2563eb' opacity='0.7' />
        <rect x='436' y='318' width='18' height='5' rx='2.5' fill='#60a5fa' opacity='0.6' />
      </g>
      <rect
        x='84'
        y='312'
        width='48'
        height='36'
        rx='10'
        fill='url(#hero-chip-deep)'
        opacity='0.92'
      />
      <circle cx='108' cy='330' r='8' fill='#eff6ff' opacity='0.9' />
      {/* Floating dots */}
      <circle cx='200' cy='120' r='4' fill='#60a5fa' opacity='0.8' />
      <circle cx='352' cy='96' r='5' fill='#93c5fd' opacity='0.9' />
      <circle cx='404' cy='140' r='3' fill='#3b82f6' opacity='0.7' />
      <circle cx='160' cy='168' r='3' fill='#22d3ee' opacity='0.8' />
    </g>
  )
}

/** Scene B — isometric disc with generated-media bars and floating screens. */
function DiscScene() {
  return (
    <g>
      <ellipse cx='280' cy='320' rx='215' ry='46' fill='url(#hero-base)' />
      <ellipse
        cx='280'
        cy='312'
        rx='192'
        ry='38'
        fill='none'
        stroke='#60a5fa'
        strokeOpacity='0.55'
        strokeWidth='1.5'
      />
      <ellipse
        cx='280'
        cy='304'
        rx='150'
        ry='30'
        fill='none'
        stroke='#fb923c'
        strokeOpacity='0.75'
        strokeWidth='2.5'
        strokeDasharray='150 320'
        strokeLinecap='round'
      />
      {/* City bars */}
      <rect x='196' y='212' width='26' height='96' rx='6' fill='url(#hero-bar)' />
      <rect x='232' y='176' width='30' height='132' rx='6' fill='#2563eb' />
      <rect x='272' y='196' width='26' height='112' rx='6' fill='url(#hero-bar)' />
      <rect x='308' y='230' width='24' height='78' rx='6' fill='#60a5fa' />
      <rect x='340' y='248' width='20' height='60' rx='6' fill='#93c5fd' />
      <rect x='236' y='180' width='22' height='8' rx='4' fill='#93c5fd' opacity='0.9' />
      <rect x='276' y='200' width='18' height='7' rx='3.5' fill='#dbeafe' opacity='0.9' />
      {/* Floating screens */}
      <g transform='rotate(-6 150 168)'>
        <rect
          x='112'
          y='136'
          width='78'
          height='58'
          rx='10'
          fill='url(#hero-screen)'
          stroke='#93c5fd'
        />
        <rect x='124' y='150' width='40' height='6' rx='3' fill='#3b82f6' opacity='0.8' />
        <rect x='124' y='164' width='54' height='5' rx='2.5' fill='#93c5fd' />
        <rect x='124' y='176' width='30' height='5' rx='2.5' fill='#bfdbfe' />
      </g>
      <g transform='rotate(5 416 150)'>
        <rect
          x='382'
          y='118'
          width='70'
          height='54'
          rx='10'
          fill='url(#hero-chip-deep)'
        />
        <polygon points='411,133 411,157 431,145' fill='#eff6ff' opacity='0.95' />
      </g>
      <g transform='rotate(-4 442 268)'>
        <rect
          x='408'
          y='240'
          width='66'
          height='52'
          rx='10'
          fill='url(#hero-screen)'
          stroke='#93c5fd'
        />
        <circle cx='441' cy='266' r='13' fill='none' stroke='#3b82f6' strokeWidth='3' />
        <circle cx='441' cy='266' r='5' fill='#fb923c' opacity='0.85' />
      </g>
      {/* Sparkles */}
      <circle cx='188' cy='104' r='4' fill='#3b82f6' opacity='0.75' />
      <circle cx='332' cy='84' r='5' fill='#93c5fd' opacity='0.9' />
      <circle cx='366' cy='120' r='3' fill='#22d3ee' opacity='0.8' />
      <circle cx='132' cy='268' r='3.5' fill='#60a5fa' opacity='0.7' />
    </g>
  )
}

export function HeroArt(props: HeroArtProps) {
  const reduced = useReducedMotion()
  const scenes = { rings: <RingsScene />, disc: <DiscScene /> } as const

  return (
    <div
      aria-hidden
      className='pointer-events-none mx-auto w-full max-w-[560px] dark:opacity-90'
    >
      <svg
        viewBox='0 0 560 420'
        fill='none'
        xmlns='http://www.w3.org/2000/svg'
        className='h-auto w-full'
        role='presentation'
      >
        <SceneGlow />
        <AnimatePresence mode='wait' initial={false}>
          <motion.g
            key={props.variant}
            initial={reduced ? { opacity: 1 } : { opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            exit={reduced ? { opacity: 0 } : { opacity: 0, y: -14 }}
            transition={{ duration: reduced ? 0 : 0.45, ease: 'easeOut' }}
          >
            {scenes[props.variant]}
          </motion.g>
        </AnimatePresence>
      </svg>
    </div>
  )
}
