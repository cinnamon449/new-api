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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { HeroTuiMockup } from '../hero-tui-mockup'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()

  return (
    <section className='st-section' style={{ paddingBottom: 64 }}>
      <div className='st-container'>
        {/* News/announcement chip */}
        <span
          className='inline-block text-[13px]'
          style={{
            background: 'var(--st-surface-dark)',
            color: 'var(--st-on-dark)',
            borderRadius: 4,
            padding: '3px 10px',
          }}
        >
          {t('Privacy-first · No email · Crypto-only')}
        </span>

        <h1
          className='mt-6 font-bold'
          style={{
            fontSize: 'clamp(28px, 5vw, 38px)',
            lineHeight: 1.5,
            letterSpacing: 0,
          }}
        >
          {t('Private AI API access.')}
          <br />
          <span style={{ color: 'var(--st-mute)' }}>
            {t('No email. Crypto only. Prompts never stored.')}
          </span>
        </h1>

        <p
          className='mt-5 max-w-2xl'
          style={{
            color: 'var(--st-body)',
            fontSize: 16,
            lineHeight: 1.6,
          }}
        >
          {t(
            'One key for every model — OpenAI, Anthropic, Google, Deepseek, Z.ai and more. Sign up with just a username, pay in crypto, and keep your identity off the providers’ servers.'
          )}
        </p>

        <div className='mt-8 flex flex-wrap items-center gap-3'>
          {props.isAuthenticated ? (
            <Link to='/dashboard' className='st-btn st-btn-primary'>
              {t('Go to Dashboard')}
            </Link>
          ) : (
            <Link to='/sign-up' className='st-btn st-btn-primary'>
              {t('Get Started')}
            </Link>
          )}
          <Link to='/pricing' className='st-btn st-btn-secondary'>
            {t('View Pricing')}
          </Link>
        </div>

        <p
          className='mt-8 text-[13px]'
          style={{ color: 'var(--st-mute)', lineHeight: 2 }}
        >
          {t('Works with Claude Code · Codex · OpenCode · OpenClaw · Gemini CLI')}
        </p>

        <HeroTuiMockup className='mt-12' />
      </div>
    </section>
  )
}
