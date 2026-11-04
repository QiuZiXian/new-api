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
import { splitBillingExprAndRequestRules } from '@/features/pricing/lib/billing-expr'

import { safeJsonParse } from '../utils/json-parser'
import {
  IMAGE_COUNT_PRICE_KEYS,
  type PricingMode,
} from './model-pricing-core'
import { formatPricingNumber } from './pricing-format'

export type ModelPricingSnapshotInput = {
  modelPrice: string
  modelRatio: string
  cacheRatio: string
  createCacheRatio: string
  completionRatio: string
  imageRatio: string
  imageCompletionRatio: string
  videoSecondPrice: string
  imageCountPrice: string
  audioRatio: string
  audioCompletionRatio: string
  billingMode: string
  billingExpr: string
}

export type ModelPricingSnapshot = {
  name: string
  price?: string
  ratio?: string
  cacheRatio?: string
  createCacheRatio?: string
  completionRatio?: string
  imageRatio?: string
  imageCompletionRatio?: string
  videoSecondPrice?: Record<string, string>
  imageCountPrice?: Record<string, string>
  audioRatio?: string
  audioCompletionRatio?: string
  billingMode?: string
  billingExpr?: string
  requestRuleExpr?: string
  hasConflict: boolean
}

export type ModelRow = ModelPricingSnapshot & {
  saved?: ModelPricingSnapshot
  draft?: ModelPricingSnapshot
  isDraftChanged: boolean
  isDraftDeleted: boolean
  isDraftNew: boolean
}

export const hasPricingValue = (value?: string) =>
  value !== undefined && value !== ''

export const isBasePricingUnset = (snapshot?: ModelPricingSnapshot) =>
  !snapshot ||
  (snapshot.billingMode !== 'tiered_expr' &&
    !hasPricingValue(snapshot.price) &&
    !hasPricingValue(snapshot.ratio) &&
    !snapshot.imageCountPrice)

const toNumberOrNull = (value?: string) => {
  if (!hasPricingValue(value)) return null
  const num = Number(value)
  return Number.isFinite(num) ? num : null
}

const ratioToPrice = (ratio?: string, denominator?: string) => {
  const ratioNumber = toNumberOrNull(ratio)
  const denominatorNumber = denominator ? toNumberOrNull(denominator) : 2
  if (ratioNumber === null || denominatorNumber === null) return ''
  return formatPricingNumber(ratioNumber * denominatorNumber)
}

export const getModeLabel = (mode?: string) => {
  if (mode === 'per-request') return 'Per-request'
  if (mode === 'per-second') return 'Per-second'
  if (mode === 'per-image') return 'Per-image'
  if (mode === 'tiered_expr') return 'Expression'
  return 'Per-token'
}

export const getModeVariant = (
  mode?: string
): 'warning' | 'info' | 'success' => {
  if (mode === 'per-request') return 'warning'
  if (mode === 'per-second') return 'warning'
  if (mode === 'per-image') return 'warning'
  if (mode === 'tiered_expr') return 'info'
  return 'success'
}

