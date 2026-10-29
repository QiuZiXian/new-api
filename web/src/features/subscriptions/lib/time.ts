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

/**
 * Remaining time until a unix timestamp (in seconds).
 *
 * Every field is derived from `endTime - now` with flooring, so the countdown
 * can never disagree with the "until" date rendered from the same timestamp.
 * Rounding up used to overstate the remaining days by one.
 */
export interface RemainingTime {
  expired: boolean
  days: number
  hours: number
  minutes: number
  /** Whole seconds left; 0 once the end time has passed. */
  totalSeconds: number
}

export function remainingUntil(
  endTimeSeconds: number,
  nowMs: number = Date.now()
): RemainingTime {
  const seconds = Math.floor(Number(endTimeSeconds || 0) - nowMs / 1000)
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return { expired: true, days: 0, hours: 0, minutes: 0, totalSeconds: 0 }
  }
  return {
    expired: false,
    days: Math.floor(seconds / 86400),
    hours: Math.floor((seconds % 86400) / 3600),
    minutes: Math.floor((seconds % 3600) / 60),
    totalSeconds: seconds,
  }
}

/**
 * Whole days from now until a future unix timestamp, rounded up: a period that
 * starts later today still reads as starting in 1 day rather than 0.
 */
export function daysUntil(startTimeSeconds: number, nowMs: number = Date.now()) {
  const seconds = Number(startTimeSeconds || 0) - nowMs / 1000
  if (!Number.isFinite(seconds) || seconds <= 0) return 0
  return Math.ceil(seconds / 86400)
}
