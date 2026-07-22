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

import { InterapiWordmark } from '../interapi-wordmark'

// Storefront-only footer (DESIGN.md `footer-section`). Replaces the shared
// <Footer/> on the redesigned landing only — the shared footer (and its
// project attribution) stays in place for the console.
export function StorefrontFooter() {
  const { t } = useTranslation()

  return (
    <footer
      className='st-rule-top'
      style={{ paddingBlock: 32, color: 'var(--st-mute)' }}
    >
      <div className='st-container'>
        <div className='flex flex-col gap-6 md:flex-row md:items-end md:justify-between'>
          <div>
            <InterapiWordmark
              className='text-[10px] leading-[1.1]'
            />
            <p className='mt-4 text-[13px]' style={{ lineHeight: 1.8 }}>
              {t('Private AI API access. No email. Crypto only.')}
            </p>
          </div>
          <nav className='flex flex-wrap gap-x-6 gap-y-2 text-[14px]'>
            <Link to='/pricing'>{t('Pricing')}</Link>
            <Link to='/about'>{t('About')}</Link>
            <Link to='/privacy-policy'>{t('Privacy')}</Link>
            <Link to='/user-agreement'>{t('Terms')}</Link>
          </nav>
        </div>
        <p className='mt-8 text-[13px]' style={{ color: 'var(--st-ash)' }}>
          © InterAPI
        </p>
      </div>
    </footer>
  )
}