const getExpressionSummary = (
  row: ModelPricingSnapshot,
  t: (key: string) => string
) => {
  const tierCount = (row.billingExpr?.match(/tier\(/g) || []).length
  if (tierCount > 0) {
    return `${t('Tiered pricing')} · ${tierCount} ${t('tiers')}`
  }
  return t('Expression pricing')
}

export const getPriceSummary = (
  row: ModelPricingSnapshot,
  t: (key: string) => string
) => {
  if (row.billingMode === 'tiered_expr') {
    return getExpressionSummary(row, t)
  }
  if (row.billingMode === 'per-request') {
    return row.price ? `$${row.price} / ${t('request')}` : t('Unset price')
  }
  if (row.billingMode === 'per-second') {
    const tiers = row.videoSecondPrice ?? {}
    const defaultPrice = tiers.default
    if (defaultPrice && hasPricingValue(defaultPrice)) {
      const tierCount = Object.keys(tiers).length
      return tierCount > 1
        ? `$${defaultPrice} / ${t('second')} · ${tierCount} ${t('tiers')}`
        : `$${defaultPrice} / ${t('second')}`
    }
    return t('Unset price')
  }
  if (row.billingMode === 'per-image') {
    const counts = row.imageCountPrice ?? {}
    const outputPrices = ['output', 'output_1k', 'output_2k']
      .map((key) => Number(counts[key] ?? ''))
      .filter((value) => Number.isFinite(value) && value > 0)
    if (outputPrices.length === 0) return t('Unset price')
    const tierCount = ['output_1k', 'output_2k'].filter((key) =>
      hasPricingValue(counts[key])
    ).length
    const tiers = tierCount > 0 ? ` · ${tierCount} ${t('tiers')}` : ''
    const inputPrice = hasPricingValue(counts.input)
      ? ` · ${t('Input')} $${counts.input}`
      : ''
    return `${t('Output')} $${Math.max(...outputPrices)} / ${t('image')}${tiers}${inputPrice}`
  }

  const inputPrice = ratioToPrice(row.ratio)
  if (!inputPrice) return t('Unset price')

  const extraCount = [
    row.completionRatio,
    row.cacheRatio,
    row.createCacheRatio,
    row.imageRatio,
    row.imageCompletionRatio,
    row.audioRatio,
    row.audioCompletionRatio,
  ].filter(hasPricingValue).length

  return extraCount > 0
    ? `${t('Input')} $${inputPrice} · ${extraCount} ${t('extras')}`
    : `${t('Input')} $${inputPrice}`
}

export const getPriceDetail = (
  row: ModelPricingSnapshot,
  t: (key: string) => string
) => {
  if (row.billingMode === 'tiered_expr') {
    return row.requestRuleExpr
      ? t('Includes request rules')
      : t('Expression based')
  }
  if (row.billingMode === 'per-request') {
    return t('Fixed request price')
  }
  if (row.billingMode === 'per-second') {
    const tiers = row.videoSecondPrice ?? {}
    const names = Object.keys(tiers).filter((key) => key !== 'default')
    return names.length > 0
      ? `${t('Per-second price')} · ${names.join(', ')}`
      : t('Single per-second price')
  }
  if (row.billingMode === 'per-image') {
    const counts = row.imageCountPrice ?? {}
    const details = ['1K', '2K']
      .filter((tier) => hasPricingValue(counts[`output_${tier.toLowerCase()}`]))
      .join(', ')
    const freeFirst =
      counts.free_first_output === 'true' ? ` · ${t('First image free')}` : ''
    return details
      ? `${t('Per-image price')} · ${details}${freeFirst}`
      : `${t('Single per-image price')}${freeFirst}`
  }

  const inputPrice = ratioToPrice(row.ratio)
  if (!inputPrice) return t('No base input price')

  const details = [
    row.completionRatio &&
      `${t('Output')} $${ratioToPrice(row.completionRatio, inputPrice)}`,
    row.cacheRatio &&
      `${t('Cache')} $${ratioToPrice(row.cacheRatio, inputPrice)}`,
    row.createCacheRatio &&
      `${t('Cache write')} $${ratioToPrice(row.createCacheRatio, inputPrice)}`,
  ]
    .filter(Boolean)
    .slice(0, 2)

  return details.length > 0 ? details.join(' · ') : t('Base input price only')
}

export const buildModelSnapshots = ({
  modelPrice,
  modelRatio,
  cacheRatio,
  createCacheRatio,
  completionRatio,
  imageRatio,
  imageCompletionRatio,
  videoSecondPrice,
  imageCountPrice,
  audioRatio,
  audioCompletionRatio,
  billingMode,
  billingExpr,
}: ModelPricingSnapshotInput): ModelPricingSnapshot[] => {
  const priceMap = safeJsonParse<Record<string, number>>(modelPrice, {
    fallback: {},
    context: 'model prices',
  })
  const ratioMap = safeJsonParse<Record<string, number>>(modelRatio, {
    fallback: {},
    context: 'model ratios',
  })
  const cacheMap = safeJsonParse<Record<string, number>>(cacheRatio, {
    fallback: {},
    context: 'cache ratios',
  })
  const createCacheMap = safeJsonParse<Record<string, number>>(
    createCacheRatio,
    { fallback: {}, context: 'create cache ratios' }
  )
  const completionMap = safeJsonParse<Record<string, number>>(completionRatio, {
    fallback: {},
    context: 'completion ratios',
  })
  const imageMap = safeJsonParse<Record<string, number>>(imageRatio, {
    fallback: {},
    context: 'image ratios',
  })
  const imageCompletionMap = safeJsonParse<Record<string, number>>(
    imageCompletionRatio,
    { fallback: {}, context: 'image completion ratios' }
  )
  const videoSecondMap = safeJsonParse<Record<string, Record<string, number>>>(
    videoSecondPrice,
    { fallback: {}, context: 'video per-second prices' }
  )
  const imageCountMap = safeJsonParse<
    Record<string, Record<string, number | boolean>>
  >(imageCountPrice, { fallback: {}, context: 'image per-image prices' })
  const audioMap = safeJsonParse<Record<string, number>>(audioRatio, {
    fallback: {},
    context: 'audio ratios',
  })
  const audioCompletionMap = safeJsonParse<Record<string, number>>(
    audioCompletionRatio,
    { fallback: {}, context: 'audio completion ratios' }
  )
  const billingModeMap = safeJsonParse<Record<string, string>>(billingMode, {
    fallback: {},
    context: 'billing mode',
  })
  const billingExprMap = safeJsonParse<Record<string, string>>(billingExpr, {
    fallback: {},
    context: 'billing expression',
  })

  const modelNames = new Set([
    ...Object.keys(priceMap),
    ...Object.keys(ratioMap),
    ...Object.keys(cacheMap),
    ...Object.keys(createCacheMap),
    ...Object.keys(completionMap),
    ...Object.keys(imageMap),
    ...Object.keys(imageCompletionMap),
    ...Object.keys(videoSecondMap),
    ...Object.keys(imageCountMap),
    ...Object.keys(audioMap),
    ...Object.keys(audioCompletionMap),
    ...Object.keys(billingModeMap),
    ...Object.keys(billingExprMap),
  ])

  return Array.from(modelNames).map((name) => {
    const price = priceMap[name]?.toString() || ''
    const ratio = ratioMap[name]?.toString() || ''
    const cache = cacheMap[name]?.toString() || ''
    const createCache = createCacheMap[name]?.toString() || ''
    const completion = completionMap[name]?.toString() || ''
    const image = imageMap[name]?.toString() || ''
    const imageCompletion = imageCompletionMap[name]?.toString() || ''
    const secondTiers: Record<string, string> = {}
    Object.entries(videoSecondMap[name] ?? {}).forEach(([tier, value]) => {
      if (typeof value === 'number' && Number.isFinite(value)) {
        secondTiers[tier] = String(value)
      }
    })
    const hasVideoSecond = Object.keys(secondTiers).length > 0
    const imageCounts: Record<string, string> = {}
    Object.entries(imageCountMap[name] ?? {}).forEach(([key, value]) => {
      if (typeof value === 'number' && Number.isFinite(value)) {
        imageCounts[key] = String(value)
        return
      }
      // free_first_output 是布尔开关，其余键都是单价；
      // 统一成字符串存入快照，避免签名比较时 1 与 true 混用。
      if (typeof value === 'boolean') {
        imageCounts[key] = String(value)
      }
    })
    // 只有至少一个正价才算启用按张计费（后端的可用性判定同样要求正价）。
    const hasImageCount = IMAGE_COUNT_PRICE_KEYS.some(
      (key) => Number(imageCounts[key] ?? 0) > 0
    )
    const audio = audioMap[name]?.toString() || ''
    const audioCompletion = audioCompletionMap[name]?.toString() || ''

    const modeForModel = billingModeMap[name]
    if (modeForModel === 'tiered_expr') {
      const fullExpr = billingExprMap[name] || ''
      const { billingExpr: pureExpr, requestRuleExpr } =
        splitBillingExprAndRequestRules(fullExpr)
      return {
        name,
        billingMode: 'tiered_expr',
        billingExpr: pureExpr,
        requestRuleExpr,
        price,
        ratio,
        cacheRatio: cache,
        createCacheRatio: createCache,
        completionRatio: completion,
        imageRatio: image,
        imageCompletionRatio: imageCompletion,
        videoSecondPrice: hasVideoSecond ? secondTiers : undefined,
        imageCountPrice: hasImageCount ? imageCounts : undefined,
        audioRatio: audio,
        audioCompletionRatio: audioCompletion,
        hasConflict: false,
      }
    }

    // 顺序与后端 /api/pricing 的 quota_type 反推保持一致：
    // 按秒(2) > 按张(3) > 按次固定价(1) > 按 token(0)。
    let billingMode: PricingMode = 'per-token'
    if (hasVideoSecond) {
      billingMode = 'per-second'
    } else if (hasImageCount) {
      billingMode = 'per-image'
    } else if (price !== '') {
      billingMode = 'per-request'
    }

    return {
      name,
      price,
      ratio,
      cacheRatio: cache,
      createCacheRatio: createCache,
      completionRatio: completion,
      imageRatio: image,
      imageCompletionRatio: imageCompletion,
      videoSecondPrice: hasVideoSecond ? secondTiers : undefined,
      imageCountPrice: hasImageCount ? imageCounts : undefined,
      audioRatio: audio,
      audioCompletionRatio: audioCompletion,
      billingMode,
      hasConflict:
        (price !== '' || hasVideoSecond || hasImageCount) &&
        (ratio !== '' ||
          completion !== '' ||
          cache !== '' ||
          createCache !== '' ||
          image !== '' ||
          imageCompletion !== '' ||
          audio !== '' ||
          audioCompletion !== ''),
    }
  })
}

export const getSnapshotSignature = (snapshot?: ModelPricingSnapshot) => {
  if (!snapshot) return ''
  return JSON.stringify({
    price: snapshot.price || '',
    ratio: snapshot.ratio || '',
    cacheRatio: snapshot.cacheRatio || '',
    createCacheRatio: snapshot.createCacheRatio || '',
    completionRatio: snapshot.completionRatio || '',
    imageRatio: snapshot.imageRatio || '',
    imageCompletionRatio: snapshot.imageCompletionRatio || '',
    videoSecondPrice: snapshot.videoSecondPrice
      ? JSON.stringify(snapshot.videoSecondPrice)
      : '',
    imageCountPrice: snapshot.imageCountPrice
      ? JSON.stringify(snapshot.imageCountPrice)
      : '',
    audioRatio: snapshot.audioRatio || '',
    audioCompletionRatio: snapshot.audioCompletionRatio || '',
    billingMode: snapshot.billingMode || 'per-token',
    billingExpr: snapshot.billingExpr || '',
    requestRuleExpr: snapshot.requestRuleExpr || '',
  })
}
