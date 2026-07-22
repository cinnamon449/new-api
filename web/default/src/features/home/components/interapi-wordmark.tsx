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
// InterAPI brand mark — a block-pixel ASCII wordmark rendered in JetBrains
// Mono (DESIGN.md: "wordmark rendered as block-pixel ASCII"). Each glyph is a
// 5-row bitmap; rows within a glyph are equal width so monospace alignment
// holds when glyphs are concatenated with a single-space gap.

const GLYPHS: Record<string, string[]> = {
  I: ['█', '█', '█', '█', '█'],
  N: ['█  █', '██ █', '█ ██', '█  █', '█  █'],
  T: ['████', '  █ ', '  █ ', '  █ ', '  █ '],
  E: ['████', '█   ', '███ ', '█   ', '████'],
  R: ['███ ', '█  █', '███ ', '█ █ ', '█  █'],
  A: [' ██ ', '█  █', '████', '█  █', '█  █'],
  P: ['███ ', '█  █', '███ ', '█   ', '█   '],
}

interface InterapiWordmarkProps {
  word?: string
  className?: string
}

export function InterapiWordmark({
  word = 'INTERAPI',
  className,
}: InterapiWordmarkProps) {
  const letters = [...word.toUpperCase()]
  const rows = [0, 1, 2, 3, 4].map((rowIndex) =>
    letters.map((letter) => GLYPHS[letter]?.[rowIndex] ?? '').join(' ')
  )

  return (
    <pre
      aria-label='InterAPI'
      role='img'
      className={`m-0 inline-block whitespace-pre leading-[1.05] ${className ?? ''}`}
    >
      {rows.join('\n')}
    </pre>
  )
}
