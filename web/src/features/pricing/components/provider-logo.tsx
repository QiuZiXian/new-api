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
import { getProviderLogoUrl } from '../lib/artwork'

interface ProviderLogoProps {
  name?: string
  size?: number
  className?: string
}

/** Fallback logo for vendors without an icon configured in model metadata. */
export function ProviderLogo(props: ProviderLogoProps) {
  const url = getProviderLogoUrl(props.name)
  if (!url) return null
  const size = props.size ?? 16

  return (
    <img
      src={url}
      alt=''
      loading='lazy'
      referrerPolicy='no-referrer'
      style={{ width: size, height: size }}
      className={props.className ?? 'object-contain'}
      onError={(event) => {
        event.currentTarget.style.display = 'none'
      }}
    />
  )
}
