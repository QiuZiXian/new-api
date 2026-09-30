/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import {
  PRICE_GROUP_BILLING_REF,
  PRICE_GROUP_KEYS,
} from '@/features/pricing/constants'

// ----------------------------------------------------------------------------
// Price display items (models.price_items)
//
// 只描述「模型广场怎么展示这一行」：分组、维度、单位、原价、启用。
// 价格数值不在这里录入，由后端按 billing_ref 从既有计量配置实时推导——
// 计量配置一改展示价自动跟着变，不会出现两处配置漂移。
// ----------------------------------------------------------------------------

export type PriceItemDraft = {
  /** 列表 key，纯本地，不落库 */
  uid: string
  id: string
  group_key: string
  group_label: string
  variant: string
  billing_ref: string
  unit: string
  origin_price: string
  enabled: boolean
}

const UNIT_OPTIONS = ['tokens', 'image', 'second', 'request'] as const

export const PRICE_ITEM_UNITS: readonly string[] = UNIT_OPTIONS

function newUid(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) {
    return crypto.randomUUID()
  }
  return `pi-${Math.random().toString(36).slice(2)}`
}

export function emptyPriceItemDraft(id: string): PriceItemDraft {
  return {
    uid: newUid(),
    id,
    group_key: PRICE_GROUP_KEYS.OUTPUT_TOKENS,
    group_label: '',
    variant: '',
    billing_ref: PRICE_GROUP_BILLING_REF[PRICE_GROUP_KEYS.OUTPUT_TOKENS],
    unit: 'tokens',
    origin_price: '',
    enabled: true,
  }
}

/** 后端存储值（models.price_items）→ 表单草稿 */
export function parsePriceItems(raw: string | undefined): PriceItemDraft[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed.map((item: Record<string, unknown>) => ({
      uid: newUid(),
      id: String(item.id ?? ''),
      group_key: String(item.group_key ?? ''),
      group_label: String(item.group_label ?? ''),
      variant: String(item.variant ?? ''),
      billing_ref: String(item.billing_ref ?? ''),
      unit: String(item.unit ?? 'tokens'),
      origin_price:
        item.origin_price === undefined || item.origin_price === null
          ? ''
          : String(item.origin_price),
      enabled: item.enabled !== false,
    }))
  } catch {
    return []
  }
}

/** 表单草稿 → 后端存储值。id 为空的行直接丢弃。 */
export function serializePriceItems(drafts: PriceItemDraft[]): string {
  const items = drafts
    .filter((d) => d.id.trim() !== '')
    .map((d) => {
      const origin = Number.parseFloat(d.origin_price)
      const item: Record<string, unknown> = {
        id: d.id.trim(),
        billing_ref: d.billing_ref,
        unit: d.unit,
        enabled: d.enabled,
      }
      if (d.group_key) item.group_key = d.group_key
      if (d.group_label.trim()) item.group_label = d.group_label.trim()
      if (d.variant.trim()) item.variant = d.variant.trim()
      if (Number.isFinite(origin) && origin > 0) item.origin_price = origin
      return item
    })
  return JSON.stringify(items)
}
