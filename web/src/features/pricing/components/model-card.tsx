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
import { BookOpen, Copy, Play } from 'lucide-react'
import {
  memo,
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type ReactNode,
} from 'react'
import { useTranslation } from 'react-i18next'

import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { getLobeIcon } from '@/lib/lobe-icon'
import { cn } from '@/lib/utils'

import { Button } from '@/components/ui/button'
import { DEFAULT_TOKEN_UNIT, QUOTA_TYPE_VALUES } from '../constants'
import { getCardArtworkUrl, getProviderLogoUrl, isVideoModel } from '../lib/artwork'
import {
  getDynamicDisplayGroupRatio,
  getDynamicPricingSummary,
} from '../lib/dynamic-price'
import { parseTags } from '../lib/filters'
import { isTokenBasedModel } from '../lib/model-helpers'
import { formatPrice, formatRequestPrice } from '../lib/price'
import type { ModelCapability, PricingModel, TokenUnit } from '../types'
import { ProviderLogo } from './provider-logo'

export interface ModelCardProps {
  model: PricingModel
  onClick: () => void
  priceRate?: number
  usdExchangeRate?: number
  tokenUnit?: TokenUnit
  showRechargePrice?: boolean
  selectedGroup?: string
  perf?: unknown
  /** Documentation link configured under system-settings/billing/quota. */
  docsLink?: string
}

const CAPABILITY_SHORT_LABELS: Partial<Record<ModelCapability, string>> = {
  function_calling: 'Function calling',
  reasoning: 'Reasoning',
  vision: 'Vision',
  streaming: 'Streaming',
  web_search: 'Web search',
  json_mode: 'JSON mode',
  structured_output: 'Structured output',
  code_interpreter: 'Code interpreter',
  caching: 'Prompt caching',
  tools: 'Tools',
}

/** Stable pastel palette so every tag keeps the same color across renders. */
const TAG_PALETTES = [
  'bg-rose-50 text-rose-600 border-rose-200 dark:bg-rose-400/10 dark:text-rose-300 dark:border-rose-400/25',
  'bg-amber-50 text-amber-600 border-amber-200 dark:bg-amber-400/10 dark:text-amber-300 dark:border-amber-400/25',
  'bg-emerald-50 text-emerald-600 border-emerald-200 dark:bg-emerald-400/10 dark:text-emerald-300 dark:border-emerald-400/25',
  'bg-sky-50 text-sky-600 border-sky-200 dark:bg-sky-400/10 dark:text-sky-300 dark:border-sky-400/25',
  'bg-indigo-50 text-indigo-600 border-indigo-200 dark:bg-indigo-400/10 dark:text-indigo-300 dark:border-indigo-400/25',
  'bg-violet-50 text-violet-600 border-violet-200 dark:bg-violet-400/10 dark:text-violet-300 dark:border-violet-400/25',
  'bg-fuchsia-50 text-fuchsia-600 border-fuchsia-200 dark:bg-fuchsia-400/10 dark:text-fuchsia-300 dark:border-fuchsia-400/25',
  'bg-teal-50 text-teal-600 border-teal-200 dark:bg-teal-400/10 dark:text-teal-300 dark:border-teal-400/25',
] as const

function tagPalette(tag: string): string {
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = (hash * 31 + tag.charCodeAt(i)) | 0
  }
  return TAG_PALETTES[Math.abs(hash) % TAG_PALETTES.length]
}

