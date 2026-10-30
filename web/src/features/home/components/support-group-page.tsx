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
import wechatGroupQrCode from '../../../../public/support/wechat-group.jpg'
import { RelayShell } from './relay-shell'

export function SupportGroupPage() {
  return (
    <RelayShell>
      <main className='mx-auto flex min-h-[70vh] max-w-3xl flex-col items-center px-6 py-16 text-center'>
        <h1 className='text-4xl font-extrabold text-[#e9eefb]'>售后群</h1>
        <p className='mt-3 text-base text-[#9fb0cf]'>
          微信扫码加入 369API售后服务群3群
        </p>
        <img
          src={wechatGroupQrCode}
          alt='369API售后服务群3群微信二维码，7天内（10月31日前）有效'
          className='mt-8 w-full max-w-[420px] rounded-2xl'
        />
      </main>
    </RelayShell>
  )
}
