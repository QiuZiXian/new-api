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
import { describe, expect, test } from 'vitest'

import {
  CHANNEL_TYPE_CII_SUBTITLE_ERASE,
  CHANNEL_TYPE_OPTIONS,
  MODEL_FETCHABLE_TYPES,
} from '../../constants'
import { CHANNEL_FORM_DEFAULT_VALUES, channelFormSchema } from '../channel-form'
import { getChannelTypeConfig } from '../channel-type-config'
import {
  getChannelTypeIcon,
  getChannelTypeLabel,
  getKeyPromptForType,
} from '../channel-utils'

function subtitleEraseForm(baseUrl: string) {
  return {
    ...CHANNEL_FORM_DEFAULT_VALUES,
    name: 'CII subtitle erase',
    type: CHANNEL_TYPE_CII_SUBTITLE_ERASE,
    base_url: baseUrl,
    key: 'test-key',
    models: 'cii-subtitle-erase',
  }
}

describe('CiiSubtitleErase channel', () => {
  test('registers the type for selection and keeps it in the video channel group', () => {
    const option = CHANNEL_TYPE_OPTIONS.find(
      (item) => item.value === CHANNEL_TYPE_CII_SUBTITLE_ERASE
    )

    expect(option).toEqual({
      value: CHANNEL_TYPE_CII_SUBTITLE_ERASE,
      label: 'CiiSubtitleErase',
    })
    expect(
      CHANNEL_TYPE_OPTIONS.findIndex(
        (item) => item.value === CHANNEL_TYPE_CII_SUBTITLE_ERASE
      )
    ).toBe(
      CHANNEL_TYPE_OPTIONS.findIndex((item) => item.value === 56) + 1
    )
  })

  test('resolves its display label instead of falling back to Unknown', () => {
    expect(getChannelTypeLabel(CHANNEL_TYPE_CII_SUBTITLE_ERASE)).toBe(
      'CiiSubtitleErase'
    )
  })

  test('uses its own logo and the CII gateway as default Base URL', () => {
    expect(getChannelTypeIcon(CHANNEL_TYPE_CII_SUBTITLE_ERASE)).toBe(
      'CiiSubtitleErase'
    )
    expect(getKeyPromptForType(CHANNEL_TYPE_CII_SUBTITLE_ERASE)).toBe(
      'Enter API key for this channel'
    )

    const config = getChannelTypeConfig(CHANNEL_TYPE_CII_SUBTITLE_ERASE)
    expect(config.icon).toBe('CiiSubtitleErase')
    expect(config.defaultBaseUrl).toBe('https://www.cii-group.com/app-api')
  })

  test('accepts a blank Base URL because the backend falls back to the CII gateway', () => {
    expect(channelFormSchema.safeParse(subtitleEraseForm('')).success).toBe(true)
    expect(
      channelFormSchema.safeParse(subtitleEraseForm('https://www.cii-group.com'))
        .success
    ).toBe(true)
  })

  test('does not advertise upstream model discovery', () => {
    expect(MODEL_FETCHABLE_TYPES.has(CHANNEL_TYPE_CII_SUBTITLE_ERASE)).toBe(
      false
    )
  })
})
