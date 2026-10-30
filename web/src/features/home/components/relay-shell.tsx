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
import { Link, useRouterState } from '@tanstack/react-router'
import type { CSSProperties, ReactNode } from 'react'

import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'

const shellStyle = {
  '--background': '#0a1120',
  '--card': '#121b30',
  '--popover': '#162033',
  '--muted': '#162033',
  '--secondary': '#162033',
  '--border': 'rgba(148, 163, 184, 0.16)',
  '--input': 'rgba(148, 163, 184, 0.2)',
  '--sidebar': '#0d1730',
} as CSSProperties

const links = [
  { label: '控制台', to: '/dashboard', guestTo: '/sign-in' },
  { label: '模型广场', to: '/pricing' },
  { label: '文档', to: '/docs', blank: true },
  { label: '代理系统简介', href: '/agent/369API-代理系统简介.docx' },
  { label: '资料下载', to: '/downloads' },
  { label: '售后群', to: '/support' },
  { label: 'CCSwitch', href: 'https://ccswitch.io/zh/' },
  { label: 'Codex安装包', to: '/codex' },
] as const

export function RelayShell({
  children,
  trailing,
  showFooter = true,
}: {
  children: ReactNode
  trailing?: ReactNode
  showFooter?: boolean
}) {
  const { systemName } = useSystemConfig()
  const brand = !systemName || systemName === 'New API' ? '369api' : systemName
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const isAuthenticated = !!useAuthStore((state) => state.auth.user)

  return (
    <div
      className='dark flex min-h-svh flex-col bg-[#0a1120] text-[#9fb0cf]'
      style={shellStyle}
    >
      <header className='sticky top-0 z-50 border-b border-white/10 bg-[#0b1220]/90 backdrop-blur'>
        <div className='mx-auto flex h-16 max-w-7xl items-center gap-3 px-4 md:px-6'>
          <Link to='/' className='flex shrink-0 items-center gap-2 text-white'>
            <span className='grid size-8 place-items-center rounded-lg bg-[#2468f2] text-sm font-bold'>
              3
            </span>
            <span className='text-lg font-semibold tracking-tight'>{brand}</span>
          </Link>
          <nav className='flex min-w-0 flex-1 items-center gap-0.5 overflow-x-auto text-sm font-semibold text-[#f9f9f9]'>
            {links.map((link) => {
              const className = 'shrink-0 rounded-md px-2 py-2 hover:text-[#2468f2]'
              if ('href' in link) {
                return (
                  <a
                    key={link.label}
                    href={link.href}
                    target='_blank'
                    rel='noreferrer'
                    className={className}
                  >
                    {link.label}
                  </a>
                )
              }
              const to = link.label === '控制台' && !isAuthenticated ? link.guestTo : link.to
              const active = pathname === to || pathname.startsWith(`${to}/`)
              if ('blank' in link) {
                return (
                  <a
                    key={link.label}
                    href={to}
                    target='_blank'
                    rel='noreferrer'
                    className={cn(className, active && 'text-[#4d84ff]')}
                  >
                    {link.label}
                  </a>
                )
              }
              return (
                <Link
                  key={link.label}
                  to={to}
                  className={cn(className, active && 'text-[#4d84ff]')}
                >
                  {link.label}
                </Link>
              )
            })}
          </nav>
          <div className='ml-auto flex shrink-0 items-center gap-2'>
            {isAuthenticated ? (
              <Link
                to='/dashboard'
                className='rounded-full bg-[#2468f2] px-4 py-1.5 text-sm font-semibold text-white'
              >
                控制台
              </Link>
            ) : (
              <>
                <Link to='/sign-in' className='px-3 py-1.5 text-sm font-semibold text-white'>
                  登录
                </Link>
                <Link
                  to='/sign-up'
                  className='rounded-full bg-[#4d84ff] px-4 py-1.5 text-sm font-semibold text-white'
                >
                  注册
                </Link>
              </>
            )}
            {trailing}
          </div>
        </div>
      </header>
      <div className='flex-1'>{children}</div>
      {showFooter ? (
        <footer className='border-t border-white/10 bg-[#0a1120] px-4 py-8 text-sm md:px-6'>
          <div className='mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-3'>
            <span>{brand}</span>
          </div>
        </footer>
      ) : null}
    </div>
  )
}
