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

interface Provider {
  name: string
  meta: string
  models: { name: string; tag: string }[]
}

const PROVIDERS: Provider[] = [
  {
    name: 'OpenAI',
    meta: 'chat · reasoning',
    models: [
      { name: 'gpt-4.1', tag: 'chat' },
      { name: 'o3', tag: 'reasoning' },
      { name: 'gpt-4o', tag: 'multimodal' },
    ],
  },
  {
    name: 'Anthropic',
    meta: 'claude',
    models: [
      { name: 'claude-opus-4-1', tag: 'flagship' },
      { name: 'claude-sonnet-4-5', tag: 'balanced' },
      { name: 'claude-haiku-4-5', tag: 'fast' },
    ],
  },
  {
    name: 'Google',
    meta: 'gemini · veo',
    models: [
      { name: 'gemini-2.5-pro', tag: 'flagship' },
      { name: 'gemini-2.5-flash', tag: 'fast' },
      { name: 'veo-3', tag: 'video' },
    ],
  },
  {
    name: 'DeepSeek',
    meta: 'open · cheap',
    models: [
      { name: 'deepseek-v3.2', tag: 'chat' },
      { name: 'deepseek-r1', tag: 'reasoning' },
      { name: 'deepseek-coder', tag: 'code' },
    ],
  },
  {
    name: 'Z.ai',
    meta: 'glm',
    models: [
      { name: 'glm-4.6', tag: 'flagship' },
      { name: 'glm-4.5v', tag: 'vision' },
      { name: 'glm-air', tag: 'fast' },
    ],
  },
  {
    name: 'Image & media',
    meta: 'gen',
    models: [
      { name: 'flux-1.1-pro', tag: 'image' },
      { name: 'gpt-image-1', tag: 'image' },
      { name: 'imagen-4', tag: 'image' },
    ],
  },
]

export function ProvidersSection() {
  const { t } = useTranslation()

  return (
    <section className='border-border-soft border-b'>
      <div className='mx-auto flex max-w-[1080px] flex-col gap-10 px-7 py-16 md:py-24'>
        <div className='flex flex-wrap items-end justify-between gap-5'>
          <div className='max-w-[46ch]'>
            <p className='text-muted-foreground mb-3 font-mono text-xs font-medium tracking-[0.14em] uppercase'>
              // {t('One key · every provider')}
            </p>
            <h2 className='text-[clamp(28px,3.6vw,42px)] font-bold tracking-tight'>
              {t('Five labs. One endpoint. Same key for all of them.')}
            </h2>
          </div>
          <p className='text-muted-foreground font-mono text-[13px]'>
            {t('swap the')} <code className='text-warning'>model</code>{' '}
            {t('field — nothing else')}
          </p>
        </div>

        <div className='border-border-soft grid grid-cols-2 gap-px overflow-hidden rounded-md border bg-[var(--border-soft)] md:grid-cols-3'>
          {PROVIDERS.map((p) => (
            <div
              key={p.name}
              className='bg-background flex flex-col gap-3 px-5 py-[22px]'
            >
              <div className='flex items-center justify-between'>
                <span className='text-base font-bold'>{p.name}</span>
                <span className='text-meta font-mono text-[11px] tracking-[0.04em] uppercase'>
                  {p.meta}
                </span>
              </div>
              <div className='flex flex-col gap-1'>
                {p.models.map((m) => (
                  <span
                    key={m.name}
                    className='text-foreground/80 font-mono text-[13px]'
                  >
                    {m.name} <span className='text-meta'>· {m.tag}</span>
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>

        <p className='text-muted-foreground font-mono text-[13px]'>
          {t(
            'Representative models per provider — live availability and per-model rates are listed in Console. New models land continuously.'
          )}
        </p>
      </div>
    </section>
  )
}
