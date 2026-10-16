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
import { Calculator } from 'lucide-react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group'
import { NativeSelect } from '@/components/ui/native-select'

import { numericDraftRegex } from './model-pricing-core'
import { formatPricingNumber } from './pricing-format'

// ARK_TOKEN_DIVISOR 是火山方舟 Seedance 的 token 换算除数：
// token 用量 = (输入视频时长 + 输出视频时长) × 输出宽 × 输出高 × 输出帧率 / 1024
const ARK_TOKEN_DIVISOR = 1024

const RESOLUTION_PRESETS = [
  { id: '480p', width: 864, height: 480 },
  { id: '720p', width: 1280, height: 720 },
  { id: '1080p', width: 1920, height: 1080 },
  { id: '4k', width: 3840, height: 2160 },
] as const

type DraftState = {
  preset: string
  unitPrice: string
  width: string
  height: string
  fps: string
  inputSeconds: string
  outputSeconds: string
  fxRate: string
}

const EMPTY_DRAFT: DraftState = {
  preset: '720p',
  unitPrice: '',
  width: '1280',
  height: '720',
  fps: '24',
  inputSeconds: '0',
  outputSeconds: '1',
  fxRate: '1',
}

// estimateTokensForSeconds 按官方公式把一段视频折算成 token 用量。
// 用户最常见的用法是 input=0 / output=1，得到「每输出一秒消耗多少 token」。
function estimateTokensForSeconds(
  width: number,
  height: number,
  fps: number,
  inputSeconds: number,
  outputSeconds: number
): number {
  const seconds = Math.max(inputSeconds, 0) + Math.max(outputSeconds, 0)
  return (seconds * width * height * fps) / ARK_TOKEN_DIVISOR
}

// estimatePricePerSecond 把「每百万 token 单价」折算成「每秒单价」。
// fxRate 为源 units 兑 1 USD 的汇率（填 1 表示同币种，不做换算）。
function estimatePricePerSecond(
  unitPricePerMillionTokens: number,
  tokensPerSecond: number,
  fxRate: number
): number {
  const costPerSecond = (tokensPerSecond / 1_000_000) * unitPricePerMillionTokens
  const rate = fxRate > 0 ? fxRate : 1
  return costPerSecond / rate
}

type NumberInputProps = {
  value: string
  placeholder?: string
  ariaLabel: string
  onChange: (value: string) => void
}

function DraftNumberInput({
  value,
  placeholder,
  ariaLabel,
  onChange,
}: NumberInputProps) {
  return (
    <Input
      inputMode='decimal'
      value={value}
      placeholder={placeholder}
      aria-label={ariaLabel}
      onChange={(event) => {
        const next = event.target.value
        if (numericDraftRegex.test(next)) onChange(next)
      }}
    />
  )
}

