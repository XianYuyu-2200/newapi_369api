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
import { RelayShell } from './relay-shell'

const sources = [
  {
    name: '百度网盘',
    action: '打开百度网盘',
    packages: [
      {
        name: 'Windows',
        href: 'https://pan.baidu.com/s/1A7lcsepiHePENOoW_8Yxtw?pwd=e6h2',
        code: 'e6h2',
      },
      {
        name: 'Mac Intel',
        href: 'https://pan.baidu.com/s/14b65p5buRGRmAeV6oFoeRA?pwd=hgnn',
        code: 'hgnn',
      },
      {
        name: 'Mac Silicon ARM64',
        href: 'https://pan.baidu.com/s/1QmUB5l4OP5K8PXboqbaagw?pwd=f6c2',
        code: 'f6c2',
      },
    ],
  },
  {
    name: '夸克网盘',
    action: '打开夸克网盘',
    packages: [
      {
        name: 'Windows',
        href: 'https://pan.quark.cn/s/8748a87fe9db',
      },
      {
        name: 'Mac Intel',
        href: 'https://pan.quark.cn/s/a7446494d1c9?pwd=dtti',
        code: 'dtti',
      },
      {
        name: 'Mac Silicon ARM64',
        href: 'https://pan.quark.cn/s/2741e905e0c9?pwd=hWKb',
        code: 'hWKb',
      },
    ],
  },
] as const

export function CodexDownloadsPage() {
  return (
    <RelayShell>
      <main className='mx-auto min-h-[70vh] max-w-3xl px-6 py-16'>
        <h1 className='text-4xl font-extrabold text-[#e9eefb]'>Codex安装包</h1>
        <p className='mt-3 max-w-2xl text-base leading-7 text-[#9fb0cf]'>
          按自己的系统选择版本。百度网盘和夸克网盘是两套链接。卡片上写了提取码的，打开后如果还要求填写，就用这个提取码。
        </p>
        <div className='mt-10 grid gap-10'>
          {sources.map((source) => (
            <section key={source.name}>
              <h2 className='text-2xl font-bold text-[#e9eefb]'>{source.name}</h2>
              <ul className='mt-4 grid gap-4'>
                {source.packages.map((item) => (
                  <li
                    key={item.name}
                    className='flex flex-col gap-4 rounded-2xl border border-white/10 bg-[#121b30] p-5 sm:flex-row sm:items-center sm:justify-between'
                  >
                    <div>
                      <h3 className='text-xl font-bold text-[#e9eefb]'>{item.name}</h3>
                      {'code' in item && item.code ? (
                        <p className='mt-2 text-sm text-[#c5d2ea]'>
                          提取码{' '}
                          <span className='font-mono text-base tracking-wide text-white'>
                            {item.code}
                          </span>
                        </p>
                      ) : null}
                    </div>
                    <a
                      href={item.href}
                      target='_blank'
                      rel='noreferrer'
                      className='inline-flex shrink-0 items-center justify-center rounded-full bg-[#2468f2] px-5 py-2 text-sm font-semibold text-white hover:bg-[#4d84ff]'
                    >
                      {source.action}
                    </a>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </div>
      </main>
    </RelayShell>
  )
}
