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
import { zodResolver } from '@hookform/resolvers/zod'
import { AlertTriangle, Plus, Save, Trash2 } from 'lucide-react'
import {
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useState,
} from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { sideDrawerContentClassName } from '@/components/drawer-layout'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from '@/components/ui/field'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
} from '@/components/ui/input-group'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'

import {
  EMPTY_LANE_ENABLED,
  EMPTY_LANE_PRICES,
  fromImageCountPrice,
  fromSecondTierRows,
  buildPreviewRows,
  createInitialLaneState,
  createModelPricingSchema,
  hasValue,
  inputLaneByLane,
  laneConfigs,
  newSecondTierId,
  numericDraftRegex,
  ratioFieldByLane,
  resolvePricingMode,
  toImageCountPrice,
  toNumberOrNull,
  toSecondTierRows,
  VIDEO_SECOND_DEFAULT_TIER,
  type ImageCountPriceDraft,
  type LaneKey,
  type ModelPricingFormValues,
  type ModelRatioData,
  type PricingMode,
  type VideoSecondTierRow,
} from './model-pricing-core'
import { PriceInput, PriceLane } from './model-pricing-inputs'
import { formatPricingNumber } from './pricing-format'
import { TieredPricingEditor } from './tiered-pricing-editor'
import { VideoSecondConverter } from './video-second-converter'

export type { ModelRatioData } from './model-pricing-core'

type ModelPricingSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  editData?: ModelRatioData | null
  onSave?: () => void | Promise<void>
  isSaving?: boolean
}

type ModelPricingEditorPanelProps = Omit<
  ModelPricingSheetProps,
  'open' | 'onOpenChange'
> & {
  className?: string
}

export type ModelPricingEditorPanelHandle = {
  commitDraft: () => Promise<ModelRatioData | null>
}

export const ModelPricingSheet = forwardRef<
  ModelPricingEditorPanelHandle,
  ModelPricingSheetProps
>(function ModelPricingSheet(
  { open, onOpenChange, editData, onSave, isSaving },
  ref
) {
  const { t } = useTranslation()
  const title = editData ? t('Edit model pricing') : t('Add model pricing')
  const description = editData?.name || t('New model')

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side='right'
        className={sideDrawerContentClassName('sm:max-w-2xl')}
      >
        <SheetHeader className='sr-only'>
          <SheetTitle>{title}</SheetTitle>
          <SheetDescription>{description}</SheetDescription>
        </SheetHeader>
        <ModelPricingEditorPanel
          ref={ref}
          editData={editData}
          onSave={onSave}
          isSaving={isSaving}
          className='h-full rounded-none border-0'
        />
      </SheetContent>
    </Sheet>
  )
})

export const ModelPricingEditorPanel = forwardRef<
  ModelPricingEditorPanelHandle,
  ModelPricingEditorPanelProps
