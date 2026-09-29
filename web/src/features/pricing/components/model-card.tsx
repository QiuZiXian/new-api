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
import { Copy } from 'lucide-react'
import { memo, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { getLobeIcon } from '@/lib/lobe-icon'
import { cn } from '@/lib/utils'

import { DEFAULT_TOKEN_UNIT } from '../constants'
import {
  getDynamicDisplayGroupRatio,
  getDynamicPricingSummary,
} from '../lib/dynamic-price'
import { parseTags } from '../lib/filters'
import { isTokenBasedModel } from '../lib/model-helpers'
import { formatPrice, formatRequestPrice } from '../lib/price'
import type { ModelCapability, PricingModel, TokenUnit } from '../types'

export interface ModelCardProps {
  model: PricingModel
  onClick: () => void
  priceRate?: number
  usdExchangeRate?: number
  tokenUnit?: TokenUnit
  showRechargePrice?: boolean
  selectedGroup?: string
  perf?: unknown
}

const TAG_PALETTES = [
  'bg-emerald-50 text-emerald-600 border-emerald-200/70 dark:bg-emerald-400/10 dark:text-emerald-300 dark:border-emerald-400/20',
  'bg-sky-50 text-sky-600 border-sky-200/70 dark:bg-sky-400/10 dark:text-sky-300 dark:border-sky-400/20',
  'bg-cyan-50 text-cyan-600 border-cyan-200/70 dark:bg-cyan-400/10 dark:text-cyan-300 dark:border-cyan-400/20',
  'bg-violet-50 text-violet-600 border-violet-200/70 dark:bg-violet-400/10 dark:text-violet-300 dark:border-violet-400/20',
  'bg-amber-50 text-amber-600 border-amber-200/70 dark:bg-amber-400/10 dark:text-amber-300 dark:border-amber-400/20',
] as const

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

function tagPalette(tag: string): string {
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = (hash * 31 + tag.charCodeAt(i)) | 0
  }
  return TAG_PALETTES[Math.abs(hash) % TAG_PALETTES.length]
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
  const modelIcon = modelIconKey ? getLobeIcon(modelIconKey, 32) : null
  const initial = props.model.model_name?.charAt(0).toUpperCase() || '?'
  const isDynamicPricing =
    props.model.billing_mode === 'tiered_expr' &&
    Boolean(props.model.billing_expr)
  const isHot = tags.some((tag) => ['hot', '热门'].includes(tag.toLowerCase()))
  // Badge text is configured per model in the admin; empty means "no badge".
  // The legacy tag-derived HOT badge stays as a fallback for existing setups.
  const hotBadgeText = props.model.hot_label?.trim() || (isHot ? 'HOT' : '')
  const discountLabel = props.model.discount_label?.trim() ?? ''

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
  if (dynamicSummary && dynamicSummary.primaryEntries.length > 0) {
    const entry = dynamicSummary.primaryEntries[0]
    priceLabel = t(entry.shortLabel)
    priceValue = entry.formatted
    priceUnit = ''
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
  } else {
    priceLabel = ''
    priceValue = formatRequestPrice(
      props.model,
      showRechargePrice,
      priceRate,
      usdExchangeRate,
      props.selectedGroup
    )
    priceUnit = ` / ${t('request')}`
  }

  const displayTags = [
    ...tags.slice(0, 2),
    ...capabilities
      .slice(0, 2)
      .map((capability) => CAPABILITY_SHORT_LABELS[capability])
      .filter((label): label is string => Boolean(label)),
  ].slice(0, 4)

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    copyToClipboard(props.model.model_name || '')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
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
      className='group border-border/50 hover:border-blue-500/40 relative flex cursor-pointer flex-col items-center overflow-hidden rounded-2xl border bg-gradient-to-b from-card via-card to-muted/45 px-4 pt-8 pb-4 text-center shadow-[0_1px_2px_rgba(15,23,42,0.06),0_10px_22px_-12px_rgba(15,23,42,0.20),inset_0_1px_0_rgba(255,255,255,0.85)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_20px_38px_-16px_rgba(37,99,235,0.42),inset_0_1px_0_rgba(255,255,255,0.9)] dark:border-white/10 dark:shadow-[0_12px_26px_-14px_rgba(0,0,0,0.7),inset_0_1px_0_rgba(255,255,255,0.06)]'
    >
      {/* Top-left popularity badge, text comes from backend settings */}
      {hotBadgeText && (
        <span className='absolute top-2.5 left-2.5 z-20 rounded-full bg-gradient-to-r from-orange-400 to-red-500 px-2 py-0.5 text-[10px] leading-none font-bold tracking-wide text-white shadow-[0_2px_6px_-1px_rgba(249,115,22,0.55)]'>
          {hotBadgeText}
        </span>
      )}

      {/* Top-right discount ribbon, text comes from backend settings */}
      {discountLabel && (
        <div className='pointer-events-none absolute top-0 right-0 z-20 h-16 w-16 overflow-hidden rounded-tr-2xl'>
          <span className='absolute top-[14px] right-[-30px] w-24 rotate-45 overflow-hidden bg-gradient-to-r from-rose-500 to-red-500 py-[3px] text-center text-[10px] leading-[14px] font-bold tracking-wide text-ellipsis whitespace-nowrap text-white shadow-[0_2px_8px_-2px_rgba(225,29,72,0.6)]'>
            {discountLabel}
          </span>
        </div>
      )}

      {/* Top-right badges: tiered pricing + hover copy */}
      <div
        className={cn(
          'absolute top-2.5 z-10 flex items-center justify-end gap-1.5',
          discountLabel ? 'right-16' : 'right-2.5'
        )}
      >
        {isDynamicPricing && !dynamicSummary?.isSpecialExpression && (
          <span className='border-amber-300/80 bg-amber-50 text-amber-600 rounded-full border px-2 py-0.5 text-[10px] leading-none font-medium dark:border-amber-400/30 dark:bg-amber-400/10 dark:text-amber-300'>
            {t('Tiered pricing')}
          </span>
        )}
        <button
          type='button'
          onClick={handleCopy}
          className='border-border/60 bg-background/90 text-muted-foreground hover:text-foreground rounded-md border p-1.5 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100'
          title={t('Copy')}
          aria-label={t('Copy')}
        >
          <Copy className='size-3.5' />
        </button>
      </div>

      {/* Icon + name + vendor */}
      <div className='border-border/50 bg-background/80 flex size-14 shrink-0 items-center justify-center rounded-2xl border shadow-[0_2px_8px_-3px_rgba(15,23,42,0.18)] dark:bg-muted/40'>
        {modelIcon || (
          <span className='text-muted-foreground text-lg font-bold'>
            {initial}
          </span>
        )}
      </div>
      <h3
        className='text-foreground mt-3 max-w-full truncate text-[15px] leading-tight font-semibold'
        title={props.model.model_name}
      >
        {props.model.model_name}
      </h3>
      {props.model.vendor_name && (
        <p className='text-muted-foreground/80 mt-0.5 truncate text-xs'>
          {props.model.vendor_name}
        </p>
      )}

      {/* Capability / tag pills */}
      {displayTags.length > 0 && (
        <div className='mt-2.5 flex max-w-full flex-wrap items-center justify-center gap-1'>
          {displayTags.map((tag) => (
            <span
              key={tag}
              className={cn(
                'rounded border px-1.5 py-0.5 text-[10px] leading-none whitespace-nowrap',
                tagPalette(tag)
              )}
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Price footer */}
      <div className='border-border/50 mt-3 flex w-full items-baseline justify-center gap-1 border-t pt-3'>
        {priceLabel && (
          <span className='text-muted-foreground text-xs'>{priceLabel}</span>
        )}
        <span className='font-mono text-[18px] font-bold text-red-500 tabular-nums dark:text-red-400'>
          {priceValue}
        </span>
        {priceUnit && (
          <span className='text-muted-foreground text-xs'>{priceUnit}</span>
        )}
      </div>
    </div>
  )
})