export function VideoSecondConverter(props: {
  onApply: (price: string) => void
}) {
  const { t } = useTranslation()
  const [draft, setDraft] = useState<DraftState>(EMPTY_DRAFT)

  const patchDraft = (patch: Partial<DraftState>) =>
    setDraft((current) => ({ ...current, ...patch }))

  const result = useMemo(() => {
    const unitPrice = Number(draft.unitPrice)
    const width = Number(draft.width)
    const height = Number(draft.height)
    const fps = Number(draft.fps)
    const inputSeconds = Number(draft.inputSeconds)
    const outputSeconds = Number(draft.outputSeconds)
    const fxRate = Number(draft.fxRate)

    if (
      !Number.isFinite(unitPrice) ||
      draft.unitPrice.trim() === '' ||
      !Number.isFinite(width) ||
      !Number.isFinite(height) ||
      !Number.isFinite(fps) ||
      width <= 0 ||
      height <= 0 ||
      fps <= 0 ||
      outputSeconds <= 0
    ) {
      return null
    }

    const tokensPerSecond = estimateTokensForSeconds(
      width,
      height,
      fps,
      Number.isFinite(inputSeconds) ? inputSeconds : 0,
      outputSeconds
    )
    const pricePerSecond = estimatePricePerSecond(
      unitPrice,
      tokensPerSecond,
      Number.isFinite(fxRate) ? fxRate : 1
    )
    return {
      tokensPerSecond,
      pricePerSecond,
      price: formatPricingNumber(pricePerSecond),
    }
  }, [draft])

  return (
    <Field className='bg-muted/20 rounded-lg border p-3'>
      <FieldGroup className='gap-3'>
        <div className='flex items-start gap-2'>
          <Calculator className='mt-0.5 size-4 shrink-0' />
          <div>
            <FieldLabel>{t('Token to per-second converter')}</FieldLabel>
            <FieldDescription>
              {t(
                'Convert an upstream token price using (input + output seconds) × width × height × fps / 1024.'
              )}
            </FieldDescription>
          </div>
        </div>

        <div className='grid gap-3 sm:grid-cols-2'>
          <Field>
            <FieldLabel>{t('Token price per 1M')}</FieldLabel>
            <InputGroup>
              <InputGroupAddon>$</InputGroupAddon>
              <InputGroupInput
                inputMode='decimal'
                value={draft.unitPrice}
                placeholder='15'
                aria-label={t('Token price per 1M')}
                onChange={(event) => {
                  const value = event.target.value
                  if (numericDraftRegex.test(value)) {
                    patchDraft({ unitPrice: value })
                  }
                }}
              />
              <InputGroupAddon align='inline-end'>/1M</InputGroupAddon>
            </InputGroup>
          </Field>

          <Field>
            <FieldLabel>{t('Resolution preset')}</FieldLabel>
            <NativeSelect
              aria-label={t('Resolution preset')}
              value={draft.preset}
              onChange={(event) => {
                const preset = RESOLUTION_PRESETS.find(
                  (item) => item.id === event.target.value
                )
                if (!preset) return
                patchDraft({
                  preset: preset.id,
                  width: String(preset.width),
                  height: String(preset.height),
                })
              }}
            >
              {RESOLUTION_PRESETS.map((preset) => (
                <option key={preset.id} value={preset.id}>
                  {preset.id} · {preset.width}×{preset.height}
                </option>
              ))}
            </NativeSelect>
          </Field>

          <Field>
            <FieldLabel>{t('Output width')}</FieldLabel>
            <DraftNumberInput
              value={draft.width}
              placeholder='1280'
              ariaLabel={t('Output width')}
              onChange={(value) => patchDraft({ width: value, preset: '' })}
            />
          </Field>

          <Field>
            <FieldLabel>{t('Output height')}</FieldLabel>
            <DraftNumberInput
              value={draft.height}
              placeholder='720'
              ariaLabel={t('Output height')}
              onChange={(value) => patchDraft({ height: value, preset: '' })}
            />
          </Field>

          <Field>
            <FieldLabel>{t('Frames per second')}</FieldLabel>
            <DraftNumberInput
              value={draft.fps}
              placeholder='24'
              ariaLabel={t('Frames per second')}
              onChange={(value) => patchDraft({ fps: value })}
            />
          </Field>

          <Field>
            <FieldLabel>{t('Input video seconds')}</FieldLabel>
            <DraftNumberInput
              value={draft.inputSeconds}
              placeholder='0'
              ariaLabel={t('Input video seconds')}
              onChange={(value) => patchDraft({ inputSeconds: value })}
            />
          </Field>

          <Field>
            <FieldLabel>{t('Output video seconds')}</FieldLabel>
            <DraftNumberInput
              value={draft.outputSeconds}
              placeholder='1'
              ariaLabel={t('Output video seconds')}
              onChange={(value) => patchDraft({ outputSeconds: value })}
            />
          </Field>

          <Field>
            <FieldLabel>{t('Exchange rate to USD')}</FieldLabel>
            <DraftNumberInput
              value={draft.fxRate}
              placeholder='1'
              ariaLabel={t('Exchange rate to USD')}
              onChange={(value) => patchDraft({ fxRate: value })}
            />
            <FieldDescription>
              {t('Ark prices are quoted per million tokens in CNY.')}
            </FieldDescription>
          </Field>
        </div>

        <div className='flex flex-wrap items-center justify-between gap-2'>
          <span className='text-sm'>
            {result ? (
              <>
                {formatPricingNumber(result.tokensPerSecond)} {t('tokens')} · $
                {result.price} / {t('second')}
              </>
            ) : (
              <span className='text-muted-foreground'>
                {t('Enter a token price to see the conversion.')}
              </span>
            )}
          </span>
          <Button
            type='button'
            variant='outline'
            size='sm'
            disabled={!result}
            onClick={() => result && props.onApply(result.price)}
          >
            {t('Apply to default tier')}
          </Button>
        </div>
      </FieldGroup>
    </Field>
  )
}
