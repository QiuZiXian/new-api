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
import { Crown } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import type { UserSubscriptionRecord } from '@/features/subscriptions/types'
import { formatQuota } from '@/lib/format'

type CurrentSubscriptionStripProps = {
  subscriptions: UserSubscriptionRecord[]
  planTitles: Map<number, string>
  onManage: () => void
}

export function CurrentSubscriptionStrip({
  subscriptions,
  planTitles,
  onManage,
}: CurrentSubscriptionStripProps) {
  const { t } = useTranslation()
  if (subscriptions.length === 0) return null

  return (
    <div className='rounded-2xl border border-sky-200/70 bg-gradient-to-r from-sky-50 via-white to-violet-50 p-4 dark:border-white/10 dark:from-slate-900 dark:via-slate-950 dark:to-slate-900'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div className='flex min-w-0 items-start gap-3'>
          <div className='flex size-9 shrink-0 items-center justify-center rounded-lg bg-white text-amber-500 shadow-sm dark:bg-white/10'>
            <Crown className='size-4' />
          </div>
          <div className='min-w-0 space-y-1'>
            {subscriptions.map((record) => {
              const subscription = record.subscription
              const total = Number(subscription.amount_total || 0)
              const used = Number(subscription.amount_used || 0)
              const remaining = total > 0 ? Math.max(0, total - used) : 0
              const endTime = new Date(
                Number(subscription.end_time || 0) * 1000
              ).toLocaleDateString()
              const title =
                planTitles.get(subscription.plan_id) || t('Subscription')
              return (
                <div key={subscription.id} className='min-w-0'>
                  <p className='truncate text-sm font-medium'>
                    {t('Current plan')} · {title}
                  </p>
                  <p className='text-muted-foreground text-xs'>
                    {t('Until')} {endTime}
                    {total > 0 ? (
                      <>
                        {' · '}
                        {t('Remaining')} {formatQuota(remaining)}
                      </>
                    ) : null}
                  </p>
                </div>
              )
            })}
          </div>
        </div>
        <Button variant='outline' size='sm' onClick={onManage}>
          {t('Manage subscriptions')}
        </Button>
      </div>
    </div>
  )
}
