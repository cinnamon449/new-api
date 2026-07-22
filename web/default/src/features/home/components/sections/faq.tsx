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

// FAQ — DESIGN.md `faq-row` blocks using native <details> (works without JS).
// Answers are worded to match what new-api actually does (honest reframe).
function FaqItem({ question, answer }: { question: string; answer: string }) {
  return (
    <details>
      <summary>
        <span className='st-toggle' aria-hidden />
        <span className='st-q'>{question}</span>
      </summary>
      <div className='st-a'>{answer}</div>
    </details>
  )
}

export function Faq() {
  const { t } = useTranslation()

  return (
    <section className='st-section st-rule-top'>
      <div className='st-container'>
        <h2 className='st-label'>{t('FAQ')}</h2>
        <div className='mt-8'>
          <FaqItem
            question={t('Do I need an email to sign up?')}
            answer={t(
              'No. Register with just a username and password. Email is optional — only add one if you want account notifications.'
            )}
          />
          <FaqItem
            question={t('How do I pay?')}
            answer={t(
              'Cryptocurrency only. Top up with BTC, ETH, USDT and more through Plisio — no card, bank, or personal financial details.'
            )}
          />
          <FaqItem
            question={t('Do you keep logs of my requests?')}
            answer={t(
              'We store only what’s required to bill you: token counts, model, and timestamps. Your message content is never stored, and request IPs are off by default. We don’t log anything we don’t need.'
            )}
          />
          <FaqItem
            question={t('Will OpenAI or Anthropic see who I am?')}
            answer={t(
              'No. Requests forwarded upstream carry only your prompt — not your IP address, user-agent, or account details.'
            )}
          />
          <FaqItem
            question={t('Are the models real?')}
            answer={t(
              'Yes. Model integrity means the model you request is the model you get — no silent substitutions or downgrades.'
            )}
          />
          <FaqItem
            question={t('Which tools are compatible?')}
            answer={t(
              'Any OpenAI, Anthropic, or Gemini-compatible client — Claude Code, Codex, OpenCode, OpenClaw, Gemini CLI, and more.'
            )}
          />
        </div>
      </div>
    </section>
  )
}
