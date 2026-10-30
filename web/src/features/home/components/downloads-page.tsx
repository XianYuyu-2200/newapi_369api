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

const files = [
  {
    name: 'Windows x64',
    file: 'Codex-Zh-Launcher-Windows-x64.exe',
    size: '168 KB',
  },
  {
    name: 'macOS Intel x64',
    file: 'Codex-Zh-Launcher-macOS-x64.zip',
    size: '385 KB',
  },
  {
    name: 'macOS Silicon ARM64',
    file: 'Codex-Zh-Launcher-macOS-arm64.zip',
    size: '373 KB',
  },
] as const

export function DownloadsPage() {
  return (
    <RelayShell>
      <main className='mx-auto min-h-[70vh] max-w-3xl px-6 py-16'>
        <h1 className='text-4xl font-extrabold text-[#e9eefb]'>资料下载</h1>
        <p className='mt-3 max-w-2xl text-base leading-7 text-[#9fb0cf]'>
          汉化 Codex 使用。按自己的系统下载对应文件。
        </p>
        <ul className='mt-8 grid gap-4'>
          {files.map((item) => (
            <li
              key={item.file}
              className='flex flex-col gap-4 rounded-2xl border border-white/10 bg-[#121b30] p-5 sm:flex-row sm:items-center sm:justify-between'
            >
              <div className='min-w-0'>
                <h2 className='text-xl font-bold text-[#e9eefb]'>{item.name}</h2>
                <p className='mt-2 truncate font-mono text-sm text-[#c5d2ea]'>
                  {item.file}
                </p>
                <p className='mt-1 text-sm text-[#9fb0cf]'>{item.size}</p>
              </div>
              <a
                href={`/downloads/${item.file}`}
                download={item.file}
                className='inline-flex shrink-0 items-center justify-center rounded-full bg-[#2468f2] px-5 py-2 text-sm font-semibold text-white hover:bg-[#4d84ff]'
              >
                下载
              </a>
            </li>
          ))}
        </ul>
      </main>
    </RelayShell>
  )
}
