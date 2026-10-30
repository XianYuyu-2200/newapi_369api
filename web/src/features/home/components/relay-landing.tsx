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

import { RelayShell } from './relay-shell'
import {
  Boxes,
  FileText,
  Lock,
  Megaphone,
  Shield,
  Zap,
} from 'lucide-react'
import { useEffect, useState } from 'react'

import { useSystemConfig } from '@/hooks/use-system-config'
import { cn } from '@/lib/utils'

type Notice = {
  id?: number | string
  content?: string
  extra?: string
  publishDate?: string
}

const advantages = [
  {
    icon: Zap,
    title: '统一入口，少改调用方',
    text: '对话和补全走同一套 OpenAI 兼容路径。换模型时优先改模型名和分组，而不是重写每个客户端。',
  },
  {
    icon: Shield,
    title: '渠道和令牌分开管',
    text: '上游密钥留在后台。调用方只使用自己的令牌，停用、限额和用量也都落在令牌上。',
  },
  {
    icon: Lock,
    title: '访问可以收紧',
    text: '按令牌、模型和分组控制可调用范围。计费、排障和异步任务仍可能保留必要记录。',
  },
  {
    icon: Boxes,
    title: '用量可以核对',
    text: '请求结果、额度和失败原因在控制台查看。对不上账时，先核对令牌和分组，再看渠道。',
  },
]

const features = [
  {
    title: '模型中转',
    text: '已配置的渠道可以汇总到一个站点。实际开通哪些模型，以模型广场和后台渠道为准。',
    visual: '转发',
  },
  {
    title: '额度与分组',
    text: '价格、倍率和可用分组在模型广场查看。这里不写死单价，避免和后台配置不一致。',
    visual: '计费',
  },
  {
    title: '调用记录',
    text: '登录后可在控制台查看请求结果、用量和失败原因，用来排查是令牌、模型还是上游的问题。',
    visual: '日志',
  },
  {
    title: '客户端接入',
    text: '支持把 base URL 指到本站、用令牌调用的客户端。具体兼容范围取决于模型和接口类型。',
    visual: '接入',
  },
]

const scenarios = ['个人调试', '内部工具', '应用接入', '多模型切换', '用量核对', '分组试用']

const fallbackNotices: Notice[] = [
  {
    id: 'notice-1',
    extra: '说明',
    publishDate: '',
    content: '价格、倍率和模型是否可用，以模型广场和后台配置为准。',
  },
  {
    id: 'local-2',
    publishDate: '',
    content: '模型是否可用、单价和分组倍率，请打开模型广场核对当前环境。',
  },
  {
    id: 'local-3',
    publishDate: '',
    content: '创建一个令牌后，按 OpenAI 兼容方式请求 /v1。不要把上游渠道密钥写进客户端。',
  },
]

