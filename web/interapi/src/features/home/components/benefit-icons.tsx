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
  strokeWidth: 1.6,
} as const

export function BenefitPriceIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-4'>
      <path d='M12 3v18M16 7H9.5a2.5 2.5 0 000 5h5a2.5 2.5 0 010 5H7' />
    </svg>
  )
}

export function BenefitShieldIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-4'>
      <path d='M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6l7-3z' />
      <path d='M9 12l2 2 4-4' />
    </svg>
  )
}

export function BenefitImageIcon() {
  return (
    <svg viewBox='0 0 24 24' {...base} className='size-4'>
      <rect x='3' y='5' width='18' height='14' rx='1' />
      <circle cx='8.5' cy='10' r='1.5' />
      <path d='M21 16l-5-4-7 7' />
    </svg>
  )
}
