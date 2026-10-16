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
import * as z from 'zod'

import { combineBillingExpr } from '@/features/pricing/lib/billing-expr'

import { formatPricingNumber } from './pricing-format'

export const createModelPricingSchema = (t: (key: string) => string) =>
  z.object({
    name: z.string().min(1, t('Model name is required')),
    price: z.string().optional(),
    ratio: z.string().optional(),
    cacheRatio: z.string().optional(),
    createCacheRatio: z.string().optional(),
    completionRatio: z.string().optional(),
    imageRatio: z.string().optional(),
    imageCompletionRatio: z.string().optional(),
    audioRatio: z.string().optional(),
    audioCompletionRatio: z.string().optional(),
  })

export type ModelPricingFormValues = z.infer<
  ReturnType<typeof createModelPricingSchema>
>

export type PricingMode = 'per-token' | 'per-request' | 'per-second' | 'tiered_expr'

// VIDEO_SECOND_DEFAULT_TIER 是按秒计费的兜底档位，同时作为分档倍率的基准价。
// 后端常量 VideoSecondDefaultVariant 与之对应。
export const VIDEO_SECOND_DEFAULT_TIER = 'default'

export type LaneKey =
  | 'completion'
  | 'cache'
  | 'createCache'
  | 'image'
  | 'imageOutput'
  | 'audioInput'
  | 'audioOutput'

export type ModelRatioData = {
  name: string
  price?: string
  ratio?: string
  cacheRatio?: string
  createCacheRatio?: string
  completionRatio?: string
  imageRatio?: string
  imageCompletionRatio?: string
  audioRatio?: string
  audioCompletionRatio?: string
  billingMode?: PricingMode
  billingExpr?: string
  requestRuleExpr?: string
  videoSecondPrice?: Record<string, string>
}

export type PreviewRow = {
  key: string
  label: string
  value: string
  multiline?: boolean
}

export const numericDraftRegex = /^(\d+(\.\d*)?|\.\d*)?$/

export const EMPTY_LANE_PRICES: Record<LaneKey, string> = {
  completion: '',
  cache: '',
  createCache: '',
  image: '',
  imageOutput: '',
  audioInput: '',
  audioOutput: '',
}

export const EMPTY_LANE_ENABLED: Record<LaneKey, boolean> = {
  completion: false,
  cache: false,
  createCache: false,
  image: false,
  imageOutput: false,
  audioInput: false,
  audioOutput: false,
}

// inputLaneByLane 记录「输出档依赖哪个输入档的价格」：这些输出倍率是相对
// 对应输入价计算的，输入档没配时输出档不能填。
export const inputLaneByLane: Partial<Record<LaneKey, LaneKey>> = {
  audioOutput: 'audioInput',
  imageOutput: 'image',
}

export const ratioFieldByLane: Record<LaneKey, keyof ModelPricingFormValues> = {
  completion: 'completionRatio',
  cache: 'cacheRatio',
  createCache: 'createCacheRatio',
  image: 'imageRatio',
  imageOutput: 'imageCompletionRatio',
  audioInput: 'audioRatio',
  audioOutput: 'audioCompletionRatio',
}

export const laneConfigs: Array<{
  key: LaneKey
  titleKey: string
  descriptionKey: string
  placeholder: string
}> = [
  {
    key: 'completion',
    titleKey: 'Completion price',
    descriptionKey: 'Output token price for generated tokens.',
    placeholder: '15',
  },
  {
    key: 'cache',
    titleKey: 'Cache read price',
    descriptionKey: 'Token price for cache reads.',
    placeholder: '0.3',
  },
  {
    key: 'createCache',
    titleKey: 'Cache write price',
    descriptionKey: 'Token price for creating cache entries.',
    placeholder: '3.75',
  },
  {
    key: 'image',
    titleKey: 'Image input price',
    descriptionKey: 'Token price for image input.',
    placeholder: '2.5',
  },
  {
    key: 'imageOutput',
    titleKey: 'Image output price',
    descriptionKey:
      'Price multiplier for generated image tokens, relative to the image input price. Defaults to 1 when unset.',
    placeholder: '4',
  },
  {
    key: 'audioInput',
    titleKey: 'Audio input price',
    descriptionKey: 'Token price for audio input.',
    placeholder: '3.81',
  },
  {
    key: 'audioOutput',
    titleKey: 'Audio output price',
    descriptionKey: 'Token price for audio output.',
    placeholder: '15.11',
  },
]

