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

import { InterapiWordmark } from './interapi-wordmark'

// The home page's single dark surface (DESIGN.md `hero-tui-mockup`): a faux
// terminal framing an InterAPI request — ASCII wordmark, a prompt row showing
// a model-tagged command with a blinking cursor, and keybinding hints.
export function HeroTuiMockup({ className }: { className?: string }) {
  const { t } = useTranslation()

  return (
    <div
      className={`overflow-hidden ${className ?? ''}`}
      style={{
        background: 'var(--st-surface-dark)',
        color: 'var(--st-on-dark)',
        borderRadius: 0,
        padding: 'clamp(32px, 6vw, 64px) clamp(20px, 4vw, 32px)',
      }}
    >
      {/* ASCII wordmark */}
      <div className='flex justify-center'>
        <InterapiWordmark className='overflow-x-auto' />
      </div>

      {/* Prompt row — the inset command line */}
      <div
        className='mx-auto mt-10 max-w-xl text-[15px] leading-relaxed'
        style={{
          background: 'var(--st-surface-dark-elevated)',
          color: 'var(--st-on-dark)',
          borderRadius: 4,
          padding: '10px 14px',
        }}
      >
        <div className='flex flex-wrap items-center gap-x-2 gap-y-1'>
          <span style={{ color: 'var(--st-success)' }}>❯</span>
          <span>chat</span>
          <span style={{ color: 'var(--st-on-dark-mute)' }}>--model</span>
          <span style={{ color: 'var(--st-accent)' }}>[claude-opus-4]</span>
          <span
            className='terminal-demo-blink'
            style={{ background: 'var(--st-on-dark)', color: 'var(--st-surface-dark)' }}
          >
            &nbsp;
          </span>
        </div>
        <div className='mt-2' style={{ color: 'var(--st-on-dark-mute)' }}>
          ↳ {t('streaming response')}
          <span style={{ color: 'var(--st-success)' }}> · ✓ 200 OK</span>
          {' · 142 tok/s · 0.42¢'}
        </div>
      </div>

      {/* Keybinding hints */}
      <div
        className='mt-8 flex flex-wrap items-center justify-center gap-x-6 gap-y-1 text-[13px]'
        style={{ color: 'var(--st-ash)' }}
      >
        <span>
          <kbd>tab</kbd> {t('switch model')}
        </span>
        <span>
          <kbd>ctrl-k</kbd> {t('commands')}
        </span>
        <span>
          <kbd>↵</kbd> {t('send')}
        </span>
      </div>
    </div>
  )
}
