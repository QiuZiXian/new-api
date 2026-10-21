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
import { useNavigate } from '@tanstack/react-router'
import { PackageOpen } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { PublicLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { Skeleton } from '@/components/ui/skeleton'
import {
  getSelfSubscriptionFull,
  getShowcasePlans,
} from '@/features/subscriptions/api'
import { SubscriptionPurchaseDialog } from '@/features/subscriptions/components/dialogs/subscription-purchase-dialog'
import type {
  PlanRecord,
  SelfSubscriptionData,
  UserSubscriptionRecord,
} from '@/features/subscriptions/types'
import { useTopupInfo } from '@/features/wallet/hooks'
import { getSelf } from '@/lib/api'
import { useAuthStore } from '@/stores/auth-store'

import { CurrentSubscriptionStrip } from './components/current-subscription-strip'
import { PlanCard } from './components/plan-card'
import { PlansHero } from './components/plans-hero'

type EpayMethod = { type: string; name?: string }

function getEpayMethods(payMethods: unknown): EpayMethod[] {
  if (!Array.isArray(payMethods)) return []
  return payMethods.filter(
    (item): item is EpayMethod =>
      !!item &&
      typeof item === 'object' &&
      typeof (item as EpayMethod).type === 'string' &&
      (item as EpayMethod).type !== 'stripe' &&
      (item as EpayMethod).type !== 'creem'
  )
}

export function Plans() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { auth } = useAuthStore()
  const isAuthed = Boolean(auth.user)

  const [plans, setPlans] = useState<PlanRecord[]>([])
  const [selfData, setSelfData] = useState<SelfSubscriptionData | null>(null)
  const [userQuota, setUserQuota] = useState(0)
  const [loading, setLoading] = useState(true)
  const [selectedPlan, setSelectedPlan] = useState<PlanRecord | null>(null)
  const [purchaseOpen, setPurchaseOpen] = useState(false)

  const { topupInfo } = useTopupInfo({ enabled: isAuthed })

  const fetchPlans = useCallback(async () => {
    try {
      const res = await getShowcasePlans()
      if (res.success) {
        setPlans(res.data || [])
      } else {
        setPlans([])
      }
    } catch {
      setPlans([])
    }
  }, [])

  const fetchSelf = useCallback(async () => {
    if (!isAuthed) {
      setSelfData(null)
      setUserQuota(0)
      return
    }
    try {
      const [subscriptionRes, selfRes] = await Promise.all([
        getSelfSubscriptionFull(),
        getSelf(),
      ])
      if (subscriptionRes.success && subscriptionRes.data) {
        setSelfData(subscriptionRes.data)
      }
      const selfPayload = selfRes as
        | { success?: boolean; data?: { quota?: number } }
        | undefined
      if (selfPayload?.success && selfPayload.data) {
        setUserQuota(Number(selfPayload.data.quota || 0))
      }
    } catch {
      // Best-effort: the plans page still renders without subscription state.
    }
  }, [isAuthed])

  useEffect(() => {
    let cancelled = false
    const init = async () => {
      setLoading(true)
      await Promise.all([fetchPlans(), fetchSelf()])
      if (!cancelled) setLoading(false)
    }
    void init()
    return () => {
      cancelled = true
    }
  }, [fetchPlans, fetchSelf])

  const activeSubscriptions: UserSubscriptionRecord[] = useMemo(
    () => selfData?.subscriptions ?? [],
    [selfData]
  )
  const allSubscriptions: UserSubscriptionRecord[] = useMemo(
    () => selfData?.all_subscriptions ?? [],
    [selfData]
  )

  const planTitles = useMemo(() => {
    const map = new Map<number, string>()
    for (const record of plans) {
      if (record.plan?.id) {
        map.set(record.plan.id, record.plan.title || '')
      }
    }
    return map
  }, [plans])

  const purchaseCountMap = useMemo(() => {
    const map = new Map<number, number>()
    for (const record of allSubscriptions) {
      const planId = record.subscription?.plan_id
      if (!planId) continue
      map.set(planId, (map.get(planId) || 0) + 1)
    }
    return map
  }, [allSubscriptions])

  const epayMethods = useMemo(
    () => getEpayMethods(topupInfo?.pay_methods),
    [topupInfo?.pay_methods]
  )

  const handleSubscribe = useCallback(
    (record: PlanRecord) => {
      if (!isAuthed) {
        void navigate({ to: '/sign-in', search: { redirect: '/plans' } })
        return
      }
      setSelectedPlan(record)
      setPurchaseOpen(true)
    },
    [isAuthed, navigate]
  )

  const handleManage = useCallback(() => {
    void navigate({ to: '/wallet' })
  }, [navigate])

  const renderPlansContent = () => {
    if (loading) {
      return (
        <div className='grid gap-5 sm:grid-cols-2 lg:grid-cols-3'>
          {['one', 'two', 'three'].map((key) => (
            <Skeleton key={key} className='h-[420px] w-full rounded-2xl' />
          ))}
        </div>
      )
    }

    if (plans.length === 0) {
      return (
        <div className='border-border/70 flex flex-col items-center justify-center rounded-2xl border border-dashed px-6 py-20 text-center'>
          <div className='bg-muted/50 text-muted-foreground flex size-12 items-center justify-center rounded-full'>
            <PackageOpen className='size-6' />
          </div>
          <p className='mt-4 text-sm font-medium'>{t('No plans available')}</p>
          <p className='text-muted-foreground mt-1 max-w-md text-xs'>
            {t(
              'Subscription plans will show up here once an administrator publishes them.'
            )}
          </p>
        </div>
      )
    }

    return (
      <div className='grid items-stretch gap-5 sm:grid-cols-2 lg:grid-cols-3'>
        {plans.map((record, index) =>
          record.plan ? (
            <PlanCard
              key={record.plan.id}
              plan={record.plan}
              index={index}
              purchaseCount={purchaseCountMap.get(record.plan.id) || 0}
              onSubscribe={() => handleSubscribe(record)}
            />
          ) : null
        )}
      </div>
    )
  }

  return (
    <PublicLayout showMainContainer={false}>
      <PlansHero />

      <PageTransition className='mx-auto w-full max-w-[1200px] px-4 pt-2 pb-16'>
        {isAuthed && activeSubscriptions.length > 0 ? (
          <div className='mb-6'>
            <CurrentSubscriptionStrip
              subscriptions={activeSubscriptions}
              planTitles={planTitles}
              onManage={handleManage}
            />
          </div>
        ) : null}

        {renderPlansContent()}
      </PageTransition>

      <SubscriptionPurchaseDialog
        open={purchaseOpen}
        onOpenChange={(open) => {
          setPurchaseOpen(open)
          if (!open) {
            void fetchSelf()
          }
        }}
        plan={selectedPlan}
        enableStripe={Boolean(topupInfo?.enable_stripe_topup)}
        enableCreem={Boolean(topupInfo?.enable_creem_topup)}
        enableWaffoPancake={Boolean(topupInfo?.enable_waffo_pancake_topup)}
        enableOnlineTopUp={Boolean(topupInfo?.enable_online_topup)}
        epayMethods={epayMethods}
        userQuota={userQuota}
        onPurchaseSuccess={fetchSelf}
        purchaseLimit={
          selectedPlan?.plan?.max_purchase_per_user
            ? Number(selectedPlan.plan.max_purchase_per_user)
            : undefined
        }
        purchaseCount={
          selectedPlan?.plan?.id
            ? purchaseCountMap.get(selectedPlan.plan.id)
            : undefined
        }
      />
    </PublicLayout>
  )
}
