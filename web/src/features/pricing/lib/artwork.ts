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
import type { PricingModel } from '../types'

/**
 * 临时占位素材，来自 https://wenming7.cn/models 。
 * 换成自有 CDN 时只需修改本文件的 BASE_URL 与文件名，组件无需改动。
 * 图片加载失败时组件会回退到内置的渐变几何图。
 */
export const ARTWORK_BASE_URL = 'https://wenming7.cn'

const BANNER_PATH = '/sales/banner'

export const HERO_ARTWORK_URL = `${ARTWORK_BASE_URL}${BANNER_PATH}/ai-app-marketplace-developer-imagegen.webp`

const CARD_ARTWORK_URLS = [
  `${ARTWORK_BASE_URL}${BANNER_PATH}/banner-deepseek-v41-flash.webp`,
  `${ARTWORK_BASE_URL}${BANNER_PATH}/ai-app-marketplace-developer-imagegen.webp`,
  `${ARTWORK_BASE_URL}${BANNER_PATH}/banner-video.webp`,
  `${ARTWORK_BASE_URL}${BANNER_PATH}/banner-lumosonicx.webp`,
  `${ARTWORK_BASE_URL}${BANNER_PATH}/banner4.webp`,
] as const

const VIDEO_ARTWORK_URL = `${ARTWORK_BASE_URL}${BANNER_PATH}/banner-video.webp`
const IMAGE_ARTWORK_URL = `${ARTWORK_BASE_URL}${BANNER_PATH}/ai-app-marketplace-developer-imagegen.webp`

export function isVideoModel(model: PricingModel): boolean {
  return (
    /video|视频|seedance|kling|vidu|wanx/i.test(
      `${model.model_name} ${model.tags ?? ''}`
    ) || model.supported_endpoint_types?.includes('openai-video') === true
  )
}

/** Pick a stable artwork per model; capability-specific images win over the hash. */
export function getCardArtworkUrl(model: PricingModel): string {
  const text = `${model.model_name} ${model.tags ?? ''} ${model.category ?? ''}`.toLowerCase()
  if (isVideoModel(model)) return VIDEO_ARTWORK_URL
  if (/image|图像|绘图|flux|midjourney|dall|stable-diffusion|qwen-image/i.test(text)) {
    return IMAGE_ARTWORK_URL
  }
  let hash = 0
  for (let i = 0; i < model.model_name.length; i++) {
    hash = (hash * 31 + model.model_name.charCodeAt(i)) | 0
  }
  return CARD_ARTWORK_URLS[Math.abs(hash) % CARD_ARTWORK_URLS.length]
}
