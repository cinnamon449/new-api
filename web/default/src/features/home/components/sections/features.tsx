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

// Privacy pillars — the InterAPI value props as DESIGN.md `list-row` blocks with
// `[+]` ASCII markers. Copy is worded to match what new-api actually does
// (honest reframe); see docs/interapi-privacy-setup.md.
function Pillar({ title, desc }: { title: string; desc: string }) {
  return (
    <div
      className='flex items-start gap-4 py-4'
      style={{ borderBottom: '1px solid var(--st-hairline)' }}
    >
      <span
        className='shrink-0 font-bold'
        style={{ color: 'var(--st-ink)' }}
        aria-hidden
      >
        [+]
      </span>
      <div className='min-w-0'>
        <p className='font-bold' style={{ color: 'var(--st-ink)', lineHeight: 1.5 }}>
          {title}
        </p>
        <p className='mt-1' style={{ color: 'var(--st-body)', lineHeight: 1.6 }}>
          {desc}
        </p>
      </div>
    </div>
  )
}

export function Features() {
  const { t } = useTranslation()

  return (
    <section className='st-section st-rule-top'>
      <div className='st-container'>
        <h2 className='st-label'>{t('Why InterAPI')}</h2>

        <div
          className='mt-8'
          style={{ borderTop: '1px solid var(--st-hairline)' }}
        >
          <Pillar
            title={t('No email required')}
            desc={t(
              'Register with a username and password. We never ask for your email, and nothing ties your account to your real identity.'
            )}
          />
          <Pillar
            title={t('Crypto-only payments')}
            desc={t(
              'Top up with BTC, ETH, USDT and more through Plisio. No card, no bank, no paper trail.'
            )}
          />
          <Pillar
            title={t('Prompts are never stored')}
            desc={t(
              'We record only the token counts we need to bill you — never your messages, inputs, or conversation content.'
            )}
          />
          <Pillar
            title={t('Providers don’t get your identity')}
            desc={t(
              'Upstream requests carry only your prompt. We forward no IP address, user-agent, or account info to the model provider.'
            )}
          />
          <Pillar
            title={t('Minimal logging')}
            desc={t(
              'Request IPs are off by default. We keep only what’s required to operate and bill — nothing more, nothing traced to you.'
            )}
          />
          <Pillar
            title={t('Model integrity')}
            desc={t(
              'What you request is what you get — no silent model swaps, downgrades, or bait-and-switch.'
            )}
          />
        </div>
      </div>
    </section>
  )
}
