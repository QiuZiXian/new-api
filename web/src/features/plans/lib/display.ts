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
import type { TFunction } from 'i18next'
import {
  Crown,
  Gift,
  LayoutGrid,
  Rocket,
  Sparkles,
  Zap,
  type LucideIcon,
} from 'lucide-react'

import { formatDuration } from '@/features/subscriptions/lib'
import type { SubscriptionPlan } from '@/features/subscriptions/types'
import { formatLocalCurrencyAmount } from '@/lib/currency'

export const PLAN_ICONS: Record<string, LucideIcon> = {
  sparkles: Sparkles,
  rocket: Rocket,
  grid: LayoutGrid,
  gift: Gift,
  crown: Crown,
  zap: Zap,
}

export const PLAN_ICON_OPTIONS = [
  'sparkles',
  'rocket',
  'grid',
  'gift',
  'crown',
  'zap',
] as const

const FALLBACK_ICON_ORDER = ['sparkles', 'rocket', 'grid', 'crown', 'gift', 'zap']

export const PLAN_ACCENTS = {
  blue: {
    tile: 'bg-gradient-to-br from-sky-400 to-blue-600',
    border: 'border-sky-400/70',
    button: 'bg-sky-600 text-white hover:bg-sky-700',
    badge: 'bg-sky-500',
    bar: 'from-sky-400 to-blue-500',
    dot: 'bg-sky-500',
  },
  violet: {
    tile: 'bg-gradient-to-br from-violet-400 to-purple-600',
    border: 'border-violet-400/70',
    button: 'bg-violet-600 text-white hover:bg-violet-700',
    badge: 'bg-violet-500',
    bar: 'from-violet-400 to-fuchsia-500',
    dot: 'bg-violet-500',
  },
  emerald: {
    tile: 'bg-gradient-to-br from-emerald-400 to-green-600',
    border: 'border-emerald-400/70',
    button: 'bg-emerald-600 text-white hover:bg-emerald-700',
    badge: 'bg-emerald-500',
    bar: 'from-emerald-400 to-green-500',
    dot: 'bg-emerald-500',
  },
  amber: {
    tile: 'bg-gradient-to-br from-amber-400 to-orange-500',
    border: 'border-amber-400/70',
    button: 'bg-amber-500 text-white hover:bg-amber-600',
    badge: 'bg-amber-500',
    bar: 'from-amber-400 to-orange-500',
    dot: 'bg-amber-500',
  },
  rose: {
    tile: 'bg-gradient-to-br from-rose-400 to-pink-600',
    border: 'border-rose-400/70',
    button: 'bg-rose-600 text-white hover:bg-rose-700',
    badge: 'bg-rose-500',
    bar: 'from-rose-400 to-pink-500',
    dot: 'bg-rose-500',
  },
  slate: {
    tile: 'bg-gradient-to-br from-slate-500 to-slate-700',
    border: 'border-slate-400/70',
    button: 'bg-slate-800 text-white hover:bg-slate-900',
    badge: 'bg-slate-600',
    bar: 'from-slate-400 to-slate-600',
    dot: 'bg-slate-500',
  },
} as const

export type PlanAccent = keyof typeof PLAN_ACCENTS

export const PLAN_ACCENT_OPTIONS = [
  'blue',
  'violet',
  'emerald',
  'amber',
  'rose',
  'slate',
] as const

const FALLBACK_ACCENT_ORDER: PlanAccent[] = [
  'blue',
  'violet',
  'emerald',
  'amber',
  'rose',
  'slate',
]

export function resolvePlanIcon(
  value: string | undefined,
  index: number
): LucideIcon {
  const key = value?.trim()
  if (key && Object.prototype.hasOwnProperty.call(PLAN_ICONS, key)) {
    return PLAN_ICONS[key]
  }
  const fallbackKey = FALLBACK_ICON_ORDER[index % FALLBACK_ICON_ORDER.length]
  return PLAN_ICONS[fallbackKey]
}

export function resolvePlanAccent(
  value: string | undefined,
  index: number
): (typeof PLAN_ACCENTS)[PlanAccent] {
  const key = value?.trim() ?? ''
  if (Object.prototype.hasOwnProperty.call(PLAN_ACCENTS, key)) {
    return PLAN_ACCENTS[key as PlanAccent]
  }
  return PLAN_ACCENTS[
    FALLBACK_ACCENT_ORDER[index % FALLBACK_ACCENT_ORDER.length]
  ]
}

const PERIOD_KEYS: Record<string, string> = {
  year: 'per year',
  month: 'per month',
  day: 'per day',
  hour: 'per hour',
}

const PERIOD_UNIT_KEYS: Record<string, string> = {
  year: 'years',
  month: 'months',
  day: 'days',
  hour: 'hours',
}

/** "/月", "/年" or "/3 months" style suffix for the plan price. */
export function formatPricePeriod(
  plan: Partial<SubscriptionPlan>,
  t: TFunction
): string {
  const unit = plan.duration_unit || 'month'
  const value = Number(plan.duration_value || 1)
  if (unit === 'custom') {
    return formatDuration(plan, t)
  }
  if (value <= 1) {
    return t(PERIOD_KEYS[unit] || PERIOD_KEYS.month)
  }
  return `${value} ${t(PERIOD_UNIT_KEYS[unit] || PERIOD_UNIT_KEYS.month)}`
}

export type SplitPrice = {
  prefix: string
  amount: string
  suffix: string
}

/** Split a formatted price so the currency symbol can be rendered smaller. */
export function splitFormattedPrice(value: number): SplitPrice {
  const formatted = formatLocalCurrencyAmount(value)
  const match = formatted.match(/^([^\d-]*)(-?[\d.,]+)(.*)$/)
  if (!match) {
    return { prefix: '', amount: formatted, suffix: '' }
  }
  return { prefix: match[1], amount: match[2], suffix: match[3] }
}
