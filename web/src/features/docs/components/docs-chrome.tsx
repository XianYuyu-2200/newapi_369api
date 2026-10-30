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

import { useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { Link } from '@tanstack/react-router'
import { ArrowUpRight, Moon, Search, Sun } from 'lucide-react'

import { cn } from '@/lib/utils'
import '@/styles/docs-site.css'

import {
  DocsChromeContext,
  useDocsChrome,
  type DocsTheme,
} from '../hooks/use-docs-chrome'
import { searchDocs, type DocsTarget } from '../lib/content'

const THEME_KEY = '369-docs-theme'

export function DocsChromeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<DocsTheme>(() => {
    if (typeof window === 'undefined') return 'dark'
    return window.localStorage.getItem(THEME_KEY) === 'light' ? 'light' : 'dark'
  })
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (!(event.metaKey || event.ctrlKey) || event.key.toLowerCase() !== 'k') {
        return
      }
      event.preventDefault()
      const inputs = [
        ...document.querySelectorAll<HTMLInputElement>('[data-docs-search]'),
      ]
      inputs.find((input) => input.getClientRects().length > 0)?.focus()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

  const value = useMemo(
    () => ({
      menuOpen,
      setMenuOpen,
      theme,
      toggleTheme: () => {
        setTheme((current) => {
          const next = current === 'dark' ? 'light' : 'dark'
          window.localStorage.setItem(THEME_KEY, next)
          return next
        })
      },
    }),
    [menuOpen, theme]
  )

  return (
    <DocsChromeContext.Provider value={value}>
      <div className='docs-site' data-theme={theme}>
        <a
          href='#docs-content'
          className='sr-only focus:not-sr-only focus:absolute focus:top-3 focus:left-3 focus:z-50 focus:rounded-md focus:bg-[var(--docs-accent-bg)] focus:px-3 focus:py-2 focus:text-white'
        >
          跳到正文
        </a>
        {children}
      </div>
    </DocsChromeContext.Provider>
  )
}

export function DocsAnchor({
  children,
  className,
  onClick,
  target,
}: {
  children: ReactNode
  className?: string
  onClick?: () => void
  target: DocsTarget
}) {
  if (target.type === 'external') {
    return (
      <a
        href={target.href}
        target='_blank'
        rel='noreferrer'
        className={className}
        onClick={onClick}
      >
        {children}
      </a>
    )
  }
  if (target.type === 'article') {
    return (
      <Link
        to='/docs/$section/$article'
        params={{ article: target.article, section: target.section }}
        className={className}
        onClick={onClick}
      >
        {children}
      </Link>
    )
  }
  if (target.type === 'section') {
    return (
      <Link
        to='/docs/$section'
        params={{ section: target.section }}
        className={className}
        onClick={onClick}
      >
        {children}
      </Link>
    )
  }
  return (
    <Link to={target.to} className={className} onClick={onClick}>
      {children}
    </Link>
  )
}

export function DocsLogo({ compact = false }: { compact?: boolean }) {
  return (
    <Link to='/docs' className='flex min-w-0 items-center gap-2.5 text-[var(--docs-text)]'>
      <span className='grid size-9 shrink-0 place-items-center rounded-xl bg-gradient-to-br from-pink-500 via-violet-500 to-sky-400 text-sm font-bold text-white'>
        3
      </span>
      {compact ? null : (
        <span className='min-w-0'>
          <span className='block text-sm leading-none font-semibold'>369api Docs</span>
          <span className='mt-1 block text-[10px] tracking-[0.14em] text-[var(--docs-faint)]'>
            API DOCUMENTATION
          </span>
        </span>
      )}
    </Link>
  )
}

export function DocsSearch({ className }: { className?: string }) {
  const [query, setQuery] = useState('')
  const [open, setOpen] = useState(false)
  const rootRef = useRef<HTMLDivElement>(null)
  const results = useMemo(() => searchDocs(query), [query])

  useEffect(() => {
    const onPointerDown = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false)
    }
    window.addEventListener('pointerdown', onPointerDown)
    return () => window.removeEventListener('pointerdown', onPointerDown)
  }, [])

  return (
    <div ref={rootRef} className={cn('relative', className)}>
      <label className='flex h-10 items-center gap-2 rounded-xl border border-[var(--docs-line)] bg-[var(--docs-input)] px-3 text-sm'>
        <Search className='size-4 shrink-0 text-[var(--docs-faint)]' aria-hidden='true' />
        <span className='sr-only'>搜索文档</span>
        <input
          data-docs-search=''
          value={query}
          onChange={(event) => {
            setQuery(event.target.value)
            setOpen(true)
          }}
          onFocus={() => setOpen(true)}
          placeholder='搜索文档'
          className='min-w-0 flex-1 bg-transparent text-[var(--docs-text)] outline-none placeholder:text-[var(--docs-faint)]'
        />
        <kbd className='hidden rounded-md border border-[var(--docs-line)] px-1.5 py-0.5 text-[10px] text-[var(--docs-faint)] sm:inline'>
          ⌘K
        </kbd>
      </label>
      {open && query.trim() ? (
        <div className='absolute top-[calc(100%+8px)] right-0 left-0 z-50 overflow-hidden rounded-xl border border-[var(--docs-line)] bg-[var(--docs-card)] shadow-2xl'>
          {results.length === 0 ? (
            <p className='px-3 py-3 text-sm text-[var(--docs-muted)]'>没有匹配的文档</p>
          ) : (
            <ul>
              {results.map((hit) => (
                <li key={`${hit.section.id}-${hit.article.slug}`}>
                  <DocsAnchor
                    target={{
                      type: 'article',
                      section: hit.section.id,
                      article: hit.article.slug,
                    }}
                    className='block px-3 py-2.5 hover:bg-[var(--docs-hover)]'
                    onClick={() => {
                      setQuery('')
                      setOpen(false)
                    }}
                  >
                    <span className='block text-sm font-medium'>{hit.article.title}</span>
                    <span className='mt-0.5 block text-xs text-[var(--docs-faint)]'>
                      {hit.section.title} / {hit.group.title}
                    </span>
                  </DocsAnchor>
                </li>
              ))}
            </ul>
          )}
        </div>
      ) : null}
    </div>
  )
}

export function DocsThemeButton() {
  const { theme, toggleTheme } = useDocsChrome()
  const label = theme === 'dark' ? '切换到浅色' : '切换到深色'
  return (
    <button
      type='button'
      className='grid size-10 place-items-center rounded-xl border border-[var(--docs-line)] bg-[var(--docs-secondary)] text-[var(--docs-text)]'
      aria-label={label}
      onClick={toggleTheme}
    >
      {theme === 'dark' ? (
        <Sun className='size-4' aria-hidden='true' />
      ) : (
        <Moon className='size-4' aria-hidden='true' />
      )}
    </button>
  )
}

export function DocsExternalIcon() {
  return <ArrowUpRight className='size-4' aria-hidden='true' />
}
