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

const base = {
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: 2.2,
} as const

export function NoEmailIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-3.5'>
      <path d='M4 7l5 5-5 5M12 17h8' />
    </svg>
  )
}

export function PayCryptoIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-3.5'>
      <circle cx='12' cy='12' r='8' />
      <path d='M9 8h4a2.5 2.5 0 010 5M9 13h4.5a2.5 2.5 0 010 5M9 8v10M11 6v2M11 18v2' />
    </svg>
  )
}

export function NoLogsIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-3.5'>
      <path d='M5 5l14 14M5 19L19 5' />
    </svg>
  )
}

export function ForwardAnonymousIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-3.5'>
      <path d='M4 12h12M12 6l6 6-6 6' />
    </svg>
  )
}
