import { cleanup, render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { afterEach, describe, expect, test, vi } from 'vitest'

import wechatGroupQrCode from '../../../../public/support/wechat-group.jpg'
import { SupportGroupPage } from '../components/support-group-page'

vi.mock('../components/relay-shell', () => ({
  RelayShell: (props: { children: ReactNode }) => props.children,
}))

afterEach(cleanup)

describe('SupportGroupPage', () => {
  test('uses the bundled QR asset instead of the stale public URL', () => {
    render(<SupportGroupPage />)

    const qrCode = screen.getByRole('img', {
      name: /369API售后服务群3群微信二维码/,
    })

    expect(qrCode).toHaveAttribute('src', wechatGroupQrCode)
    expect(qrCode).not.toHaveAttribute('src', '/support/wechat-group.jpg')
  })
})