export function hasValue(value: unknown): boolean {
  return (
    value !== '' && value !== null && value !== undefined && value !== false
  )
}

export function toNumberOrNull(value: unknown): number | null {
  if (!hasValue(value) && value !== 0) return null
  const num = Number(value)
  return Number.isFinite(num) ? num : null
}

function ratioToBasePrice(ratio: unknown): string {
  const num = toNumberOrNull(ratio)
  if (num === null) return ''
  return formatPricingNumber(num * 2)
}

function deriveLanePrice(
  ratio: unknown,
  denominator: unknown,
  fallback = ''
): string {
  const ratioNumber = toNumberOrNull(ratio)
  const denominatorNumber = toNumberOrNull(denominator)
  if (ratioNumber === null || denominatorNumber === null) return fallback
  return formatPricingNumber(ratioNumber * denominatorNumber)
}

export type VideoSecondTierRow = { id: string; variant: string; price: string }

export const EMPTY_SECOND_TIERS: VideoSecondTierRow[] = [
  { id: 'tier-default', variant: VIDEO_SECOND_DEFAULT_TIER, price: '' },
]

let secondTierSeq = 0

// newSecondTierId 给新增档位行一个稳定的 React key（行可以被删除，
// 所以不能拿数组下标当 key）。
export function newSecondTierId(): string {
  secondTierSeq += 1
  return `tier-${secondTierSeq}`
}

// resolvePricingMode 由已保存的定价数据回推编辑器的模式页签。
export function resolvePricingMode(
  data?: Pick<ModelRatioData, 'billingMode' | 'price'> | null
): PricingMode {
  if (data?.billingMode === 'tiered_expr') return 'tiered_expr'
  if (data?.billingMode === 'per-second') return 'per-second'
  if (data?.price) return 'per-request'
  return 'per-token'
}

// toSecondTierRows 把 per-model 的分档秒价表还原成可编辑的行列表。
export function toSecondTierRows(
  map?: Record<string, string>
): VideoSecondTierRow[] {
  if (!map) return [{ ...EMPTY_SECOND_TIERS[0] }]
  const rows = Object.entries(map).map(([variant, price]) => ({
    id: newSecondTierId(),
    variant,
    price,
  }))
  if (rows.length === 0) return [{ ...EMPTY_SECOND_TIERS[0] }]
  return rows
}

// fromSecondTierRows 把行列表收敛成持久化用的「档位 → 每秒单价」，
// 丢弃档位名或价格为空的行；全部为空时返回 undefined（表示未启用按秒计费）。
export function fromSecondTierRows(
  rows: VideoSecondTierRow[]
): Record<string, string> | undefined {
  const map: Record<string, string> = {}
  rows.forEach(({ variant, price }) => {
    const key = variant.trim().toLowerCase()
    if (!key || price.trim() === '') return
    const num = Number(price)
    if (!Number.isFinite(num)) return
    map[key] = String(num)
  })
  return Object.keys(map).length > 0 ? map : undefined
}

export function createInitialLaneState(data?: ModelRatioData | null) {
  if (!data) {
    return {
      promptPrice: '',
      prices: { ...EMPTY_LANE_PRICES },
      enabled: { ...EMPTY_LANE_ENABLED },
      secondTiers: toSecondTierRows(undefined),
    }
  }

  const promptPrice = ratioToBasePrice(data.ratio)
  const audioInputPrice = deriveLanePrice(data.audioRatio, promptPrice)
  const imageInputPrice = deriveLanePrice(data.imageRatio, promptPrice)
  const prices: Record<LaneKey, string> = {
    completion: deriveLanePrice(data.completionRatio, promptPrice),
    cache: deriveLanePrice(data.cacheRatio, promptPrice),
    createCache: deriveLanePrice(data.createCacheRatio, promptPrice),
    image: imageInputPrice,
    imageOutput: deriveLanePrice(data.imageCompletionRatio, imageInputPrice),
    audioInput: audioInputPrice,
    audioOutput: deriveLanePrice(data.audioCompletionRatio, audioInputPrice),
  }

  return {
    promptPrice,
    prices,
    enabled: {
      completion: hasValue(data.completionRatio),
      cache: hasValue(data.cacheRatio),
      createCache: hasValue(data.createCacheRatio),
      image: hasValue(data.imageRatio),
      imageOutput: hasValue(data.imageCompletionRatio),
      audioInput: hasValue(data.audioRatio),
      audioOutput: hasValue(data.audioCompletionRatio),
    },
    secondTiers: toSecondTierRows(data.videoSecondPrice),
  }
}

