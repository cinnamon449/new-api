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

export function CtaStrip() {
  const { t } = useTranslation()

  return (
    <section className='border-border-soft border-b text-center'>
      <div className='mx-auto flex max-w-[640px] flex-col items-center px-7 py-16 md:py-24'>
        <h2 className='text-[clamp(28px,3.6vw,42px)] font-bold tracking-tight'>
          {t('One key. Every model. Your identity stays yours.')}
        </h2>
        <p className='text-muted-foreground mt-3 max-w-[58ch] text-[18px] leading-[1.55]'>
          {t(
            'Generate a key in seconds — no email required. Top up with Monero and send your first request.'
          )}
        </p>
        <div className='mt-8 flex flex-wrap justify-center gap-3'>
          <Button
            render={<Link to='/sign-in' />}
            className='h-10 rounded-sm px-5 text-sm font-medium'
          >
            {t('Open Console')}
          </Button>
          <Button
            variant='outline'
            render={<a href='#' />}
            className='h-10 rounded-sm px-5 text-sm font-medium'
          >
            {t('Read the docs')}
          </Button>
        </div>
      </div>
    </section>
  )
}
