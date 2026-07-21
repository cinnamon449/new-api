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

import { Button } from '@/components/ui/button'

import { HeroTerminal } from './hero-terminal'

export function Hero() {
  const { t } = useTranslation()

  return (
    <section className='border-border-soft border-b'>
      <div className='mx-auto grid max-w-[1080px] grid-cols-1 gap-14 px-7 py-16 md:py-24 lg:grid-cols-[1.05fr_1fr] lg:gap-24'>
        <div>
          <p className='text-muted-foreground mb-5 font-mono text-xs font-medium tracking-[0.14em] uppercase'>
            // {t('Private AI infrastructure')}
          </p>
          <h1 className='text-[clamp(40px,6vw,68px)] leading-[1.04] font-bold tracking-tight'>
            {t('Many models.')}
            <br />
            {t('One key.')}
            <br />
            {t('Zero tracking.')}
          </h1>
          <p className='text-foreground/80 mt-5 max-w-[58ch] text-[18px] leading-[1.55]'>
            {t(
              'One endpoint and one key for every major model — OpenAI, Anthropic, Google, DeepSeek, Z.ai. No email to sign up, no logs stored, no identity on the wire. Pay in Monero.'
            )}
          </p>
          <div className='mt-8 flex flex-wrap gap-3'>
            <Button
              render={<Link to='/sign-in' />}
              className='h-10 rounded-sm px-5 text-sm font-medium'
            >
              {t('Open Console')}
            </Button>
            <Button
              variant='ghost'
              render={<Link to='/pricing' />}
              className='group h-10 rounded-sm px-1.5 text-sm font-medium'
            >
              {t('See how pricing works')}
              <span className='transition-transform group-hover:translate-x-0.5'>
                →
              </span>
            </Button>
          </div>
          <div className='mt-8 flex flex-wrap gap-2'>
            {[
              t('No email'),
              t('No logs'),
              'XMR · BTC · ETH · USDT',
              t('OpenAI-compliant endpoint'),
            ].map((tag) => (
              <span
                key={tag}
                className='border-border rounded-sm border px-2.5 py-1 font-mono text-xs'
              >
                {tag}
              </span>
            ))}
          </div>
        </div>

        <HeroTerminal className='self-center' />
      </div>
    </section>
  )
}
