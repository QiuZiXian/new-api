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
import { formatQuota } from '@/lib/format'

import type { PlanQuotaUnit } from '../types'

/**
 * Normalize the quota unit of a plan or subscription.
 *
 * A plan has exactly one unit: `amount` (framework quota / money-like),
 * `call` (requests) or `token`. Unknown or legacy empty values fall back to
 * `amount`, which is how plans were billed before the unit existed.
 */
export function normalizePlanQuotaUnit(unit?: string): PlanQuotaUnit {
  const value = unit?.trim().toLowerCase()
  if (value === 'call' || value === 'calls') return 'call'
  if (value === 'token' || value === 'tokens') return 'token'
  return 'amount'
}

const compactFormatter = new Intl.NumberFormat(undefined, {
  notation: 'compact',
  maximumFractionDigits: 1,
})

const countFormatter = new Intl.NumberFormat()

/**
 * Format a quota value in its unit.
 * `amount` goes through the shared quota formatter (quota-per-unit + currency).
 */
export function formatQuotaInUnit(value: number, unit?: string): string {
  const amount = Number(value) || 0
  switch (normalizePlanQuotaUnit(unit)) {
    case 'call':
      return countFormatter.format(amount)
    case 'token':
      return compactFormatter.format(amount)
    default:
      return formatQuota(amount)
  }
}

/** i18n key for the unit suffix ("calls" / "tokens"); empty for amount. */
export function quotaUnitSuffixKey(unit?: string): 'calls' | 'tokens' | '' {
  switch (normalizePlanQuotaUnit(unit)) {
    case 'call':
      return 'calls'
    case 'token':
      return 'tokens'
    default:
      return ''
  }
}

/** Percentage of a quota that has been used, clamped to 0-100. */
export function quotaUsagePercent(used: number, limit: number): number {
  const total = Number(limit) || 0
  const consumed = Number(used) || 0
  if (total <= 0) return 0
  const percent = Math.round((consumed / total) * 100)
  if (percent < 0) return 0
  return percent > 100 ? 100 : percent
}