>(function ModelPricingEditorPanel(
  { editData, className, onSave, isSaving },
  ref
) {
  const { t } = useTranslation()
  const [pricingMode, setPricingMode] = useState<PricingMode>('per-token')
  const [promptPrice, setPromptPrice] = useState('')
  const [lanePrices, setLanePrices] = useState<Record<LaneKey, string>>({
    ...EMPTY_LANE_PRICES,
  })
  const [laneEnabled, setLaneEnabled] = useState<Record<LaneKey, boolean>>({
    ...EMPTY_LANE_ENABLED,
  })
  const [billingExpr, setBillingExpr] = useState('')
  const [requestRuleExpr, setRequestRuleExpr] = useState('')
  const [editorReloadToken, setEditorReloadToken] = useState(0)
  const [secondTiers, setSecondTiers] = useState<VideoSecondTierRow[]>(() =>
    toSecondTierRows(undefined)
  )
  const [imageCountPrice, setImageCountPrice] = useState<
    ReturnType<typeof toImageCountPrice>
  >(() => toImageCountPrice(undefined))
  const isEditMode = !!editData

  const form = useForm<ModelPricingFormValues>({
    resolver: zodResolver(createModelPricingSchema(t)),
    defaultValues: {
      name: '',
      price: '',
      ratio: '',
      cacheRatio: '',
      createCacheRatio: '',
      completionRatio: '',
      imageRatio: '',
      imageCompletionRatio: '',
      audioRatio: '',
      audioCompletionRatio: '',
    },
  })

  useEffect(() => {
    const nextLaneState = createInitialLaneState(editData)

    if (editData) {
      form.reset({
        name: editData.name,
        price: editData.price || '',
        ratio: editData.ratio || '',
        cacheRatio: editData.cacheRatio || '',
        createCacheRatio: editData.createCacheRatio || '',
        completionRatio: editData.completionRatio || '',
        imageRatio: editData.imageRatio || '',
        imageCompletionRatio: editData.imageCompletionRatio || '',
        audioRatio: editData.audioRatio || '',
        audioCompletionRatio: editData.audioCompletionRatio || '',
      })
      setPricingMode(resolvePricingMode(editData))
      setBillingExpr(editData.billingExpr || '')
      setRequestRuleExpr(editData.requestRuleExpr || '')
    } else {
      form.reset({
        name: '',
        price: '',
        ratio: '',
        cacheRatio: '',
        createCacheRatio: '',
        completionRatio: '',
        imageRatio: '',
        imageCompletionRatio: '',
        audioRatio: '',
        audioCompletionRatio: '',
      })
      setPricingMode('per-token')
      setBillingExpr('')
      setRequestRuleExpr('')
    }

    setPromptPrice(nextLaneState.promptPrice)
    setLanePrices(nextLaneState.prices)
    setLaneEnabled(nextLaneState.enabled)
    setSecondTiers(nextLaneState.secondTiers)
    setImageCountPrice(nextLaneState.imageCountPrice)
    setEditorReloadToken((token) => token + 1)
  }, [editData, form])

  const setFormValue = (field: keyof ModelPricingFormValues, value: string) => {
    form.setValue(field, value, {
      shouldDirty: true,
      shouldValidate: true,
    })
  }

  const deriveLaneRatio = (
    lane: LaneKey,
    price: string,
    nextPromptPrice = promptPrice,
    nextLanePrices = lanePrices
  ) => {
    const priceNumber = toNumberOrNull(price)
    if (priceNumber === null) return ''

    if (lane === 'audioOutput') {
      const audioInputPrice = toNumberOrNull(nextLanePrices.audioInput)
      if (audioInputPrice === null || audioInputPrice === 0) return ''
      return formatPricingNumber(priceNumber / audioInputPrice)
    }

    // 图片输出倍率相对图片输入价，而非提示词基准价
    if (lane === 'imageOutput') {
      const imageInputPrice = toNumberOrNull(nextLanePrices.image)
      if (imageInputPrice === null || imageInputPrice === 0) return ''
      return formatPricingNumber(priceNumber / imageInputPrice)
    }

    const inputPrice = toNumberOrNull(nextPromptPrice)
    if (inputPrice === null || inputPrice === 0) return ''
    return formatPricingNumber(priceNumber / inputPrice)
  }

  const syncLaneRatios = (
    nextPromptPrice = promptPrice,
    nextLanePrices = lanePrices,
    nextLaneEnabled = laneEnabled
  ) => {
    const inputPrice = toNumberOrNull(nextPromptPrice)
    setFormValue(
      'ratio',
      inputPrice !== null ? formatPricingNumber(inputPrice / 2) : ''
    )

    laneConfigs.forEach(({ key }) => {
      const ratioField = ratioFieldByLane[key]
      if (!nextLaneEnabled[key]) {
        setFormValue(ratioField, '')
        return
      }
      setFormValue(
        ratioField,
        deriveLaneRatio(
          key,
          nextLanePrices[key],
          nextPromptPrice,
          nextLanePrices
        )
      )
    })
  }

  const handlePromptPriceChange = (value: string) => {
    if (!numericDraftRegex.test(value)) return
    setPromptPrice(value)
    syncLaneRatios(value, lanePrices, laneEnabled)
  }

  const handleLanePriceChange = (lane: LaneKey, value: string) => {
    if (!numericDraftRegex.test(value)) return
    const nextLanePrices = { ...lanePrices, [lane]: value }
    setLanePrices(nextLanePrices)

    if (laneEnabled[lane]) {
      setFormValue(
        ratioFieldByLane[lane],
        deriveLaneRatio(lane, value, promptPrice, nextLanePrices)
      )
    }

    if (lane === 'audioInput' && laneEnabled.audioOutput) {
      setFormValue(
        'audioCompletionRatio',
        deriveLaneRatio(
          'audioOutput',
          nextLanePrices.audioOutput,
          promptPrice,
          nextLanePrices
        )
      )
    }

    if (lane === 'image' && laneEnabled.imageOutput) {
      setFormValue(
        'imageCompletionRatio',
        deriveLaneRatio(
          'imageOutput',
          nextLanePrices.imageOutput,
          promptPrice,
          nextLanePrices
        )
      )
    }
  }

  const handleLaneToggle = (lane: LaneKey, checked: boolean) => {
    const nextEnabled = { ...laneEnabled, [lane]: checked }
    let nextPrices = lanePrices

    if (!checked) {
      nextPrices = { ...nextPrices, [lane]: '' }
      setFormValue(ratioFieldByLane[lane], '')
      if (lane === 'audioInput') {
        nextEnabled.audioOutput = false
        nextPrices.audioOutput = ''
        setFormValue('audioCompletionRatio', '')
      }
      if (lane === 'image') {
        nextEnabled.imageOutput = false
        nextPrices.imageOutput = ''
        setFormValue('imageCompletionRatio', '')
      }
    }

    setLaneEnabled(nextEnabled)
    setLanePrices(nextPrices)

    if (checked) {
      setFormValue(
        ratioFieldByLane[lane],
        deriveLaneRatio(lane, nextPrices[lane], promptPrice, nextPrices)
      )
    }
  }

  const handleModeChange = (value: string) => {
    const nextMode = value as PricingMode
    setPricingMode(nextMode)
    if (nextMode === 'tiered_expr' && !billingExpr) {
      setBillingExpr('tier("base", p * 0 + c * 0)')
    }
    if (nextMode === 'per-second' && secondTiers.length === 0) {
      setSecondTiers(toSecondTierRows(undefined))
    }
  }

  const handleSecondTierChange = (
    index: number,
    patch: Partial<VideoSecondTierRow>
  ) => {
    setSecondTiers((rows) =>
      rows.map((row, i) => (i === index ? { ...row, ...patch } : row))
    )
  }

  const handleSecondTierAdd = () => {
    setSecondTiers((rows) => [
      ...rows,
      { id: newSecondTierId(), variant: '', price: '' },
    ])
  }

  const handleSecondTierRemove = (index: number) => {
    if (secondTiers[index]?.variant === VIDEO_SECOND_DEFAULT_TIER) return
    setSecondTiers((rows) => rows.filter((_, i) => i !== index))
  }

  const handleApplySecondPrice = (price: string) => {
    setSecondTiers((rows) => {
      const index = rows.findIndex(
        (row) =>
          row.variant.trim().toLowerCase() === VIDEO_SECOND_DEFAULT_TIER
      )
      if (index === -1) {
        return [
          {
            id: newSecondTierId(),
            variant: VIDEO_SECOND_DEFAULT_TIER,
            price,
          },
          ...rows,
        ]
      }
      const next = [...rows]
      next[index] = { ...next[index], price }
      return next
    })
  }

  const handleImageCountPriceChange = (
    key: keyof Omit<ImageCountPriceDraft, 'freeFirstOutput'>,
    value: string
  ) => {
    if (!numericDraftRegex.test(value)) return
    setImageCountPrice((prev) => ({ ...prev, [key]: value }))
  }

  const handleFreeFirstOutputChange = (checked: boolean) => {
    setImageCountPrice((prev) => ({ ...prev, freeFirstOutput: checked }))
  }

  const watchedValues = form.watch()
  const previewRows = useMemo(
    () =>
      buildPreviewRows(
        watchedValues,
        pricingMode,
        billingExpr,
        requestRuleExpr,
        promptPrice,
        lanePrices,
        laneEnabled,
        secondTiers,
        imageCountPrice,
        t
      ),
    [
      billingExpr,
      laneEnabled,
      lanePrices,
      pricingMode,
      promptPrice,
      requestRuleExpr,
      secondTiers,
      imageCountPrice,
      t,
      watchedValues,
    ]
  )

  const warnings = useMemo(() => {
    const nextWarnings: string[] = []
    const hasConflict =
      !!editData?.price &&
      [
        editData.ratio,
        editData.completionRatio,
        editData.cacheRatio,
        editData.createCacheRatio,
        editData.imageRatio,
        editData.imageCompletionRatio,
        editData.audioRatio,
        editData.audioCompletionRatio,
      ].some(hasValue)

    if (hasConflict) {
      nextWarnings.push(
        t(
          'This model has both fixed-price and token-price settings. Saving the current mode will rewrite the conflicting fields.'
        )
      )
    }

    if (
      pricingMode === 'per-token' &&
      toNumberOrNull(promptPrice) === null &&
      laneConfigs.some(
        ({ key }) => laneEnabled[key] && hasValue(lanePrices[key])
      )
    ) {
      nextWarnings.push(
        t('Input price is required before saving dependent prices.')
      )
    }

    if (
      pricingMode === 'per-token' &&
      laneEnabled.audioOutput &&
      !hasValue(lanePrices.audioInput)
    ) {
      nextWarnings.push(t('Audio output price requires an audio input price.'))
    }

    if (
      pricingMode === 'per-token' &&
      laneEnabled.imageOutput &&
      !hasValue(lanePrices.image)
    ) {
      nextWarnings.push(t('Image output price requires an image input price.'))
    }

    if (pricingMode === 'per-second') {
      // 只填了 720p/1080p 也是合法配置：提交时会自动用最低档价补 default，
      // 与后端 GetVideoSecondBasePrice 的兜底口径一致，不再阻断保存。
      if (!fromSecondTierRows(secondTiers)) {
        nextWarnings.push(t('Per-second billing needs at least one tier price.'))
      }
    }

    if (pricingMode === 'per-image' && !fromImageCountPrice(imageCountPrice)) {
      nextWarnings.push(
        t('Per-image billing needs at least one price before saving.')
      )
    }

    return nextWarnings
  }, [
    editData,
    imageCountPrice,
    laneEnabled,
    lanePrices,
    pricingMode,
    promptPrice,
    secondTiers,
    t,
  ])

  const validatePricingValues = useCallback(() => {
    if (
      pricingMode === 'per-token' &&
      toNumberOrNull(promptPrice) === null &&
      laneConfigs.some(
        ({ key }) => laneEnabled[key] && hasValue(lanePrices[key])
      )
    ) {
      form.setError('ratio', {
        message: t('Input price is required before saving dependent prices.'),
      })
      return false
    }

    if (
      pricingMode === 'per-token' &&
      laneEnabled.audioOutput &&
      !hasValue(lanePrices.audioInput)
    ) {
      form.setError('audioRatio', {
        message: t('Audio output price requires an audio input price.'),
      })
      return false
    }

    if (
      pricingMode === 'per-token' &&
      laneEnabled.imageOutput &&
      !hasValue(lanePrices.image)
    ) {
      form.setError('imageRatio', {
        message: t('Image output price requires an image input price.'),
      })
      return false
    }

    if (pricingMode === 'per-second') {
      // 缺少 default 档不再阻断：保存时自动补最低档价，否则管理员只填了
      // 720p 会一直保存失败，表现为「配了按秒却仍提示先配置模型价格」。
      if (!fromSecondTierRows(secondTiers)) {
        form.setError('name', {
          message: t('Per-second billing needs at least one tier price.'),
        })
        return false
      }
    }

    if (pricingMode === 'per-image' && !fromImageCountPrice(imageCountPrice)) {
      form.setError('name', {
        message: t('Per-image billing needs at least one price before saving.'),
      })
      return false
    }

    return true
  }, [
    form,
    imageCountPrice,
    laneEnabled,
    lanePrices,
    pricingMode,
    promptPrice,
    secondTiers,
    t,
  ])

  const buildSubmitData = useCallback(
    (values: ModelPricingFormValues) => {
      const data: ModelRatioData = {
        name: values.name.trim(),
        billingMode: pricingMode,
        price: values.price || '',
        ratio: values.ratio || '',
        cacheRatio: values.cacheRatio || '',
        createCacheRatio: values.createCacheRatio || '',
        completionRatio: values.completionRatio || '',
        imageRatio: values.imageRatio || '',
        imageCompletionRatio: values.imageCompletionRatio || '',
        audioRatio: values.audioRatio || '',
        audioCompletionRatio: values.audioCompletionRatio || '',
      }

      if (pricingMode === 'per-second') {
        const tiers = fromSecondTierRows(secondTiers)
        if (tiers && tiers[VIDEO_SECOND_DEFAULT_TIER] === undefined) {
          const positivePrices = Object.values(tiers)
            .map((value) => Number(value))
            .filter((value) => Number.isFinite(value) && value > 0)
          // default 档用于请求未指定分辨率时计价；后端会以最低档价兜底，
          // 这里显式补齐，保证前后端口径完全一致。
          if (positivePrices.length > 0) {
            tiers[VIDEO_SECOND_DEFAULT_TIER] = String(
              Math.min(...positivePrices)
            )
          }
        }
        data.videoSecondPrice = tiers
      }

      if (pricingMode === 'per-image') {
        data.imageCountPrice = fromImageCountPrice(imageCountPrice)
      }

      if (pricingMode === 'tiered_expr') {
        data.billingExpr = billingExpr
        data.requestRuleExpr = requestRuleExpr
      }

      return data
    },
    [billingExpr, imageCountPrice, pricingMode, requestRuleExpr, secondTiers]
  )

  useImperativeHandle(
    ref,
    () => ({
      commitDraft: async () => {
        const isValid = await form.trigger()
        if (!isValid || !validatePricingValues()) return null
        return buildSubmitData(form.getValues())
      },
    }),
    [form, validatePricingValues, buildSubmitData]
  )

  const showActions = Boolean(onSave)

  return (
    <div
      className={cn(
        'bg-background flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl border',
        className
      )}
    >
      <div className='border-b p-4'>
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div className='min-w-0'>
            <h3 className='truncate text-base font-medium'>
              {isEditMode ? t('Edit model pricing') : t('Add model pricing')}
            </h3>
          </div>
        </div>
      </div>

      <Form {...form}>
        <form
          onSubmit={(event) => event.preventDefault()}
          className='flex min-h-0 flex-1 flex-col'
          autoComplete='off'
        >
          <div className='min-h-0 flex-1 overflow-y-auto p-4 pb-6'>
            <div className='grid items-start gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(220px,260px)]'>
              <FieldGroup>
                {warnings.length > 0 && (
                  <Alert variant='destructive'>
                    <AlertTriangle data-icon='inline-start' />
                    <AlertDescription>
                      <div className='flex flex-col gap-1'>
                        {warnings.map((warning) => (
                          <span key={warning}>{warning}</span>
                        ))}
                      </div>
                    </AlertDescription>
                  </Alert>
                )}

                <FormField
                  control={form.control}
                  name='name'
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t('Model name')}</FormLabel>
                      <FormControl>
                        <Input
                          placeholder={t('gpt-4')}
                          {...field}
                          disabled={isEditMode}
                        />
                      </FormControl>
                      <FormDescription>
                        {t(
                          'The exact model identifier as used in API requests.'
                        )}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Tabs
                  value={pricingMode}
                  onValueChange={handleModeChange}
                  className='gap-4'
                >
                  <TabsList className='grid w-full grid-cols-5'>
                    <TabsTrigger value='per-token'>
                      {t('Per-token')}
                    </TabsTrigger>
                    <TabsTrigger value='per-request'>
                      {t('Per-request')}
                    </TabsTrigger>
                    <TabsTrigger value='per-second'>
                      {t('Per-second')}
                    </TabsTrigger>
                    <TabsTrigger value='per-image'>
                      {t('Per-image')}
                    </TabsTrigger>
                    <TabsTrigger value='tiered_expr'>
                      {t('Expression')}
                    </TabsTrigger>
                  </TabsList>

                  <TabsContent value='per-token' className='pt-0'>
                    <FieldGroup className='gap-5'>
                      <Field>
                        <FieldLabel>{t('Input price')}</FieldLabel>
                        <PriceInput
                          value={promptPrice}
                          placeholder='3'
                          onChange={handlePromptPriceChange}
                        />
                        <FieldDescription>
                          {t('USD price per 1M input tokens.')}
                        </FieldDescription>
                      </Field>

                      <div className='grid gap-3 sm:grid-cols-[repeat(auto-fit,minmax(400px,1fr))]'>
                        {laneConfigs.map((lane) => {
                          const inputLane = inputLaneByLane[lane.key]
                          const disabled =
                            inputLane !== undefined &&
                            (!laneEnabled[inputLane] ||
                              !hasValue(lanePrices[inputLane]))
                          return (
                            <PriceLane
                              key={lane.key}
                              title={t(lane.titleKey)}
                              description={t(lane.descriptionKey)}
                              placeholder={lane.placeholder}
                              value={lanePrices[lane.key]}
                              enabled={laneEnabled[lane.key]}
                              disabled={disabled}
                              onEnabledChange={(checked) =>
                                handleLaneToggle(lane.key, checked)
                              }
                              onChange={(value) =>
                                handleLanePriceChange(lane.key, value)
                              }
                            />
                          )
                        })}
                      </div>
                    </FieldGroup>
                  </TabsContent>

                  <TabsContent value='per-request' className='pt-0'>
                    <FieldGroup className='gap-5'>
                      <FormField
                        control={form.control}
                        name='price'
                        render={({ field }) => (
                          <FormItem className='contents'>
                            <Field>
                              <FieldLabel>{t('Fixed price')}</FieldLabel>
                              <FormControl>
                                <InputGroup>
                                  <InputGroupAddon>$</InputGroupAddon>
                                  <InputGroupInput
                                    inputMode='decimal'
                                    placeholder='0.01'
                                    {...field}
                                    onChange={(event) => {
                                      const value = event.target.value
                                      if (numericDraftRegex.test(value)) {
                                        field.onChange(value)
                                      }
                                    }}
                                  />
                                  <InputGroupAddon align='inline-end'>
                                    {t('per request')}
                                  </InputGroupAddon>
                                </InputGroup>
                              </FormControl>
                              <FieldDescription>
                                {t(
                                  'Cost in USD per request, regardless of tokens used.'
                                )}
                              </FieldDescription>
                              <FormMessage />
                            </Field>
                          </FormItem>
                        )}
                      />
                    </FieldGroup>
                  </TabsContent>

                  <TabsContent value='per-image' className='pt-0'>
                    <FieldGroup className='gap-5'>
                      <Field>
                        <FieldLabel>{t('Image per-image price')}</FieldLabel>
                        <FieldDescription>
                          {t(
                            'USD price per image. Input counts reference images, output counts successfully generated images. The two are added together, not multiplied.'
                          )}
                        </FieldDescription>
                        <FieldDescription>
                          {t(
                            'Tiers follow the output resolution area: 1K is up to 2,250,000 pixels, anything above is 2K. Leave a tier empty to fall back to the default price.'
                          )}
                        </FieldDescription>
                      </Field>

                      <div className='flex flex-col gap-2'>
                        <div className='grid grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1fr)] items-end gap-2'>
                          <span className='text-muted-foreground text-xs'>
                            {t('Resolution tier')}
                          </span>
                          <span className='text-muted-foreground text-xs'>
                            {t('Input price per image')}
                          </span>
                          <span className='text-muted-foreground text-xs'>
                            {t('Output price per image')}
                          </span>
                        </div>
                        {(
                          [
                            {
                              key: 'default',
                              label: t('Default tier'),
                              hint: '0.02',
                              inputKey: 'input',
                              outputKey: 'output',
                            },
                            {
                              key: '1k',
                              label: '1K',
                              hint: '0.02',
                              inputKey: 'input1k',
                              outputKey: 'output1k',
                            },
                            {
                              key: '2k',
                              label: '2K',
                              hint: '0.18',
                              inputKey: 'input2k',
                              outputKey: 'output2k',
                            },
                          ] as const
                        ).map((row) => (
                          <div
                            key={row.key}
                            className='grid grid-cols-[minmax(0,1.1fr)_minmax(0,1fr)_minmax(0,1fr)] items-center gap-2'
                          >
                            <span className='text-sm'>{row.label}</span>
                            <InputGroup>
                              <InputGroupAddon>$</InputGroupAddon>
                              <InputGroupInput
                                inputMode='decimal'
                                placeholder={row.hint}
                                value={imageCountPrice[row.inputKey]}
                                onChange={(event) =>
                                  handleImageCountPriceChange(
                                    row.inputKey,
                                    event.target.value
                                  )
                                }
                              />
                            </InputGroup>
                            <InputGroup>
                              <InputGroupAddon>$</InputGroupAddon>
                              <InputGroupInput
                                inputMode='decimal'
                                placeholder={row.hint}
                                value={imageCountPrice[row.outputKey]}
                                onChange={(event) =>
                                  handleImageCountPriceChange(
                                    row.outputKey,
                                    event.target.value
                                  )
                                }
                              />
                            </InputGroup>
                          </div>
                        ))}
                        <p className='text-muted-foreground text-xs'>
                          {t(
                            'Each cell is a price per image. Empty tier cells reuse the default price.'
                          )}
                        </p>
                      </div>

                      <Field orientation='horizontal'>
                        <FieldContent>
                          <FieldLabel htmlFor='image-free-first-output'>
                            {t('First output image free')}
                          </FieldLabel>
                          <FieldDescription>
                            {t(
                              'When enabled, the first generated image of every request is not billed.'
                            )}
                          </FieldDescription>
                        </FieldContent>
                        <Switch
                          id='image-free-first-output'
                          checked={imageCountPrice.freeFirstOutput}
                          onCheckedChange={handleFreeFirstOutputChange}
                        />
                      </Field>
                    </FieldGroup>
                  </TabsContent>

                  <TabsContent value='per-second' className='pt-0'>
                    <FieldGroup className='gap-5'>
                      <Field>
                        <FieldLabel>{t('Per-second tiers')}</FieldLabel>
                        <FieldDescription>
                          {t(
                            'USD price per second of generated video, one row per resolution tier. Videos are charged for their real duration once the task finishes.'
                          )}
                        </FieldDescription>
                      </Field>
                      <VideoSecondConverter onApply={handleApplySecondPrice} />

                      <div className='flex flex-col gap-2'>
                        <div className='grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-end gap-2'>
                          <span className='text-muted-foreground text-xs'>
                            {t('Resolution tier')}
                          </span>
                          <span className='text-muted-foreground text-xs'>
                            {t('Price per second')}
                          </span>
                          <span className='sr-only'>{t('Actions')}</span>
                        </div>
                        {secondTiers.map((tier, index) => {
                          const isDefault =
                            tier.variant.trim().toLowerCase() ===
                            VIDEO_SECOND_DEFAULT_TIER
                          const missingDefault =
                            pricingMode === 'per-second' &&
                            isDefault &&
                            tier.price.trim() === ''
                          return (
                            <div key={tier.id}>
                              <div className='grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto] items-center gap-2'>
                                <Input
                                  value={tier.variant}
                                  placeholder={isDefault ? 'default' : '1080p'}
                                  readOnly={isDefault}
                                  aria-label={t('Resolution tier')}
                                  onChange={(event) =>
                                    handleSecondTierChange(index, {
                                      variant: event.target.value,
                                    })
                                  }
                                />
                                <InputGroup>
                                  <InputGroupAddon>$</InputGroupAddon>
                                  <InputGroupInput
                                    inputMode='decimal'
                                    value={tier.price}
                                    placeholder='0.99'
                                    aria-label={t('Price per second')}
                                    onChange={(event) => {
                                      const value = event.target.value
                                      if (numericDraftRegex.test(value)) {
                                        handleSecondTierChange(index, {
                                          price: value,
                                        })
                                      }
                                    }}
                                  />
                                  <InputGroupAddon align='inline-end'>
                                    {t('per second')}
                                  </InputGroupAddon>
                                </InputGroup>
                                <Button
                                  type='button'
                                  variant='ghost'
                                  size='icon'
                                  aria-label={t('Remove tier')}
                                  disabled={isDefault}
                                  onClick={() => handleSecondTierRemove(index)}
                                >
                                  <Trash2 />
                                </Button>
                              </div>
                              {missingDefault && (
                                <p className='text-destructive mt-1 text-xs'>
                                  {t(
                                    'A default tier price is required: it bills videos that omit a resolution.'
                                  )}
                                </p>
                              )}
                            </div>
                          )
                        })}
                      </div>
                      <Button
                        type='button'
                        variant='outline'
                        size='sm'
                        className='w-full sm:w-auto'
                        onClick={handleSecondTierAdd}
                      >
                        <Plus data-icon='inline-start' />
                        {t('Add tier')}
                      </Button>
                    </FieldGroup>
                  </TabsContent>

                  <TabsContent value='tiered_expr' className='pt-0'>
                    <FieldGroup className='gap-5'>
                      <TieredPricingEditor
                        key={editorReloadToken}
                        modelName={watchedValues.name}
                        billingExpr={billingExpr}
                        requestRuleExpr={requestRuleExpr}
                        onBillingExprChange={setBillingExpr}
                        onRequestRuleExprChange={setRequestRuleExpr}
                      />
                    </FieldGroup>
                  </TabsContent>
                </Tabs>
              </FieldGroup>

              <aside className='bg-muted/20 sticky top-0 rounded-lg border'>
                <div className='border-b px-3 py-2'>
                  <div className='text-sm font-medium'>{t('Preview')}</div>
                </div>
                <div className='divide-y'>
                  {previewRows.map((row) => (
                    <div key={row.key} className='grid gap-1 px-3 py-2.5'>
                      <span className='text-muted-foreground text-xs'>
                        {row.label}
                      </span>
                      <span
                        className={cn(
                          'min-w-0 text-sm',
                          row.multiline
                            ? 'font-mono text-xs leading-5 break-words whitespace-pre-wrap'
                            : 'truncate'
                        )}
                      >
                        {row.value}
                      </span>
                    </div>
                  ))}
                </div>
              </aside>
            </div>
          </div>
          {showActions && (
            <div className='bg-background/95 supports-[backdrop-filter]:bg-background/80 shrink-0 border-t p-3 backdrop-blur'>
              <div className='flex flex-col-reverse gap-2 sm:flex-row sm:justify-end'>
                {onSave && (
                  <Button
                    type='button'
                    onClick={onSave}
                    disabled={isSaving}
                    className='w-full sm:w-auto'
                  >
                    <Save data-icon='inline-start' />
                    {isSaving ? t('Saving...') : t('Save model prices')}
                  </Button>
                )}
              </div>
            </div>
          )}
        </form>
      </Form>
    </div>
  )
})
