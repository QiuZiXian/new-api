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
import { type SVGProps } from 'react'

type IconSubtitleEraseProps = SVGProps<SVGSVGElement> & {
  size?: number
}

/**
 * Subtitle erase channel icon: a video frame whose subtitle bars are struck through.
 * Uses currentColor so it follows the light/dark theme.
 */
export function IconSubtitleErase({
  size = 20,
  ...props
}: IconSubtitleEraseProps) {
  return (
    <svg
      xmlns='http://www.w3.org/2000/svg'
      viewBox='0 0 24 24'
      width={size}
      height={size}
      fill='none'
      stroke='currentColor'
      strokeLinecap='round'
      strokeLinejoin='round'
      aria-hidden='true'
      {...props}
    >
      <rect
        x='2.75'
        y='4.75'
        width='18.5'
        height='14.5'
        rx='3'
        strokeWidth='1.6'
        opacity='0.55'
      />
      <path d='M6.5 15.25h4' strokeWidth='1.6' opacity='0.55' />
      <path d='M13.25 15.25h4.25' strokeWidth='1.6' opacity='0.55' />
      <path d='M4.75 18.5 19.25 5.5' strokeWidth='2.2' />
    </svg>
  )
}
