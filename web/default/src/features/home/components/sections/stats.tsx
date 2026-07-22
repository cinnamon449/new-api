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
import { useTranslation } from 'react-i18next'

// DESIGN.md `chart-tile` — abstract sparse-line ASCII plots beneath a stat,
// each with a `Fig N.` caption. Static (no JS) so it degrades safely.
function Tile({
  value,
  plot,
  caption,
}: {
  value: string
  plot: string
  caption: string
}) {
  return (
    <div
      className='st-block'
      style={{ padding: 16 }}
    >
      <div
        className='font-bold tabular-nums'
        style={{ color: 'var(--st-ink)', fontSize: 'clamp(28px, 4vw, 36px)', lineHeight: 1.1 }}
      >
        {value}
      </div>
      <pre
        className='mt-3 overflow-x-auto'
        style={{
          color: 'var(--st-body)',
          margin: 0,
          fontSize: 16,
          lineHeight: 1.4,
        }}
        aria-hidden
      >
        {plot}
      </pre>
      <p
        className='mt-3'
        style={{ color: 'var(--st-mute)', fontSize: 14, lineHeight: 1.8 }}
      >
        {caption}
      </p>
    </div>
  )
}

export function Stats() {
  const { t } = useTranslation()

  return (
    <section className='st-section st-rule-top'>
      <div className='st-container'>
        <h2 className='st-label'>{t('By the numbers')}</h2>
        <div className='mt-8 grid grid-cols-1 gap-px md:grid-cols-3'>
          <Tile
            value={t('Up to 70%')}
            plot='█▇▆▅▄▃▂▁'
            caption={t('Fig 1. lower cost than going direct to each provider')}
          />
          <Tile
            value={t('40+')}
            plot='▁▂▃▄▅▆▇█'
            caption={t(
              'Fig 2. models from OpenAI, Anthropic, Google, Deepseek, Z.ai & more'
            )}
          />
          <Tile
            value='0'
            plot='· · · · · · · ·'
            caption={t('Fig 3. prompts stored — your messages are never persisted')}
          />
        </div>
      </div>
    </section>
  )
}
