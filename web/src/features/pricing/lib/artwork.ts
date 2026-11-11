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
 * 临时占位素材，取自线上 https://wenming7.cn/models 的 ModelsPage bundle。
 * 路径与选图规则都按该页面原样复刻（/sales/images/models/*.webp）。
 * 换成自有 CDN 时只改 ARTWORK_BASE_URL 与下面的文件名，组件无需改动；
 * 图片加载失败时组件会回退到内置渐变几何图 / 首字母。
 */
export const ARTWORK_BASE_URL = 'https://wenming7.cn'

const MODEL_IMAGE_PATH = '/sales/images/models'

function modelImage(name: string): string {
  return `${ARTWORK_BASE_URL}${MODEL_IMAGE_PATH}/${name}.webp`
}

/** Hero: AI chip close-up layered over the landscape. */
export const HERO_CHIP_ARTWORK_URL = modelImage('provider-hero-chip-hd')
export const HERO_LANDSCAPE_ARTWORK_URL = modelImage('provider-landscape-hd')

/** Card artwork, mirroring the upstream category mapping. */
const CARD_ARTWORKS = {
  video: modelImage('provider-landscape-hd'),
  image: modelImage('provider-image-hd'),
  orbit: modelImage('provider-art-orbit-hd'),
  cube: modelImage('provider-art-cube-hd'),
} as const

/** Provider logos shipped by the same page, used when a vendor has no icon set. */
const PROVIDER_LOGO_FILES: Record<string, string> = {
  vidu: 'provider-vidu.webp',
  'glm-5.1': 'provider-glm-5.1.svg',
  'deepseek-ai': 'provider-deepseek-ai.svg',
  deepseek: 'provider-deepseek-ai.svg',
  minimax: 'provider-minimax.svg',
  qwen: 'provider-qwen.svg',
  wan: 'provider-wan.svg',
  moonshotai: 'provider-moonshotai.webp',
  moonshot: 'provider-moonshotai.webp',
  xiaomi: 'provider-xiaomi.webp',
  'bytedance-seed': 'provider-bytedance-seed.svg',
  wenming: 'provider-wenming.webp',
  // Common Chinese vendor names used by local metadata.
  智谱: 'provider-glm-5.1.svg',
  月之暗面: 'provider-moonshotai.webp',
  字节跳动: 'provider-bytedance-seed.svg',
  通义千问: 'provider-qwen.svg',
  小米: 'provider-xiaomi.webp',
}

export function getProviderLogoUrl(name?: string): string | undefined {
  if (!name) return undefined
  const key = name.trim().toLowerCase()
  const file = PROVIDER_LOGO_FILES[key]
  if (file) return `${ARTWORK_BASE_URL}${MODEL_IMAGE_PATH}/${file}`
  const partial = Object.keys(PROVIDER_LOGO_FILES).find((candidate) =>
    key.includes(candidate)
  )
  return partial
    ? `${ARTWORK_BASE_URL}${MODEL_IMAGE_PATH}/${PROVIDER_LOGO_FILES[partial]}`
    : undefined
}

export function isVideoModel(model: PricingModel): boolean {
  return (
    /video|视频|seedance|kling|vidu|wanx/i.test(
      `${model.model_name} ${model.tags ?? ''} ${model.category ?? ''}`
    ) || model.supported_endpoint_types?.includes('openai-video') === true
  )
}

/**
 * Upstream rule: video -> landscape, image -> provider-image,
 * DeepSeek/Moonshot -> orbit, everything else -> cube.
 */
export function getCardArtworkUrl(model: PricingModel): string {
  const text = [
    model.model_name,
    model.tags ?? '',
    model.category ?? '',
    model.vendor_name ?? '',
  ]
    .join(' ')
    .toLowerCase()

  if (isVideoModel(model)) return CARD_ARTWORKS.video
  if (/image|图像|绘图|flux|midjourney|dall|stable-diffusion|qwen-image/i.test(text)) {
    return CARD_ARTWORKS.image
  }
  if (/deepseek|moonshot|月之暗面|kimi/i.test(text)) return CARD_ARTWORKS.orbit
  return CARD_ARTWORKS.cube
}
