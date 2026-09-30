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
import { Plus, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  PRICE_GROUP_BILLING_REF,
  PRICE_GROUP_KEYS,
  PRICE_GROUP_LABELS,
  type PriceGroupKey,
} from '@/features/pricing/constants'
import { cn } from '@/lib/utils'

import {
  PRICE_ITEM_UNITS,
  emptyPriceItemDraft,
  type PriceItemDraft,
} from '../lib/price-items'

// ----------------------------------------------------------------------------
// Price display items editor
//
// 只编辑「怎么展示」：分组、维度、单位、原价、启用。价格数值不在这里录入，
// 由后端根据 billing_ref 从既有计量配置实时推导——计量配置一改，展示价自动跟着变。
// ----------------------------------------------------------------------------

export function PriceItemsEditor(props: {
  value: PriceItemDraft[]
  onChange: (next: PriceItemDraft[]) => void
  /** 用于生成默认计费标识 */
  modelName: string
}) {
  const { t } = useTranslation()
  const { value, onChange } = props

  const patch = (index: number, next: Partial<PriceItemDraft>) => {
    onChange(value.map((item, i) => (i === index ? { ...item, ...next } : item)))
  }

  const addRow = () => {
    const suffix = value.length + 1
    onChange([
      ...value,
      emptyPriceItemDraft(`${props.modelName || 'model'}.item-${suffix}`),
    ])
  }

  const removeRow = (uid: string) => {
    onChange(value.filter((item) => item.uid !== uid))
  }

  return (
    <div className='space-y-3'>
      {value.length === 0 && (
        <p className='text-muted-foreground text-xs'>
          {t(
            'Leave empty to derive standard input / output / cache rows from the billing configuration'
          )}
        </p>
      )}

      {value.map((item, index) => (
        <div
          key={item.uid}
          className={cn(
            'rounded-lg border p-3',
            !item.enabled && 'opacity-60'
          )}
        >
          <div className='flex items-center justify-between gap-2'>
            <Input
              className='font-mono text-xs'
              value={item.id}
              placeholder={t('Billing ID')}
              onChange={(e) => patch(index, { id: e.target.value })}
            />
            <div className='flex shrink-0 items-center gap-2'>
              <Switch
                checked={item.enabled}
                onCheckedChange={(checked) =>
                  patch(index, { enabled: checked })
                }
                aria-label={t('Enabled')}
              />
              <Button
                type='button'
                variant='ghost'
                size='icon'
                onClick={() => removeRow(item.uid)}
                aria-label={t('Delete')}
              >
                <Trash2 className='h-4 w-4' />
              </Button>
            </div>
          </div>

          <div className='mt-2 grid gap-2 sm:grid-cols-2'>
            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>
                {t('Function')}
              </label>
              <Select
                value={item.group_key || 'custom'}
                onValueChange={(next) => {
                  if (next === 'custom') {
                    patch(index, { group_key: '' })
                    return
                  }
                  if (!next) return
                  patch(index, {
                    group_key: next,
                    group_label: '',
                    billing_ref:
                      PRICE_GROUP_BILLING_REF[next as PriceGroupKey] ??
                      item.billing_ref,
                  })
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.values(PRICE_GROUP_KEYS).map((key) => (
                    <SelectItem key={key} value={key}>
                      {t(PRICE_GROUP_LABELS[key])}
                    </SelectItem>
                  ))}
                  <SelectItem value='custom'>
                    {t('Custom label')}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>
                {t('Variant')}
              </label>
              <Input
                value={item.variant}
                placeholder={t('e.g. Fast 480p')}
                onChange={(e) => patch(index, { variant: e.target.value })}
              />
            </div>

            {!item.group_key && (
              <div className='space-y-1'>
                <label className='text-muted-foreground text-xs'>
                  {t('Custom label')}
                </label>
                <Input
                  value={item.group_label}
                  onChange={(e) =>
                    patch(index, { group_label: e.target.value })
                  }
                />
              </div>
            )}

            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>
                {t('Unit')}
              </label>
              <Select
                value={item.unit}
                onValueChange={(next) => {
                  if (!next) return
                  patch(index, { unit: next })
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {PRICE_ITEM_UNITS.map((unit) => (
                    <SelectItem key={unit} value={unit}>
                      {unit}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>
                {t('Origin price')}
              </label>
              <Input
                value={item.origin_price}
                inputMode='decimal'
                placeholder={t('Optional, enables the discount badge')}
                onChange={(e) =>
                  patch(index, { origin_price: e.target.value })
                }
              />
            </div>
          </div>
        </div>
      ))}

      <Button type='button' variant='outline' size='sm' onClick={addRow}>
        <Plus className='mr-1 h-4 w-4' />
        {t('Add price item')}
      </Button>
    </div>
  )
}