function CardArtwork({ video, imageUrl }: { video: boolean; imageUrl: string }) {
  return (
    <div className='bg-[linear-gradient(135deg,#eaf4ff,#cfe2ff)] relative h-full min-h-[132px] w-full overflow-hidden rounded-lg dark:bg-[linear-gradient(135deg,rgba(30,64,175,0.35),rgba(15,23,42,0.4))]'>
      <div className='absolute -top-8 -right-8 size-24 rounded-full bg-white/40 blur-2xl dark:bg-blue-400/10' />
      <svg
        viewBox='0 0 200 140'
        className='absolute inset-0 h-full w-full'
        aria-hidden='true'
      >
        <g stroke='rgba(37,99,235,0.45)' strokeWidth='1' fill='none'>
          <path d='M100 18 L162 52 L100 86 L38 52 Z' fill='rgba(147,197,253,0.55)' />
          <path d='M38 52 L100 86 L100 132 L38 98 Z' fill='rgba(59,130,246,0.30)' />
          <path d='M162 52 L100 86 L100 132 L162 98 Z' fill='rgba(96,165,250,0.45)' />
          <path d='M69 35 L131 69 M131 35 L69 69 M69 70 L69 116 M131 70 L131 116' strokeDasharray='3 4' opacity='0.55' />
          <circle cx='100' cy='18' r='2.5' fill='#2563eb' stroke='none' />
          <circle cx='38' cy='52' r='2' fill='#2563eb' stroke='none' />
          <circle cx='162' cy='52' r='2' fill='#2563eb' stroke='none' />
        </g>
      </svg>
      {/* External placeholder artwork; falls back to the gradient above on error. */}
      <img
        key={imageUrl}
        src={imageUrl}
        alt=''
        loading='lazy'
        referrerPolicy='no-referrer'
        onError={(event) => {
          event.currentTarget.style.display = 'none'
        }}
        className='absolute inset-0 h-full w-full object-cover'
      />
      {video && (
        <span className='absolute inset-0 flex items-center justify-center'>
          <span className='flex size-9 items-center justify-center rounded-full border border-white/80 bg-white/70 text-blue-600 shadow-sm backdrop-blur-sm dark:border-white/20 dark:bg-white/15 dark:text-white'>
            <Play className='size-4 fill-current' />
          </span>
        </span>
      )}
    </div>
  )
}

