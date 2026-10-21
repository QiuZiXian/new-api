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
import { z } from 'zod'

import { parseQuotaFromDollars, quotaUnitsToDollars } from '@/lib/format'

import type { SubscriptionPlan, PlanPayload } from '../types'

export function getPlanFormSchema(t: TFunction) {
  return z.object({
    title: z.string().min(1, t('Please enter plan title')),
    subtitle: z.string().optional(),
    price_amount: z.coerce.number().min(0, t('Please enter amount')),
    duration_unit: z.enum(['year', 'month', 'day', 'hour', 'custom']),
    duration_value: z.coerce.number().min(1),
    custom_seconds: z.coerce.number().min(0).optional(),
    quota_reset_period: z.enum([
      'never',
      'daily',
      'weekly',
      'monthly',
      'custom',
    ]),
    quota_reset_custom_seconds: z.coerce.number().min(0).optional(),
    enabled: z.boolean(),
    sort_order: z.coerce.number(),
    allow_balance_pay: z.boolean(),
    allow_wallet_overflow: z.boolean(),
    max_purchase_per_user: z.coerce.number().min(0),
    // A plan has ONE quota unit: framework quota (amount), calls or tokens.
    quota_unit: z.enum(['amount', 'call', 'token']),
    total_amount: z.coerce.number().min(0),
    upgrade_group: z.string().optional(),
    downgrade_group: z.string().optional(),
    stripe_price_id: z.string().optional(),
    creem_product_id: z.string().optional(),
    waffo_pancake_product_id: z.string().optional(),
    // Plans page presentation
    billing_note: z.string().optional(),
    badge: z.string().optional(),
    icon: z.string().optional(),
    accent: z.string().optional(),
    highlight: z.boolean(),
    quota_items: z
      .array(
        z.object({
          model: z.string(),
          count: z.coerce.number().min(0),
        })
      )
      .optional(),
    features_text: z.string().optional(),
  }).superRefine((values, ctx) => {
    // The per-model breakdown must fit inside the plan total; the remainder (if
    // any) stays available as a shared pool for models that are not listed.
    const allocated = (values.quota_items ?? []).reduce(
      (sum, item) => sum + Math.max(0, Number(item.count) || 0),
      0
    )
    const total = Number(values.total_amount) || 0
    if (allocated > 0 && total > 0 && allocated > total) {
      ctx.addIssue({
        code: 'custom',
        path: ['quota_items'],
        message: t('Allocated per-model quota exceeds the plan total'),
      })
    }
  })
}

export type PlanFormValues = z.infer<ReturnType<typeof getPlanFormSchema>>

export const PLAN_FORM_DEFAULTS: PlanFormValues = {
  title: '',
  subtitle: '',
  price_amount: 0,
  duration_unit: 'month',
  duration_value: 1,
  custom_seconds: 0,
  quota_reset_period: 'never',
  quota_reset_custom_seconds: 0,
  enabled: true,
  sort_order: 0,
  allow_balance_pay: true,
  allow_wallet_overflow: true,
  max_purchase_per_user: 0,
  quota_unit: 'amount',
  total_amount: 0,
  upgrade_group: '',
  downgrade_group: '',
  stripe_price_id: '',
  creem_product_id: '',
  waffo_pancake_product_id: '',
  billing_note: '',
  badge: '',
  icon: '',
  accent: '',
  highlight: false,
  quota_items: [],
  features_text: '',
}

export function planToFormValues(plan: SubscriptionPlan): PlanFormValues {
  const quotaUnit = normalizeFormQuotaUnit(plan.quota_unit)
  return {
    title: plan.title || '',
    subtitle: plan.subtitle || '',
    price_amount: Number(plan.price_amount || 0),
    duration_unit: plan.duration_unit || 'month',
    duration_value: Number(plan.duration_value || 1),
    custom_seconds: Number(plan.custom_seconds || 0),
    quota_reset_period: plan.quota_reset_period || 'never',
    quota_reset_custom_seconds: Number(plan.quota_reset_custom_seconds || 0),
    enabled: plan.enabled !== false,
    sort_order: Number(plan.sort_order || 0),
    allow_balance_pay: plan.allow_balance_pay !== false,
    allow_wallet_overflow: plan.allow_wallet_overflow !== false,
    max_purchase_per_user: Number(plan.max_purchase_per_user || 0),
    quota_unit: quotaUnit,
    total_amount: quotaToDisplay(
      Number(plan.total_amount || 0),
      quotaUnit
    ),
    upgrade_group: plan.upgrade_group || '',
    downgrade_group: plan.downgrade_group || '',
    stripe_price_id: plan.stripe_price_id || '',
    creem_product_id: plan.creem_product_id || '',
    waffo_pancake_product_id: plan.waffo_pancake_product_id || '',
    billing_note: plan.billing_note || '',
    badge: plan.badge || '',
    icon: plan.icon || '',
    accent: plan.accent || '',
    highlight: plan.highlight === true,
    quota_items: (plan.quota_items ?? []).map((item) => ({
      model: item.model,
      count: quotaToDisplay(Number(item.count || 0), quotaUnit),
    })),
    features_text: (plan.features ?? []).join('\n'),
  }
}

export function formValuesToPlanPayload(values: PlanFormValues): PlanPayload {
  const { features_text, quota_items, ...rest } = values
  const quotaUnit = normalizeFormQuotaUnit(values.quota_unit)
  return {
    plan: {
      ...rest,
      price_amount: Number(values.price_amount || 0),
      currency: 'USD',
      duration_value: Number(values.duration_value || 0),
      custom_seconds: Number(values.custom_seconds || 0),
      quota_reset_period: values.quota_reset_period || 'never',
      quota_reset_custom_seconds:
        values.quota_reset_period === 'custom'
          ? Number(values.quota_reset_custom_seconds || 0)
          : 0,
      sort_order: Number(values.sort_order || 0),
      max_purchase_per_user: Number(values.max_purchase_per_user || 0),
      quota_unit: quotaUnit,
      total_amount: quotaFromDisplay(
        Number(values.total_amount || 0),
        quotaUnit
      ),
      upgrade_group: values.upgrade_group || '',
      downgrade_group: values.downgrade_group || '',
      billing_note: values.billing_note?.trim() || '',
      badge: values.badge?.trim() || '',
      icon: values.icon || '',
      accent: values.accent || '',
      highlight: Boolean(values.highlight),
      quota_items: (quota_items ?? [])
        .map((item) => ({
          model: item.model.trim(),
          count: quotaFromDisplay(Number(item.count || 0), quotaUnit),
          unit: quotaUnit,
        }))
        .filter((item) => item.model !== '' && item.count > 0),
      features: (features_text ?? '')
        .split('\n')
        .map((line) => line.trim())
        .filter((line) => line !== ''),
    },
  }
}

function normalizeFormQuotaUnit(unit?: string): 'amount' | 'call' | 'token' {
  if (unit === 'call' || unit === 'token') return unit
  return 'amount'
}

/**
 * Convert a stored quota value to the number shown in the form.
 * `amount` is stored in framework quota units but edited as currency/tokens,
 * while call/token counts are plain integers.
 */
function quotaToDisplay(value: number, unit: string): number {
  return unit === 'amount' ? quotaUnitsToDollars(value) : value
}

/** Reverse of quotaToDisplay. */
function quotaFromDisplay(value: number, unit: string): number {
  return unit === 'amount'
    ? parseQuotaFromDollars(value)
    : Math.max(0, Math.round(value))
}
