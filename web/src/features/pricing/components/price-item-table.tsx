/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, version 3, of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useTranslation } from 'react-i18next'

import { formatCurrencyFromUSD } from '@/lib/currency'

import {
  DEFAULT_PRICE_VARIANT_LABEL,
  PRICE_GROUP_LABELS,
  PRICE_TABLE_COLUMNS,
  PRICE_UNIT_LABELS,
  TOKEN_UNIT_DIVISORS,
  type PriceGroupKey,
} from '../constants'
import type { PriceItem, PriceUnit, TokenUnit } from '../types'

// ----------------------------------------------------------------------------
// Price item table
//
// 模型广场详情页「价格信息」表的通用渲染器。四种排版（大语言模型 / 文生图 /
// OCR / 视频多档）共用同一份结构，差异全部由后端下发的 price_items 决定：
//   - 「功能」列：group_key 走 i18n，group_label 为管理员自定义文案
//   - 「维度」列：variant，视频多档时同一功能下挂多行
//   - 「单位」列：tokens 跟随 K/M 切换，其余显示张/秒/次
// 价格数值后端已按每单位算好，这里只做单位换算与货币格式化，不做价格计算。
// ----------------------------------------------------------------------------

function formatPriceValue(
  item: PriceItem,
  tokenUnit: TokenUnit,
  applyTokenDivisor: boolean
): string {
  const value = applyTokenDivisor
    ? item.price / TOKEN_UNIT_DIVISORS[tokenUnit]
    : item.price
  return formatCurrencyFromUSD(value, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

function formatOriginValue(
  origin: number,
  tokenUnit: TokenUnit,
  applyTokenDivisor: boolean
): string {
  const value = applyTokenDivisor
    ? origin / TOKEN_UNIT_DIVISORS[tokenUnit]
    : origin
  return formatCurrencyFromUSD(value, {
    digitsLarge: 4,
    digitsSmall: 6,
    abbreviate: false,
  })
}

function unitText(
  unit: PriceUnit,
  tokenUnit: TokenUnit,
  t: (key: string) => string
): string {
  if (unit === 'tokens') return `${tokenUnit} Tokens`
  return t(PRICE_UNIT_LABELS[unit] ?? unit)
}

// 计算每行的分组单元格跨度，以及该行是否被前一行的 rowSpan 覆盖。
// 「视频多档」这类排版下，同一功能挂多个维度行，功能列需要纵向合并。
function computeGroupLayout(items: PriceItem[]): {
  spans: number[]
  covered: boolean[]
} {
  const spans = items.map((item, index) => {
    let span = 1
    for (let next = index + 1; next < items.length; next += 1) {
      if (items[next].group_key !== item.group_key) break
      if (items[next].group_label !== item.group_label) break
      span += 1
    }
    return span
  })

  const covered = Array.from({ length: items.length }, () => false)
  let remaining = 0
  for (let i = 0; i < items.length; i += 1) {
    if (remaining > 0) {
      covered[i] = true
      remaining -= 1
      continue
    }
    remaining = spans[i] - 1
  }
  return { spans, covered }
}

export function PriceItemTable(props: {
  items: PriceItem[]
  tokenUnit: TokenUnit
}) {
  const { t } = useTranslation()
  const items = props.items
  const { spans, covered } = computeGroupLayout(items)

  return (
    <div className='overflow-hidden rounded-lg border'>
      <table className='w-full border-collapse text-sm'>
        <thead>
          <tr className='bg-muted/30 text-muted-foreground text-xs'>
            <th className='px-3 py-2 text-left font-medium'>
              {t(PRICE_TABLE_COLUMNS.FUNCTION)}
            </th>
            <th className='px-3 py-2 text-left font-medium'>
              {t(PRICE_TABLE_COLUMNS.VARIANT)}
            </th>
            <th className='px-3 py-2 text-right font-medium'>
              {t(PRICE_TABLE_COLUMNS.PRICE)}
            </th>
            <th className='px-3 py-2 text-left font-medium'>
              {t(PRICE_TABLE_COLUMNS.UNIT)}
            </th>
            <th className='px-3 py-2 text-left font-medium'>
              {t(PRICE_TABLE_COLUMNS.BILLING_ID)}
            </th>
          </tr>
        </thead>
        <tbody>
          {items.map((item, index) => {
            const isTokens = item.unit === 'tokens'
            const span = spans[index]
            // 被前一行的 rowSpan 覆盖时不重复渲染分组单元格
            const skipGroup = covered[index]
            const groupLabel = item.group_label
              ? item.group_label
              : t(PRICE_GROUP_LABELS[item.group_key as PriceGroupKey] ?? '')

            return (
              <tr key={item.id} className='border-t'>
                {!skipGroup && (
                  <td
                    className='text-foreground px-3 py-2.5 align-top'
                    rowSpan={span > 1 ? span : undefined}
                  >
                    {groupLabel}
                  </td>
                )}
                <td className='text-muted-foreground px-3 py-2.5 align-top'>
                  {item.variant || t(DEFAULT_PRICE_VARIANT_LABEL)}
                </td>
                <td className='px-3 py-2.5 text-right align-top'>
                  <div className='flex flex-wrap items-baseline justify-end gap-x-2 gap-y-0.5'>
                    <span className='text-foreground font-mono font-semibold tabular-nums'>
                      {formatPriceValue(item, props.tokenUnit, isTokens)}
                    </span>
                    {item.origin_price != null && item.origin_price > 0 && (
                      <span className='text-muted-foreground/60 font-mono text-xs tabular-nums line-through'>
                        {formatOriginValue(
                          item.origin_price,
                          props.tokenUnit,
                          isTokens
                        )}
                      </span>
                    )}
                  </div>
                  {(item.discount_label || item.saved_percent) && (
                    <div className='text-muted-foreground mt-0.5 text-xs'>
                      {item.discount_label}
                      {item.discount_label && item.saved_percent && (
                        <span className='mx-1'>·</span>
                      )}
                      {item.saved_percent &&
                        t('Save {{percent}}', {
                          percent: item.saved_percent,
                        })}
                    </div>
                  )}
                </td>
                <td className='text-muted-foreground px-3 py-2.5 align-top'>
                  {unitText(item.unit, props.tokenUnit, t)}
                </td>
                <td className='text-muted-foreground/70 px-3 py-2.5 align-top font-mono text-xs'>
                  {item.id}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
