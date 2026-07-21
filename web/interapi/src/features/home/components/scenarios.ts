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

export interface TerminalScenario {
  id: string
  label: string
  path: string
  model: string
  endpoint: string
  kind: 'chat' | 'image'
  response: string
}

export const TERMINAL_SCENARIOS: TerminalScenario[] = [
  {
    id: 'claude',
    label: 'claude',
    path: 'interapi — /v1/chat/completions',
    model: 'claude-opus-4-1',
    endpoint: '/v1/chat/completions',
    kind: 'chat',
    response: '← 200 OK · model: claude-opus-4-1 · role: assistant · streamed',
  },
  {
    id: 'gpt',
    label: 'gpt',
    path: 'interapi — /v1/chat/completions',
    model: 'gpt-4.1',
    endpoint: '/v1/chat/completions',
    kind: 'chat',
    response: '← 200 OK · model: gpt-4.1 · role: assistant · streamed',
  },
  {
    id: 'gemini',
    label: 'gemini',
    path: 'interapi — /v1/chat/completions',
    model: 'gemini-2.5-pro',
    endpoint: '/v1/chat/completions',
    kind: 'chat',
    response: '← 200 OK · model: gemini-2.5-pro · role: assistant · streamed',
  },
  {
    id: 'deepseek',
    label: 'deepseek',
    path: 'interapi — /v1/chat/completions',
    model: 'deepseek-v3.2',
    endpoint: '/v1/chat/completions',
    kind: 'chat',
    response: '← 200 OK · model: deepseek-v3.2 · role: assistant · streamed',
  },
  {
    id: 'image',
    label: 'image',
    path: 'interapi — /v1/images/generations',
    model: 'flux-1.1-pro',
    endpoint: '/v1/images/generations',
    kind: 'image',
    response: '← 201 Created · model: flux-1.1-pro · image/png · returned',
  },
]

export const TERMINAL_ROTATION_ORDER = [
  'claude',
  'gpt',
  'gemini',
  'deepseek',
  'image',
] as const
