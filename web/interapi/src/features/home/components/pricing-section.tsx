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

interface PriceStep {
  step: string
  title: string
  body: string
}

export function PricingSection() {
  const { t } = useTranslation()

  const steps: PriceStep[] = [
    {
      step: `01 · ${t('Top up')}`,
      title: t('Fund with crypto or card'),
      body: t(
        'Add credit with BTC, ETH, USDT or Monero — or a card if you prefer. Monero keeps the top-up unlinkable from your key.'
      ),
    },
    {
      step: `02 · ${t('Use')}`,
      title: t('Metered per call'),
      body: t(
        "Every request is billed at the underlying provider's rate plus a thin margin, debited from your balance in real time as tokens flow."
      ),
    },
    {
      step: `03 · ${t('Never subscribe')}`,
      title: t('Zero recurring charges'),
      body: t(
        "No monthly fee, no minimum, no seat licenses. Credit sits until you spend it. Stop any time — there's nothing to cancel."
      ),
    },
  ]

  return (
    <section id='pricing' className='border-border-soft border-b'>
      <div className='mx-auto flex max-w-[1080px] flex-col gap-10 px-7 py-16 md:py-24'>
        <div className='mx-auto max-w-[40ch] text-center'>
          <p className='text-muted-foreground mb-3 font-mono text-xs font-medium tracking-[0.14em] uppercase'>
            // {t('Pricing')}
          </p>
          <p className='text-[clamp(24px,3vw,34px)] font-bold tracking-tight'>
            {t('Pay per call. No plans, no tiers, no seats.')}
          </p>
          <p className='text-foreground/80 mt-4 text-[18px] leading-[1.55]'>
            {t(
              "You're not subscribing to software. You're buying tokens of compute, metered in real time, at the model's real rate."
            )}
          </p>
        </div>

        <div className='border-border-soft grid grid-cols-1 gap-px overflow-hidden rounded-md border bg-[var(--border-soft)] md:grid-cols-3'>
          {steps.map((s) => (
            <div key={s.step} className='bg-background px-[22px] py-[26px]'>
              <p className='text-primary mb-2.5 font-mono text-[11px] tracking-[0.1em] uppercase'>
                {s.step}
              </p>
              <h3 className='mb-2 text-[17px] font-semibold'>{s.title}</h3>
              <p className='text-foreground/80 text-sm leading-[1.55]'>
                {s.body}
              </p>
            </div>
          ))}
        </div>

        <p className='text-muted-foreground mx-auto mt-2 max-w-[60ch] text-center font-mono text-[13px]'>
          <strong className='text-foreground font-bold'>
            {t('Up to 70% cheaper')}
          </strong>{' '}
          {t(
            'than direct provider billing on common workloads · live per-model rates are published in'
          )}{' '}
          <strong className='text-foreground font-bold'>{t('Console')}</strong>.
        </p>
      </div>
    </section>
  )
}
