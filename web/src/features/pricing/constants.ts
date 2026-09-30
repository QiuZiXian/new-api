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
import { type TFunction } from 'i18next'

import type { TokenUnit } from './types'

// ----------------------------------------------------------------------------
// Pricing Constants
// ----------------------------------------------------------------------------

/** Sort options for pricing models */
export const SORT_OPTIONS = {
  NAME: 'name',
  PRICE_LOW: 'price-low',
  PRICE_HIGH: 'price-high',
} as const

export type SortOption = (typeof SORT_OPTIONS)[keyof typeof SORT_OPTIONS]

export function getSortLabels(t: TFunction): Record<SortOption, string> {
  return {
    [SORT_OPTIONS.NAME]: t('Name'),
    [SORT_OPTIONS.PRICE_LOW]: t('Price: Low to High'),
    [SORT_OPTIONS.PRICE_HIGH]: t('Price: High to Low'),
  }
}

/** Filter values */
export const FILTER_ALL = 'all'

/** Quota type options */
export const QUOTA_TYPES = {
  ALL: 'all',
  TOKEN: 'token',
  REQUEST: 'request',
} as const

export type QuotaTypeOption = (typeof QUOTA_TYPES)[keyof typeof QUOTA_TYPES]

/** Quota type labels */
export function getQuotaTypeLabels(
  t: TFunction
): Record<QuotaTypeOption, string> {
  return {
    [QUOTA_TYPES.ALL]: t('All Models'),
    [QUOTA_TYPES.TOKEN]: t('Token-based'),
    [QUOTA_TYPES.REQUEST]: t('Per Request'),
  }
}

/** Endpoint type options */
export const ENDPOINT_TYPES = {
  ALL: 'all',
  OPENAI: 'openai',
  OPENAI_RESPONSE: 'openai-response',
  ANTHROPIC: 'anthropic',
  GEMINI: 'gemini',
  JINA_RERANK: 'jina-rerank',
  IMAGE_GENERATION: 'image-generation',
  EMBEDDINGS: 'embeddings',
  OPENAI_VIDEO: 'openai-video',
} as const

export type EndpointTypeOption =
  (typeof ENDPOINT_TYPES)[keyof typeof ENDPOINT_TYPES]

/** Endpoint type labels */
export function getEndpointTypeLabels(
  t: TFunction
): Record<EndpointTypeOption, string> {
  return {
    [ENDPOINT_TYPES.ALL]: t('All Types'),
    [ENDPOINT_TYPES.OPENAI]: 'Chat',
    [ENDPOINT_TYPES.OPENAI_RESPONSE]: 'Response',
    [ENDPOINT_TYPES.ANTHROPIC]: 'Anthropic',
    [ENDPOINT_TYPES.GEMINI]: 'Gemini',
    [ENDPOINT_TYPES.JINA_RERANK]: 'Rerank',
    [ENDPOINT_TYPES.IMAGE_GENERATION]: t('Image'),
    [ENDPOINT_TYPES.EMBEDDINGS]: t('Embeddings'),
    [ENDPOINT_TYPES.OPENAI_VIDEO]: t('Video'),
  }
}

/** Filter section keys */
export const FILTER_SECTIONS = {
  PRICING_TYPE: 'pricingType',
  ENDPOINT_TYPE: 'endpointType',
  VENDOR: 'vendor',
  GROUP: 'group',
  TAG: 'tag',
} as const

/** Maximum number of tags to display in model row */
export const MAX_TAGS_DISPLAY = 5

/** Maximum number of filter items to display before showing "More..." */
export const MAX_FILTER_ITEMS = 5

/** Sidebar width */
export const SIDEBAR_WIDTH = 'w-64'

/** Excluded groups */
export const EXCLUDED_GROUPS = ['', 'auto']

/** Quota type values */
export const QUOTA_TYPE_VALUES = {
  TOKEN: 0,
  REQUEST: 1,
} as const

/** Token unit divisors */
export const TOKEN_UNIT_DIVISORS = {
  M: 1,
  K: 1000,
} as const

/** Default token unit for pricing display */
export const DEFAULT_TOKEN_UNIT: TokenUnit = 'M'

/** View mode options */
export const VIEW_MODES = {
  CARD: 'card',
  TABLE: 'table',
} as const

export type ViewMode = (typeof VIEW_MODES)[keyof typeof VIEW_MODES]

/** Default page size for pricing table */
export const DEFAULT_PRICING_PAGE_SIZE = 20

// ----------------------------------------------------------------------------
// Price display items
// ----------------------------------------------------------------------------

/** 分组语义键 → i18n 源串。后端只下发语义键，文案由前端翻译。 */
export const PRICE_GROUP_KEYS = {
  INPUT_TOKENS: 'input_tokens',
  OUTPUT_TOKENS: 'output_tokens',
  CACHED_INPUT: 'cached_input_tokens',
  CACHE_WRITE: 'cache_write_tokens',
  IMAGE_INPUT_TOKENS: 'image_input_tokens',
  AUDIO_INPUT: 'audio_input',
  AUDIO_OUTPUT: 'audio_output',
  IMAGE_OUTPUT: 'image_output',
  VIDEO_OUTPUT: 'video_output',
} as const

export type PriceGroupKey =
  (typeof PRICE_GROUP_KEYS)[keyof typeof PRICE_GROUP_KEYS]

/**
 * 分组语义键 → i18n 源串（「功能」列）。
 * 存英文源串是为了让多语言走既有 i18n 链路，不要在这里写死中文。
 */
export const PRICE_GROUP_LABELS: Record<PriceGroupKey, string> = {
  input_tokens: 'Input tokens',
  output_tokens: 'Output tokens',
  cached_input_tokens: 'Cached tokens',
  cache_write_tokens: 'Cache write tokens',
  image_input_tokens: 'Image input',
  audio_input: 'Audio input',
  audio_output: 'Audio output',
  image_output: 'Output',
  video_output: 'Output',
}

/** 默认维度文案（「维度」列），走已有 i18n 键 */
export const DEFAULT_PRICE_VARIANT_LABEL = 'Base Price'

/**
 * 分组语义键 → 计量绑定（billing_ref）。
 * 管理端新增展示行时按分组自动带出，避免管理员手填错绑定。
 */
export const PRICE_GROUP_BILLING_REF: Record<PriceGroupKey, string> = {
  input_tokens: 'input_tokens',
  output_tokens: 'output_tokens',
  cached_input_tokens: 'cached_input_tokens',
  cache_write_tokens: 'cache_write_tokens',
  image_input_tokens: 'input_tokens',
  audio_input: 'audio_input_tokens',
  audio_output: 'audio_output_tokens',
  image_output: 'image_output_count',
  video_output: 'video_output_second',
}

/** 展示单位 → i18n 源串（「单位」列） */
export const PRICE_UNIT_LABELS: Record<string, string> = {
  image: 'Unit image',
  second: 'Unit second',
  request: 'Unit request',
}

/** 表格列头 i18n 源串 */
export const PRICE_TABLE_COLUMNS = {
  FUNCTION: 'Function',
  VARIANT: 'Variant',
  PRICE: 'Price',
  UNIT: 'Unit',
  BILLING_ID: 'Billing ID',
} as const
