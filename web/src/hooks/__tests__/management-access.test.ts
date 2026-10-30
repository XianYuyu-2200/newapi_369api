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
import { afterEach, describe, expect, it } from 'vitest'

import { Route as ChannelsRoute } from '@/routes/_authenticated/channels/index'
import { Route as ModelSectionRoute } from '@/routes/_authenticated/models/$section'
import { Route as ModelsRoute } from '@/routes/_authenticated/models/index'
import { useAuthStore } from '@/stores/auth-store'

afterEach(() => useAuthStore.getState().auth.reset())

describe('channel and model page access', () => {
  const pages = [ModelsRoute, ModelSectionRoute]
  it.each([undefined, 1, 10])(
    'role %s cannot bypass navigation via a direct URL',
    (role) => {
      if (role !== undefined) {
        useAuthStore
          .getState()
          .auth.setUser({ id: 1, username: 'manager', role })
      }
      for (const page of pages) {
        const guard = page.options.beforeLoad as (context: {
          params: { section: string }
        }) => unknown
        expect(() => guard({ params: { section: 'metadata' } })).toThrow(
          expect.objectContaining({
            options: expect.objectContaining({ to: '/403' }),
          })
        )
      }
    }
  )
  it.each([undefined, 1])('role %s cannot open channels', (role) => {
    if (role !== undefined) {
      useAuthStore.getState().auth.setUser({ id: 1, username: 'user', role })
    }
    const guard = ChannelsRoute.options.beforeLoad as () => unknown
    expect(() => guard()).toThrow(
      expect.objectContaining({
        options: expect.objectContaining({ to: '/403' }),
      })
    )
  })
  it('an admin can open channels', () => {
    useAuthStore.getState().auth.setUser({ id: 1, username: 'admin', role: 10 })
    const guard = ChannelsRoute.options.beforeLoad as () => unknown
    expect(() => guard()).not.toThrow()
  })
  it('root can open channels and model sections', () => {
    useAuthStore.getState().auth.setUser({ id: 1, username: 'root', role: 100 })
    const channelGuard = ChannelsRoute.options.beforeLoad as () => unknown
    const sectionGuard = ModelSectionRoute.options.beforeLoad as (context: {
      params: { section: string }
    }) => unknown
    expect(() => channelGuard()).not.toThrow()
    for (const section of ['metadata', 'vendors', 'deployments']) {
      expect(() => sectionGuard({ params: { section } })).not.toThrow()
    }
    const indexGuard = ModelsRoute.options.beforeLoad as () => unknown
    expect(() => indexGuard()).toThrow(
      expect.objectContaining({
        options: expect.objectContaining({ to: '/models/$section' }),
      })
    )
  })
})