export function buildPreviewRows(
  values: ModelPricingFormValues,
  mode: PricingMode,
  billingExpr: string,
  requestRuleExpr: string,
  promptPrice: string,
  lanePrices: Record<LaneKey, string>,
  laneEnabled: Record<LaneKey, boolean>,
  secondTiers: VideoSecondTierRow[],
  t: (key: string) => string
): PreviewRow[] {
  if (mode === 'tiered_expr') {
    const effectiveExpr = combineBillingExpr(billingExpr, requestRuleExpr)
    return [
      { key: 'mode', label: 'BillingMode', value: 'tiered_expr' },
      {
        key: 'expr',
        label: t('Expression'),
        value: effectiveExpr || t('Empty'),
        multiline: true,
      },
    ]
  }

  if (mode === 'per-request') {
    return [
      {
        key: 'price',
        label: 'ModelPrice',
        value: values.price || t('Empty'),
      },
    ]
  }

  if (mode === 'per-second') {
    const tiers = fromSecondTierRows(secondTiers)
    if (!tiers) {
      return [
        { key: 'mode', label: 'BillingMode', value: 'per_second' },
        { key: 'default', label: t('Default tier'), value: t('Empty') },
      ]
    }
    const rows: PreviewRow[] = [
      { key: 'mode', label: 'BillingMode', value: 'per_second' },
      {
        key: 'default',
        label: `${t('Default tier')} ($/s)`,
        value: tiers[VIDEO_SECOND_DEFAULT_TIER]
          ? `$${tiers[VIDEO_SECOND_DEFAULT_TIER]}`
          : t('Empty'),
      },
    ]
    Object.entries(tiers).forEach(([variant, price]) => {
      if (variant === VIDEO_SECOND_DEFAULT_TIER) return
      rows.push({
        key: variant,
        label: `${variant} ($/s)`,
        value: `$${price}`,
      })
    })
    return rows
  }

  return [
    {
      key: 'inputPrice',
      label: t('Input price'),
      value: promptPrice ? `$${promptPrice}` : t('Empty'),
    },
    {
      key: 'completion',
      label: t('Completion price'),
      value:
        laneEnabled.completion && lanePrices.completion
          ? `$${lanePrices.completion}`
          : t('Empty'),
    },
    {
      key: 'cache',
      label: t('Cache read price'),
      value:
        laneEnabled.cache && lanePrices.cache
          ? `$${lanePrices.cache}`
          : t('Empty'),
    },
    {
      key: 'createCache',
      label: t('Cache write price'),
      value:
        laneEnabled.createCache && lanePrices.createCache
          ? `$${lanePrices.createCache}`
          : t('Empty'),
    },
    {
      key: 'image',
      label: t('Image input price'),
      value:
        laneEnabled.image && lanePrices.image
          ? `$${lanePrices.image}`
          : t('Empty'),
    },
    {
      key: 'imageOutput',
      label: t('Image output price'),
      value:
        laneEnabled.imageOutput && lanePrices.imageOutput
          ? `$${lanePrices.imageOutput}`
          : t('Empty'),
    },
    {
      key: 'audio',
      label: t('Audio input price'),
      value:
        laneEnabled.audioInput && lanePrices.audioInput
          ? `$${lanePrices.audioInput}`
          : t('Empty'),
    },
    {
      key: 'audioCompletion',
      label: t('Audio output price'),
      value:
        laneEnabled.audioOutput && lanePrices.audioOutput
          ? `$${lanePrices.audioOutput}`
          : t('Empty'),
    },
  ]
}
