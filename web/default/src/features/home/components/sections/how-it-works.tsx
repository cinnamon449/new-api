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
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

// "Works with your tools" — a DESIGN.md `button-tab` strip + `install-snippet`
// showing how to point each tool at InterAPI. Tool names are proper nouns
// (untranslated); only chrome strings go through t().
const TOOLS = [
  {
    id: 'claude-code',
    label: 'Claude Code',
    snippet: `export ANTHROPIC_BASE_URL=https://api.interapi.ai
export ANTHROPIC_AUTH_TOKEN=sk-your-key`,
  },
  {
    id: 'codex',
    label: 'Codex',
    snippet: `export OPENAI_BASE_URL=https://api.interapi.ai/v1
export OPENAI_API_KEY=sk-your-key`,
  },
  {
    id: 'opencode',
    label: 'OpenCode',
    snippet: `export OPENAI_BASE_URL=https://api.interapi.ai/v1
export OPENAI_API_KEY=sk-your-key`,
  },
  {
    id: 'openclaw',
    label: 'OpenClaw',
    snippet: `export OPENAI_BASE_URL=https://api.interapi.ai/v1
export OPENAI_API_KEY=sk-your-key`,
  },
  {
    id: 'gemini-cli',
    label: 'Gemini CLI',
    snippet: `export GEMINI_API_BASE=https://api.interapi.ai/v1beta
export GEMINI_API_KEY=sk-your-key`,
  },
]

export function HowItWorks() {
  const { t } = useTranslation()
  const [activeId, setActiveId] = useState(TOOLS[0].id)
  const [copied, setCopied] = useState(false)

  const active = TOOLS.find((tool) => tool.id === activeId) ?? TOOLS[0]

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(active.snippet)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1600)
    } catch {
      // Clipboard may be unavailable (e.g. non-secure context); ignore.
    }
  }

  return (
    <section className='st-section st-rule-top'>
      <div className='st-container'>
        <h2 className='st-label'>{t('Works with your tools')}</h2>
        <p
          className='mt-3 max-w-2xl'
          style={{ color: 'var(--st-body)', lineHeight: 1.6 }}
        >
          {t(
            'Point your existing tools at InterAPI — one key, every model. No new app to install.'
          )}
        </p>

        {/* Tab strip */}
        <div
          className='mt-8 flex flex-wrap gap-2'
          style={{ borderBottom: '1px solid var(--st-hairline-strong)' }}
        >
          {TOOLS.map((tool) => {
            const isActive = tool.id === activeId
            return (
              <button
                key={tool.id}
                type='button'
                onClick={() => setActiveId(tool.id)}
                style={{
                  background: 'transparent',
                  color: isActive ? 'var(--st-ink)' : 'var(--st-mute)',
                  padding: '8px 16px',
                  border: 'none',
                  borderBottom: isActive
                    ? '2px solid var(--st-ash)'
                    : '2px solid transparent',
                  cursor: 'pointer',
                  fontSize: 16,
                  fontWeight: 500,
                  lineHeight: 2,
                }}
              >
                {tool.label}
              </button>
            )
          })}
        </div>

        {/* Install snippet */}
        <div className='mt-6 flex items-stretch gap-2'>
          <pre
            className='st-snippet flex-1'
            style={{ margin: 0 }}
          >
            {active.snippet}
          </pre>
          <button type='button' onClick={copy} className='st-btn st-btn-secondary'>
            {copied ? t('Copied') : t('Copy')}
          </button>
        </div>
      </div>
    </section>
  )
}
