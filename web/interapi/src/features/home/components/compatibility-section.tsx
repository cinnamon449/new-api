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

interface CompatCell {
  name: string
  tag: string
  lines: { k?: string; v?: string; plain?: string }[]
}

const COMPAT_CELLS: CompatCell[] = [
  {
    name: 'Claude Code',
    tag: 'anthropic api',
    lines: [
      { k: 'ANTHROPIC_BASE_URL', v: 'https://api.interapi.dev' },
      { k: 'ANTHROPIC_API_KEY', v: '$INTERAPI_KEY' },
      { plain: 'claude' },
    ],
  },
  {
    name: 'Codex',
    tag: 'openai api',
    lines: [
      { k: 'OPENAI_BASE_URL', v: 'https://api.interapi.dev/v1' },
      { k: 'OPENAI_API_KEY', v: '$INTERAPI_KEY' },
      { plain: 'codex' },
    ],
  },
  {
    name: 'OpenCode',
    tag: 'provider gateway',
    lines: [
      { plain: 'provider: openai-compatible' },
      { plain: 'base: https://api.interapi.dev/v1' },
      { plain: 'key:  $INTERAPI_KEY' },
    ],
  },
  {
    name: 'OpenClaw',
    tag: 'openai api',
    lines: [
      { k: 'OPENAI_BASE_URL', v: 'https://api.interapi.dev/v1' },
      { k: 'OPENAI_API_KEY', v: '$INTERAPI_KEY' },
    ],
  },
  {
    name: 'Gemini CLI',
    tag: 'custom endpoint',
    lines: [
      { k: 'GEMINI_API_BASE', v: 'https://api.interapi.dev/v1' },
      { k: 'GEMINI_API_KEY', v: '$INTERAPI_KEY' },
    ],
  },
  {
    name: 'Anything else',
    tag: 'raw http',
    lines: [
      { plain: 'curl https://api.interapi.dev/v1/chat/completions \\' },
      { plain: '  -H "Authorization: Bearer $KEY"' },
    ],
  },
]

export function CompatibilitySection() {
  const { t } = useTranslation()

  return (
    <section className='border-border-soft border-b'>
      <div className='mx-auto flex max-w-[1080px] flex-col gap-10 px-7 py-16 md:py-24'>
        <div className='max-w-[48ch]'>
          <p className='text-muted-foreground mb-3 font-mono text-xs font-medium tracking-[0.14em] uppercase'>
            // {t('Works with your stack')}
          </p>
          <h2 className='text-[clamp(28px,3.6vw,42px)] font-bold tracking-tight'>
            {t('Point your tools at one URL. Done.')}
          </h2>
          <p className='text-foreground/80 mt-3.5 max-w-[58ch] text-[18px] leading-[1.55]'>
            {t(
              'InterAPI speaks the OpenAI and Anthropic wire formats, so the agents and CLIs you already use just work — set a base URL and your key.'
            )}
          </p>
        </div>

        <div className='grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-[repeat(auto-fill,minmax(280px,1fr))]'>
          {COMPAT_CELLS.map((cell) => (
            <div
              key={cell.name}
              className='bg-surface-2 border-border rounded-md border p-5'
            >
              <div className='flex items-center justify-between text-base font-bold'>
                <span>{cell.name}</span>
                <span className='text-meta font-mono text-[10px] font-medium tracking-[0.06em] uppercase'>
                  {cell.tag}
                </span>
              </div>
              <pre className='text-foreground/80 mt-3.5 font-mono text-xs leading-relaxed break-words whitespace-pre-wrap'>
                {cell.lines.map((line) => {
                  const lineKey =
                    line.k ?? line.plain ?? line.v ?? Math.random().toString(36)
                  return (
                    <div key={lineKey}>
                      {line.k && <span className='text-primary'>{line.k}</span>}
                      {line.k && line.v && '='}
                      {line.v && <span className='text-success'>{line.v}</span>}
                      {line.plain && <span>{line.plain}</span>}
                    </div>
                  )
                })}
              </pre>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