function formatDate(value?: string) {
  if (!value) return '公告'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export function RelayLanding({ isAuthenticated }: { isAuthenticated: boolean }) {
  const { systemName } = useSystemConfig()
  const brand = !systemName || systemName === 'New API' ? '369api' : systemName
  const [notices, setNotices] = useState<Notice[]>(fallbackNotices)
  const startTo = isAuthenticated ? '/dashboard' : '/sign-up'

  useEffect(() => {
    let cancelled = false
    fetch('/api/status')
      .then((response) => response.json())
      .then((payload) => {
        const list = payload?.data?.announcements
        if (!cancelled && Array.isArray(list) && list.length > 0) {
          setNotices(list)
        }
      })
      .catch(() => {
        // Keep the local design notes when status is unavailable.
      })
    return () => {
      cancelled = true
    }
  }, [])

  const [pinned, ...rest] = notices

  return (
    <RelayShell>
      <main id='top'>
        <section
          className='relative overflow-hidden px-4 pt-16 pb-14 md:px-6 md:pt-24 md:pb-16'
          style={{
            backgroundImage:
              'radial-gradient(820px 420px at 88% -4%, rgba(56,116,255,0.22), transparent 64%), radial-gradient(640px 360px at 4% 24%, rgba(56,116,255,0.12), transparent 62%), linear-gradient(rgb(13,23,48), rgb(11,19,34) 58%, rgb(10,17,32))',
          }}
        >
          <div className='mx-auto grid max-w-7xl items-start gap-10 lg:grid-cols-[minmax(0,1.05fr)_minmax(320px,460px)]'>
            <div>
              <h1 className='text-5xl font-extrabold tracking-tight text-[#2468f2] md:text-6xl'>
                {brand}
              </h1>
              <h2 className='mt-3 max-w-xl text-4xl leading-tight font-extrabold text-[#e9eefb] md:text-[42px]'>
                企业级 AI 接口中转平台
              </h2>
              <p className='mt-6 text-lg font-semibold text-[#e9eefb]'>
                统一调用入口 · 令牌和渠道分开 · 价格以后台为准
              </p>
              <p className='mt-5 max-w-xl text-base leading-7'>
                {brand} 把已接入的模型放到同一个 OpenAI 兼容接口后面。开通情况和价格请前往{' '}
                <Link to='/pricing' className='text-[#4d84ff] underline-offset-4 hover:underline'>
                  模型广场
                </Link>{' '}
                查看，不要以首页文案当作报价。
              </p>
              <div className='mt-8 flex flex-wrap gap-3'>
                <Link
                  to={startTo}
                  className='inline-flex h-12 items-center rounded-[10px] bg-gradient-to-br from-[#3b82ff] to-[#1a5dee] px-9 text-base font-semibold text-white shadow-[0_12px_26px_-10px_rgba(36,104,242,0.65)]'
                >
                  {isAuthenticated ? '进入控制台' : '免费开通使用'}
                </Link>
                <a
                  href='#about'
                  className='inline-flex h-12 items-center gap-2 rounded-[10px] border border-[#2468f2]/50 bg-[#121b30] px-7 text-base font-semibold text-[#2468f2]'
                >
                  <FileText className='size-4' />
                  查看使用说明
                </a>
              </div>
              <div className='mt-8 max-w-xl rounded-xl border border-white/10 bg-[#10192e] px-4 py-3'>
                <div className='text-xs tracking-wide text-[#7f93b5]'>OpenAI 兼容地址</div>
                <div className='mt-1 font-mono text-sm text-[#e9eefb]'>https://369api.top/v1</div>
              </div>
            </div>

            <section className='rounded-[18px] border border-slate-400/15 bg-[rgba(19,28,50,0.82)] shadow-[0_18px_40px_-22px_rgba(0,0,0,0.6)]'>
              <div className='flex items-center gap-3 border-b border-white/10 px-5 py-4 text-[#e9eefb]'>
                <span className='grid size-9 place-items-center rounded-lg bg-[#24304a] text-[#8eb4ff]'>
                  <Megaphone className='size-4' />
                </span>
                <h2 className='text-lg font-semibold'>公告栏</h2>
              </div>
              {pinned ? (
                <div className='m-4 rounded-[10px] bg-[rgba(77,132,255,0.16)] px-4 py-3.5 text-[#e9eefb]'>
                  <div className='mb-2 text-sm text-[#8eb4ff]'>{pinned.extra || '置顶'}</div>
                  <p className='text-sm leading-6'>{pinned.content}</p>
                </div>
              ) : null}
              <div className='max-h-[420px] space-y-4 overflow-auto px-5 pb-5'>
                {rest.map((notice) => (
                  <article key={notice.id ?? notice.content} className='border-b border-white/10 pb-4 last:border-b-0'>
                    <div className='text-sm text-[#7f93b5]'>{formatDate(notice.publishDate)}</div>
                    <p className='mt-1 text-sm leading-6 text-[#d5def2]'>{notice.content}</p>
                  </article>
                ))}
              </div>
            </section>
          </div>
        </section>

        <section id='about' className='bg-[#0a1120] px-4 py-16 md:px-6 md:py-20'>
          <div className='mx-auto grid max-w-7xl items-center gap-12 lg:grid-cols-[280px_minmax(0,1fr)]'>
            <div className='relative mx-auto grid size-56 place-items-center'>
              <div className='absolute inset-3 rounded-full border border-[#2468f2]/30' />
              <div className='absolute inset-8 rounded-full border border-[#2468f2]/20' />
              <div className='relative grid size-36 place-items-center rounded-full bg-[#162445] text-center shadow-[0_0_40px_rgba(36,104,242,0.25)]'>
                <Shield className='size-10 text-[#4d84ff]' />
                <div className='mt-1 text-sm font-semibold text-white'>{brand}</div>
              </div>
            </div>
            <div>
              <h2 className='text-3xl font-bold text-[#e9eefb]'>使用前先看这几条</h2>
              <p className='mt-4 max-w-3xl leading-7'>
                中转负责统一入口和计费，不替代上游模型自己的使用规则。代码、账号和未公开数据发出去之前，先确认能否提交。
              </p>
              <ol className='mt-6 space-y-3 text-sm leading-6'>
                <li>1. 令牌只发给需要调用的人。上游渠道密钥留在后台，不写进客户端配置。</li>
                <li>2. 请求会转发到所选模型的上游。不要把中转理解成所有环节都不留记录。</li>
                <li>3. 价格、倍率和可用分组以模型广场当前数据为准，首页不维护另一份价目表。</li>
                <li>4. 需要保密的代码和账号不要直接提交。能在本地处理的内容，先处理再调用。</li>
              </ol>
            </div>
          </div>
        </section>

        <section className='bg-[#0a1120] px-4 py-4 pb-16 md:px-6'>
          <div className='mx-auto max-w-7xl'>
            <h2 className='text-center text-3xl font-bold text-[#e9eefb]'>核心优势</h2>
            <p className='mx-auto mt-3 max-w-2xl text-center'>
              一个入口管理令牌、模型和用量，不用为每个客户端单独保存一套上游密钥。
            </p>
            <div className='mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4'>
              {advantages.map((item) => (
                <article key={item.title} className='rounded-[14px] border border-slate-400/15 bg-[#121b30] px-6 py-7'>
                  <item.icon className='size-6 text-[#4d84ff]' />
                  <h3 className='mt-5 text-lg font-semibold text-[#e9eefb]'>{item.title}</h3>
                  <p className='mt-3 text-sm leading-6'>{item.text}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section id='features' className='bg-[#0a1120] px-4 py-8 md:px-6 md:py-12'>
          <div className='mx-auto max-w-7xl'>
            <h2 className='text-center text-3xl font-bold text-[#e9eefb]'>站点实际能做的事</h2>
            <div className='mt-10 grid gap-4 md:grid-cols-2'>
              {features.map((item) => (
                <article key={item.title} className='rounded-[14px] border border-slate-400/15 bg-[#121b30] p-4 pb-6'>
                  <div className='mb-5 flex h-28 flex-col justify-center gap-2 rounded-xl border border-white/10 bg-[#0d1730] px-5'>
                    <div className='flex items-center justify-between text-xs text-[#7f93b5]'>
                      <span>示意</span>
                      <span className='rounded-md bg-[#2468f2]/15 px-2 py-1 text-[#8eb4ff]'>{item.visual}</span>
                    </div>
                    <div className='h-2 rounded-full bg-white/10' />
                    <div className='h-2 w-2/3 rounded-full bg-[#2468f2]/40' />
                  </div>
                  <h3 className='text-xl font-semibold text-[#e9eefb]'>{item.title}</h3>
                  <p className='mt-3 text-sm leading-6'>{item.text}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className='bg-[#0a1120] px-4 py-16 md:px-6'>
          <div className='mx-auto max-w-7xl text-center'>
            <h2 className='text-3xl font-bold text-[#e9eefb]'>适合先在本地确认的场景</h2>
            <div className='mt-8 flex flex-wrap justify-center gap-3'>
              {scenarios.map((item) => (
                <span key={item} className='rounded-full border border-white/10 bg-[#121b30] px-4 py-2 text-sm text-[#d5def2]'>
                  {item}
                </span>
              ))}
            </div>
            <p className='mx-auto mt-6 max-w-3xl leading-7'>
              先用本机确认导航、模型和调用说明，再决定哪一版给用户。用户正在用的站点保持不动。
            </p>
          </div>
        </section>

        <section
          className='px-4 py-20 text-center md:px-6 md:py-24'
          style={{ backgroundImage: 'linear-gradient(rgb(12,21,38), rgb(14,28,56))' }}
        >
          <h2 className='text-3xl font-bold text-[#e9eefb] md:text-4xl'>
            用 {brand} 接入已开通的模型
          </h2>
          <p className='mt-4 text-lg'>注册后创建令牌，价格和分组到模型广场确认</p>
          <Link
            to={startTo}
            className={cn(
              'mt-8 inline-flex h-12 items-center rounded-[10px] bg-gradient-to-br from-[#3b82ff] to-[#1a5dee] px-9 font-semibold text-white'
            )}
          >
            {isAuthenticated ? '进入控制台' : '注册并创建令牌'}
          </Link>
        </section>
      </main>

    </RelayShell>
  )
}
