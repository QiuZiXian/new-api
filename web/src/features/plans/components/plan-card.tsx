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
import { Check } from 'lucide-react'
import { createElement } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  formatQuotaInUnit,
  normalizePlanQuotaUnit,
  quotaUnitSuffixKey,
  quotaUsagePercent,
} from '@/features/subscriptions/lib'
import type {
  SubscriptionPlan,
  SubscriptionQuotaUsage,
} from '@/features/subscriptions/types'
import { cn } from '@/lib/utils'

import {
  formatPricePeriod,
  resolvePlanAccent,
  resolvePlanIcon,
  splitFormattedPrice,
} from '../lib/display'

/** Live consumption of the plan the user already owns, if any. */
export type PlanUsage = {
  totalUsed: number
  totalLimit: number
  totalRemaining: number
  modelUsage: SubscriptionQuotaUsage[]
}

type PlanCardProps = {
  plan: SubscriptionPlan
  index: number
  purchaseCount: number
  usage?: PlanUsage | null
  onSubscribe: () => void
}

export function PlanCard({
  plan,
  index,
  purchaseCount,
  usage,
  onSubscribe,
}: PlanCardProps) {
  const { t } = useTranslation()
  const iconNode = createElement(resolvePlanIcon(plan.icon, index), {
    className: 'size-5',
  })
  const accent = resolvePlanAccent(plan.accent, index)
  const featured = Boolean(plan.highlight)
  const price = splitFormattedPrice(Number(plan.price_amount || 0))
  const period = formatPricePeriod(plan, t)
  const quotaUnit = normalizePlanQuotaUnit(plan.quota_unit)
  const unitLabel = quotaUnitSuffixKey(quotaUnit)
  const quotaItems = plan.quota_items ?? []
  const features = plan.features ?? []
  const totalQuota = Number(plan.total_amount || 0)
  const sharedLimit = Math.max(0, totalQuota - allocatedTotal(quotaItems))
  const purchaseLimit = Number(plan.max_purchase_per_user || 0)
  const limitReached = purchaseLimit > 0 && purchaseCount >= purchaseLimit

  const usageByModel = new Map<string, SubscriptionQuotaUsage>()
  for (const item of usage?.modelUsage ?? []) {
    usageByModel.set(item.model.trim().toLowerCase(), item)
  }

  return (
    <article
      className={cn(
        'bg-card relative flex h-full flex-col rounded-2xl p-5 transition-shadow sm:p-6',
        featured
          ? cn('border-2 shadow-lg', accent.border)
          : 'border-border/70 hover:border-border border shadow-[0_1px_2px_rgba(15,23,42,0.04)]'
      )}
    >
      {featured && plan.badge ? (
        <span
          className={cn(
            'absolute -top-3 left-1/2 -translate-x-1/2 rounded-full px-3 py-1 text-[11px] font-medium whitespace-nowrap text-white shadow-sm',
            accent.badge
          )}
        >
          {plan.badge}
        </span>
      ) : null}

      <div className='flex items-start gap-3'>
        <div
          className={cn(
            'flex size-11 shrink-0 items-center justify-center rounded-xl text-white shadow-sm',
            accent.tile
          )}
        >
          {iconNode}
        </div>
        <div className='min-w-0 pt-0.5'>
          <h3 className='truncate text-lg leading-tight font-bold'>
            {plan.title}
          </h3>
          {plan.billing_note ? (
            <p className='text-muted-foreground mt-1 text-[11px]'>
              {plan.billing_note}
            </p>
          ) : null}
        </div>
      </div>

      {plan.subtitle ? (
        <p className='text-muted-foreground mt-3 text-xs'>{plan.subtitle}</p>
      ) : null}

      <div className='mt-4 flex items-baseline'>
        {price.prefix ? (
          <span className='text-foreground/80 text-lg font-semibold'>
            {price.prefix}
          </span>
        ) : null}
        <span className='ml-0.5 text-[2.25rem] leading-none font-bold tracking-tight'>
          {price.amount}
        </span>
        <span className='text-muted-foreground ml-1 text-xs'>/{period}</span>
      </div>

      <div className='bg-muted/40 mt-4 rounded-xl p-3'>
        {quotaItems.length > 0 ? (
          <div className='space-y-3'>
            {quotaItems.map((item) => {
              const used = usageByModel.get(item.model.trim().toLowerCase())?.used
              const percent =
                used === undefined ? 0 : quotaUsagePercent(used, item.count)
              return (
                <div key={item.model}>
                  <div className='flex items-center justify-between gap-2'>
                    <span className='flex min-w-0 items-center gap-1.5'>
                      <span
                        className={cn(
                          'size-1.5 shrink-0 rotate-45 rounded-[1px]',
                          accent.dot
                        )}
                      />
                      <span className='truncate font-mono text-[11px] font-medium'>
                        {item.model}
                      </span>
                    </span>
                    <span className='shrink-0 text-[11px] font-semibold tabular-nums'>
                      {used === undefined ? null : (
                        <span className='text-muted-foreground font-normal'>
                          {formatQuotaInUnit(used, quotaUnit)} /{' '}
                        </span>
                      )}
                      {formatQuotaInUnit(item.count, quotaUnit)}
                      {unitLabel ? (
                        <span className='text-muted-foreground ml-1 font-normal'>
                          {t(unitLabel)}
                        </span>
                      ) : null}
                    </span>
                  </div>
                  <div className='bg-border/60 mt-1.5 h-1 w-full overflow-hidden rounded-full'>
                    <div
                      className={cn(
                        'h-full rounded-full bg-gradient-to-r transition-[width] duration-500',
                        accent.bar
                      )}
                      style={{ width: `${used === undefined ? 100 : percent}%` }}
                    />
                  </div>
                </div>
              )
            })}
            {sharedLimit > 0 ? (
              <div className='text-muted-foreground flex items-center justify-between text-[11px]'>
                <span>{t('Shared quota')}</span>
                <span className='font-medium tabular-nums'>
                  {formatQuotaInUnit(sharedLimit, quotaUnit)}{' '}
                  {unitLabel ? t(unitLabel) : null}
                </span>
              </div>
            ) : null}
            <div className='border-border/70 border-t pt-2.5 text-[11px]'>
              <div className='flex items-center justify-between'>
                <span className='text-muted-foreground'>
                  {usage ? t('Remaining') : t(totalLabelKey(quotaUnit))}
                </span>
                <span className='font-semibold tabular-nums'>
                  {renderTotalQuota(usage, totalQuota, quotaUnit, t)}
                  {unitLabel && !(usage && usage.totalLimit <= 0)
                    ? ` ${t(unitLabel)}`
                    : ''}
                </span>
              </div>
              {usage ? (
                <div className='bg-border/60 mt-1.5 h-1.5 w-full overflow-hidden rounded-full'>
                  <div
                    className={cn(
                      'h-full rounded-full bg-gradient-to-r',
                      accent.bar
                    )}
                    style={{
                      width: `${quotaUsagePercent(usage.totalUsed, usage.totalLimit)}%`,
                    }}
                  />
                </div>
              ) : null}
            </div>
          </div>
        ) : (
          <div className='flex items-center justify-between text-[11px]'>
            <span className='text-muted-foreground'>
              {usage ? t('Remaining') : t('Plan Quota')}
            </span>
            <span className='font-semibold'>
              {renderTotalQuota(usage, totalQuota, quotaUnit, t)}
            </span>
          </div>
        )}
      </div>

      <Button
        type='button'
        className={cn(
          'mt-4 h-10 w-full rounded-lg text-sm font-semibold',
          featured
            ? accent.button
            : 'bg-slate-900 text-white hover:bg-slate-800 dark:bg-white dark:text-slate-900 dark:hover:bg-slate-200'
        )}
        disabled={limitReached}
        onClick={onSubscribe}
      >
        {limitReached ? t('Limit Reached') : t('Subscribe Now')}
      </Button>
      {limitReached ? (
        <p className='text-muted-foreground mt-1.5 text-center text-[11px]'>
          {t('Purchase limit reached')} ({purchaseCount}/{purchaseLimit})
        </p>
      ) : null}

      {features.length > 0 ? (
        <div className='mt-5 flex-1'>
          <p className='text-muted-foreground text-xs font-medium'>
            {t("What's included")}
          </p>
          <ul className='mt-2.5 space-y-2'>
            {features.map((feature) => (
              <li
                key={feature}
                className='flex items-start gap-2 text-xs leading-relaxed'
              >
                <span className='mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full bg-emerald-100 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-400'>
                  <Check className='size-2.5' strokeWidth={3} />
                </span>
                <span>{feature}</span>
              </li>
            ))}
          </ul>
        </div>
      ) : (
        <div className='flex-1' />
      )}
    </article>
  )
}

function allocatedTotal(
  items: Array<{ count: number }> | undefined
): number {
  return (items ?? []).reduce(
    (sum, item) => sum + Math.max(0, Number(item.count) || 0),
    0
  )
}

function totalLabelKey(unit: string): string {
  switch (unit) {
    case 'token':
      return 'Total tokens'
    case 'call':
      return 'Total calls'
    default:
      return 'Total quota'
  }
}

function renderTotalQuota(
  usage: PlanUsage | null | undefined,
  planTotal: number,
  unit: string,
  t: (key: string) => string
): string {
  if (!usage) {
    return planTotal > 0 ? formatQuotaInUnit(planTotal, unit) : t('Unlimited')
  }
  if (usage.totalLimit <= 0) {
    return t('Unlimited')
  }
  return `${formatQuotaInUnit(usage.totalRemaining, unit)} / ${formatQuotaInUnit(usage.totalLimit, unit)}`
}
