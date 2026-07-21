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
import { useEffect, useRef, useState } from 'react'

import {
  TERMINAL_ROTATION_ORDER,
  TERMINAL_SCENARIOS,
  type TerminalScenario,
} from './scenarios'

interface HeroTerminalProps {
  className?: string
}

function buildCommandLines(s: TerminalScenario): string {
  if (s.kind === 'image') {
    return [
      `$ curl https://api.interapi.dev${s.endpoint} \\`,
      '  -H "Authorization: Bearer $INTERAPI_KEY" \\',
      "  -d '{",
      `    "model": "${s.model}",`,
      '    "prompt": "a quiet terminal at 3am, monochrome",',
      '    "size": "1024x1024"',
      "  }'",
      '',
    ].join('\n')
  }
  return [
    `$ curl https://api.interapi.dev${s.endpoint} \\`,
    '  -H "Authorization: Bearer $INTERAPI_KEY" \\',
    "  -d '{",
    `    "model": "${s.model}",`,
    '    "messages": [{"role":"user","content":"summarize this repo"}]',
    "  }'",
    '',
  ].join('\n')
}

export function HeroTerminal(props: HeroTerminalProps) {
  const [activeId, setActiveId] = useState<string>(TERMINAL_SCENARIOS[0].id)
  const [typedResponse, setTypedResponse] = useState<string>('')
  const typingTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const rotateTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const rotationPausedRef = useRef(false)

  const activeScenario =
    TERMINAL_SCENARIOS.find((s) => s.id === activeId) ?? TERMINAL_SCENARIOS[0]

  const stopTyping = () => {
    if (typingTimerRef.current) {
      clearInterval(typingTimerRef.current)
      typingTimerRef.current = null
    }
  }

  const stopRotation = () => {
    if (rotateTimerRef.current) {
      clearTimeout(rotateTimerRef.current)
      rotateTimerRef.current = null
    }
  }

  const typeOut = (text: string, reduced: boolean) => {
    stopTyping()
    if (reduced) {
      setTypedResponse(text)
      return
    }
    let i = 0
    typingTimerRef.current = setInterval(() => {
      i += 1
      setTypedResponse(text.slice(0, i))
      if (i >= text.length) {
        stopTyping()
      }
    }, 18)
  }

  const selectScenario = (id: string, fromUser: boolean) => {
    const scenario =
      TERMINAL_SCENARIOS.find((s) => s.id === id) ?? TERMINAL_SCENARIOS[0]
    setActiveId(id)
    setTypedResponse('')
    typeOut(
      scenario.response,
      typeof window !== 'undefined' &&
        window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
    if (fromUser) {
      stopRotation()
      scheduleRotation(true)
    }
  }

  const advance = () => {
    const idx = TERMINAL_ROTATION_ORDER.indexOf(
      activeId as (typeof TERMINAL_ROTATION_ORDER)[number]
    )
    const next =
      idx === -1
        ? TERMINAL_ROTATION_ORDER[0]
        : TERMINAL_ROTATION_ORDER[(idx + 1) % TERMINAL_ROTATION_ORDER.length]
    selectScenario(next, false)
  }

  const scheduleRotation = (pauseFirst: boolean) => {
    stopRotation()
    if (rotationPausedRef.current) return
    rotateTimerRef.current = setTimeout(
      () => {
        advance()
        scheduleRotation(false)
      },
      pauseFirst ? 6000 : 4500
    )
  }

  // Boot: initial selection + start rotation
  useEffect(() => {
    selectScenario(TERMINAL_SCENARIOS[0].id, false)
    scheduleRotation(false)
    return () => {
      stopTyping()
      stopRotation()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <div
      className={`bg-surface-2 border-border overflow-hidden rounded-md border font-mono ${props.className ?? ''}`}
      aria-label='Interactive request demo'
      onMouseEnter={() => {
        rotationPausedRef.current = true
        stopRotation()
      }}
      onMouseLeave={() => {
        rotationPausedRef.current = false
        scheduleRotation(false)
      }}
    >
      <div className='border-border-soft text-muted-foreground flex items-center gap-2.5 border-b px-3.5 py-2 text-xs'>
        <span className='border-border bg-surface size-2.5 rounded-full border' />
        <span className='border-border bg-surface size-2.5 rounded-full border' />
        <span className='border-border bg-surface size-2.5 rounded-full border' />
        <span className='text-foreground/80 ml-1.5'>{activeScenario.path}</span>
      </div>
      <div
        role='tablist'
        aria-label='Pick a model'
        className='border-border-soft flex [scrollbar-width:none] gap-0.5 overflow-x-auto border-b px-2 [&::-webkit-scrollbar]:hidden'
      >
        {TERMINAL_SCENARIOS.map((s) => {
          const selected = s.id === activeId
          return (
            <button
              key={s.id}
              role='tab'
              type='button'
              aria-selected={selected}
              onClick={() => selectScenario(s.id, true)}
              className={`cursor-pointer border-b-2 px-3 py-2.5 text-xs whitespace-nowrap transition-colors ${
                selected
                  ? 'border-primary text-foreground'
                  : 'text-muted-foreground hover:text-foreground/80 border-transparent'
              }`}
            >
              {s.label}
            </button>
          )
        })}
      </div>
      <pre className='bg-surface-2 text-foreground/80 min-h-[232px] overflow-x-auto px-5 py-4 text-[12.5px] leading-relaxed break-words whitespace-pre-wrap'>
        <span className='text-primary whitespace-pre'>
          {buildCommandLines(activeScenario)}
        </span>
        {typedResponse && <span className='text-success'>{typedResponse}</span>}
        <span className='bg-primary animate-terminal-demo-blink ml-0.5 inline-block h-3.5 w-[7px] translate-y-[-0.1em] align-middle' />
      </pre>
    </div>
  )
}
