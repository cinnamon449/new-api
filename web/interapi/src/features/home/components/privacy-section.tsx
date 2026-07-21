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

import {
  NoEmailIcon,
  NoLogsIcon,
  PayCryptoIcon,
  ForwardAnonymousIcon,
} from './privacy-icons'

interface PrivacyCell {
  Icon: () => React.JSX.Element
  title: string
  body: string
  keyword: string
}

export function PrivacySection() {
  const { t } = useTranslation()

  const cells: PrivacyCell[] = [
    {
      Icon: NoEmailIcon,
      title: t('No email to sign up'),
      body: t(
        'Generate a key. That key is the account — no email, no phone, no recovery quiz that links back to a real person.'
      ),
      keyword: '→ identity: none',
    },
    {
      Icon: PayCryptoIcon,
      title: t('Pay in crypto, including Monero'),
      body: t(
        "Top up with BTC, ETH, USDT or XMR. Monero means the payment itself can't be traced back to your key."
      ),
      keyword: '→ payment: unlinkable',
    },
    {
      Icon: NoLogsIcon,
      title: t('No logs, no telemetry'),
      body: t(
        "Prompts aren't stored, requests aren't logged, analytics aren't collected. There's nothing to subpoena and nothing to breach."
      ),
      keyword: '→ retention: zero',
    },
    {
      Icon: ForwardAnonymousIcon,
      title: t('Providers see your prompt, not you'),
      body: t(
        'We strip your identity before forwarding. The model maker receives a request — never a profile, an account, or a billing trail.'
      ),
      keyword: '→ upstream: anonymous',
    },
  ]

  return (
    <section className='border-border-soft border-b'>
      <div className='mx-auto flex max-w-[1080px] flex-col gap-10 px-7 py-16 md:py-24'>
        <div className='max-w-[48ch]'>
          <p className='text-muted-foreground mb-3 font-mono text-xs font-medium tracking-[0.14em] uppercase'>
            // {t('Private by default')}
          </p>
          <h2 className='text-[clamp(28px,3.6vw,42px)] font-bold tracking-tight'>
            {t("Privacy isn't a toggle. It's the architecture.")}
          </h2>
          <p className='text-foreground/80 mt-3.5 max-w-[58ch] text-[18px] leading-[1.55]'>
            {t(
              'Most "private" AI gateways add tracking on top. InterAPI removes it — no account footprint, untraceable payment, and nothing stored to leak or hand over.'
            )}
          </p>
        </div>

        <div className='border-border-soft grid grid-cols-1 gap-px overflow-hidden rounded-md border bg-[var(--border-soft)] md:grid-cols-2'>
          {cells.map((cell) => (
            <div key={cell.title} className='bg-background px-6 py-[26px]'>
              <div className='mb-2.5 flex items-center gap-3'>
                <span className='text-success flex size-[26px] shrink-0 place-items-center rounded-sm border border-[color-mix(in_srgb,var(--success)_45%,transparent)]'>
                  <cell.Icon />
                </span>
                <h3 className='text-[17px] font-semibold'>{cell.title}</h3>
              </div>
              <p className='text-foreground/80 text-sm leading-[1.55]'>
                {cell.body}
              </p>
              <p className='text-success mt-3 font-mono text-[11px] tracking-[0.06em] uppercase'>
                {cell.keyword}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
