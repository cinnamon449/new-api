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
  BenefitImageIcon,
  BenefitPriceIcon,
  BenefitShieldIcon,
} from './benefit-icons'

interface Benefit {
  Icon: () => React.JSX.Element
  title: string
  body: string
}

export function BenefitsSection() {
  const { t } = useTranslation()

  const benefits: Benefit[] = [
    {
      Icon: BenefitPriceIcon,
      title: t('Up to 70% cheaper'),
      body: t(
        'Provider rates plus a thin margin, billed per call — no subscription and no per-seat tax. Often dramatically cheaper than going direct once you stop paying for the brand.'
      ),
    },
    {
      Icon: BenefitShieldIcon,
      title: t('Model integrity'),
      body: t(
        'The model you request is the model you get — no silent swaps, no quiet downgrades, no "turbo" bait-and-switch. The returned headers confirm exactly what ran.'
      ),
    },
    {
      Icon: BenefitImageIcon,
      title: t('Image generation'),
      body: t(
        'Same key, same endpoint — FLUX, gpt-image, Imagen. Text-to-image and edits without opening a second account or wiring a second gateway.'
      ),
    },
  ]

  return (
    <section className='border-border-soft border-b'>
      <div className='mx-auto flex max-w-[1080px] flex-col gap-12 px-7 py-16 md:py-24'>
        <div className='max-w-[40ch]'>
          <p className='text-muted-foreground mb-3 font-mono text-xs font-medium tracking-[0.14em] uppercase'>
            // {t('Why InterAPI')}
          </p>
          <h2 className='text-[clamp(28px,3.6vw,42px)] font-bold tracking-tight'>
            {t('Cheaper, honest, and it draws.')}
          </h2>
        </div>
        <div className='grid grid-cols-1 gap-8 md:grid-cols-3'>
          {benefits.map((b) => (
            <div key={b.title}>
              <span className='border-border text-primary mb-5 grid size-8 place-items-center rounded-sm border'>
                <b.Icon />
              </span>
              <h3 className='mb-1.5 text-[19px] font-semibold'>{b.title}</h3>
              <p className='text-foreground/80 text-sm leading-[1.55]'>
                {b.body}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