export const ModelCard = memo(function ModelCard(props: ModelCardProps) {
  const { t } = useTranslation()
  const { copyToClipboard } = useCopyToClipboard()
  const tokenUnit = props.tokenUnit ?? DEFAULT_TOKEN_UNIT
  const priceRate = props.priceRate ?? 1
  const usdExchangeRate = props.usdExchangeRate ?? 1
  const showRechargePrice = props.showRechargePrice ?? false
  const isTokenBased = isTokenBasedModel(props.model)
  const tokenUnitLabel = tokenUnit === 'K' ? '1K' : '1M'
  const tags = parseTags(props.model.tags)
  const capabilities = props.model.capabilities ?? []
  const modelIconKey = props.model.icon || props.model.vendor_icon
  const modelIcon = modelIconKey ? getLobeIcon(modelIconKey, 20) : null
  const initial = props.model.model_name?.charAt(0).toUpperCase() || '?'
  const isDynamicPricing =
    props.model.billing_mode === 'tiered_expr' &&
    Boolean(props.model.billing_expr)
  const isHot = tags.some((tag) => ['hot', '热门'].includes(tag.toLowerCase()))
  const hotBadgeText = props.model.hot_label?.trim() || (isHot ? 'HOT' : '')
  const discountLabel = props.model.discount_label?.trim() ?? ''
  const isVideo = isVideoModel(props.model)
  const artworkUrl = getCardArtworkUrl(props.model)
  const providerLogoUrl = modelIcon
    ? undefined
    : getProviderLogoUrl(props.model.vendor_name)

  const dynamicSummary = isDynamicPricing
    ? getDynamicPricingSummary(props.model, {
        tokenUnit,
        showRechargePrice,
        priceRate,
        usdExchangeRate,
        groupRatioMultiplier: getDynamicDisplayGroupRatio(
          props.model,
          props.selectedGroup
        ),
      })
    : null

  let priceLabel: string
  let priceValue: ReactNode
  let priceUnit: string
  // Secondary line under the main price: input price shown in smaller text.
  let inputPrice: { label: string; value: string; unit: string } | null = null
  if (dynamicSummary && dynamicSummary.primaryEntries.length > 0) {
    const outputEntry = dynamicSummary.primaryEntries.find(
      (item) => item.field === 'outputPrice'
    )
    const entry = outputEntry ?? dynamicSummary.primaryEntries[0]
    priceLabel = t(entry.shortLabel)
    priceValue = entry.formatted
    priceUnit = ''
    const inputEntry = dynamicSummary.primaryEntries.find(
      (item) => item.field === 'inputPrice'
    )
    if (inputEntry && inputEntry !== entry) {
      inputPrice = {
        label: t(inputEntry.shortLabel),
        value: inputEntry.formatted,
        unit: '',
      }
    }
  } else if (isTokenBased) {
    priceLabel = t('Output')
    priceValue = formatPrice(
      props.model,
      'output',
      tokenUnit,
      showRechargePrice,
      priceRate,
      usdExchangeRate,
      props.selectedGroup
    )
    priceUnit = ` /${tokenUnitLabel}`
    inputPrice = {
      label: t('Input'),
      value: formatPrice(
        props.model,
        'input',
        tokenUnit,
        showRechargePrice,
        priceRate,
        usdExchangeRate,
        props.selectedGroup
      ),
      unit: ` /${tokenUnitLabel}`,
    }
  } else {
    // 固定价形态共用 model_price，但单位由计费模式决定：按秒 / 按张
    // 不能继续显示「/ request」，否则展示价与实际扣费口径对不上。
    const unitByQuotaType: Record<number, string> = {
      [QUOTA_TYPE_VALUES.PER_SECOND]: t('second'),
      [QUOTA_TYPE_VALUES.IMAGE_COUNT]: t('image'),
    }
    priceLabel = ''
    priceValue = formatRequestPrice(
      props.model,
      showRechargePrice,
      priceRate,
      usdExchangeRate,
      props.selectedGroup
    )
    priceUnit = ` / ${unitByQuotaType[props.model.quota_type] ?? t('request')}`
  }

  const displayTags = [
    ...tags.slice(0, 2),
    ...capabilities
      .slice(0, 2)
      .map((capability) => CAPABILITY_SHORT_LABELS[capability])
      .filter((label): label is string => Boolean(label)),
  ].slice(0, 3)

  const handleCopy = (e: ReactMouseEvent) => {
    e.stopPropagation()
    copyToClipboard(props.model.model_name || '')
  }

  const handleDocs = (e: ReactMouseEvent) => {
    e.stopPropagation()
    const target = props.docsLink?.trim() || '/docs'
    window.open(target, '_blank', 'noopener,noreferrer')
  }

  const handleDetails = (e: ReactMouseEvent) => {
    e.stopPropagation()
    props.onClick()
  }

  const handleKeyDown = (e: ReactKeyboardEvent) => {
    if (e.target !== e.currentTarget) return
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      props.onClick()
    }
  }

  return (
    <div
      role='button'
      tabIndex={0}
      onClick={props.onClick}
      onKeyDown={handleKeyDown}
      className='group border-border/60 hover:border-blue-400/70 bg-card relative flex cursor-pointer gap-3 overflow-hidden rounded-xl border p-4 shadow-[0_1px_2px_rgba(15,23,42,0.04)] transition-all duration-200 hover:-translate-y-0.5 hover:shadow-[0_12px_28px_-16px_rgba(37,99,235,0.45)] dark:border-white/10'
    >
      <div className='flex min-w-0 flex-1 flex-col'>
        <div className='flex items-center gap-1.5'>
          <span className='border-border/60 bg-background flex size-5 shrink-0 items-center justify-center rounded border text-[10px] font-bold'>
            {modelIcon ||
              (providerLogoUrl ? (
                <ProviderLogo name={props.model.vendor_name} size={16} />
              ) : (
                initial
              ))}
          </span>
          <span className='text-muted-foreground truncate text-xs font-medium'>
            {props.model.vendor_name || t('Unknown vendor')}
          </span>
          {hotBadgeText && (
            <span className='shrink-0 rounded-full bg-gradient-to-r from-orange-400 to-red-500 px-1.5 py-px text-[10px] leading-4 font-bold text-white'>
              {hotBadgeText}
            </span>
          )}
          {discountLabel && (
            <span className='shrink-0 rounded border border-rose-300 bg-rose-50 px-1.5 py-px text-[10px] leading-4 text-rose-600 dark:border-rose-400/30 dark:bg-rose-400/10 dark:text-rose-300'>
              {discountLabel}
            </span>
          )}
        </div>

        <div className='mt-1.5 flex min-w-0 items-center gap-1'>
          <h3
            className='text-foreground min-w-0 truncate text-[15px] leading-tight font-semibold'
            title={props.model.model_name}
          >
            {props.model.model_name}
          </h3>
          {/* 复制的是模型名称，按钮紧跟名称比浮在卡片右上角更容易理解。 */}
          <button
            type='button'
            onClick={handleCopy}
            className='text-muted-foreground hover:bg-accent hover:text-foreground focus-visible:ring-ring shrink-0 rounded p-1 opacity-60 transition-all hover:opacity-100 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:outline-none'
            title={t('Copy model name')}
            aria-label={t('Copy model name')}
          >
            <Copy className='size-3' />
          </button>
        </div>

        {displayTags.length > 0 && (
          <div className='mt-1.5 flex flex-wrap items-center gap-1'>
            {displayTags.map((tag) => (
              <span
                key={tag}
                className={cn(
                  'rounded-full border px-2 py-0.5 text-[10px] leading-none font-medium whitespace-nowrap',
                  tagPalette(tag)
                )}
              >
                {tag}
              </span>
            ))}
          </div>
        )}

        {props.model.description && (
          <p className='text-muted-foreground/90 mt-1.5 line-clamp-2 text-xs leading-relaxed'>
            {props.model.description}
          </p>
        )}

        <div className='mt-2 flex items-baseline gap-1'>
          <span className='text-muted-foreground text-[11px]'>
            {priceLabel || t('Price')}
          </span>
          <span className='font-mono text-lg font-bold text-red-500 tabular-nums dark:text-red-400'>
            {priceValue}
          </span>
          <span className='text-muted-foreground text-[11px]'>{priceUnit}</span>
        </div>
        {inputPrice && (
          <div className='text-muted-foreground/80 mt-0.5 flex items-baseline gap-1 text-[11px] leading-none'>
            <span>{inputPrice.label}</span>
            <span className='font-mono tabular-nums'>{inputPrice.value}</span>
            <span>{inputPrice.unit}</span>
          </div>
        )}

        <div className='mt-auto flex flex-wrap items-center gap-2 pt-3'>
          <Button
            size='sm'
            onClick={handleDocs}
            className='h-7 gap-1 px-2.5 text-xs'
          >
            <BookOpen className='size-3.5' />
            {t('Online Documentation')}
          </Button>
          <Button
            size='sm'
            variant='outline'
            onClick={handleDetails}
            className='h-7 px-2.5 text-xs'
          >
            {t('Model Details')}
          </Button>
        </div>
      </div>

      <div className='hidden w-32 shrink-0 sm:block lg:w-40'>
        <CardArtwork video={isVideo} imageUrl={artworkUrl} />
      </div>

      {isDynamicPricing && !dynamicSummary?.isSpecialExpression && (
        <span className='border-amber-300/80 bg-amber-50 text-amber-600 absolute right-2 bottom-2 rounded-full border px-2 py-0.5 text-[10px] leading-none font-medium dark:border-amber-400/30 dark:bg-amber-400/10 dark:text-amber-300'>
          {t('Tiered pricing')}
        </span>
      )}
    </div>
  )
})
