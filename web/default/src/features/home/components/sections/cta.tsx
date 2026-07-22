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

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  return (
    <section className='st-section st-rule-top'>
      <div className='st-container'>
        <div
          className='st-block'
          style={{ padding: 'clamp(32px, 5vw, 56px)' }}
        >
          <h2
            className='font-bold'
            style={{ fontSize: 'clamp(24px, 4vw, 32px)', lineHeight: 1.4 }}
          >
            {t('Start building privately.')}
          </h2>
          <p
            className='mt-4 max-w-xl'
            style={{ color: 'var(--st-body)', lineHeight: 1.6 }}
          >
            {t('Create an account in seconds — no email required, pay in crypto.')}
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
        </div>
      </div>
    </section>
  )
}
