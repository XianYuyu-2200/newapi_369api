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
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { useAuthStore } from '@/stores/auth-store'

import { CommandMenu } from '../command-menu'

vi.mock('@tanstack/react-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@tanstack/react-router')>()
  return {
    ...actual,
    useLocation: () => ({ pathname: '/dashboard' }),
    useNavigate: () => vi.fn(),
  }
})
vi.mock('@/context/search-provider', () => ({
  useSearch: () => ({ open: true, setOpen: vi.fn() }),
}))
vi.mock('@/context/theme-provider', () => ({
  useTheme: () => ({ setTheme: vi.fn() }),
}))

afterEach(() => {
  cleanup()
  useAuthStore.getState().auth.reset()
})

it.each([10, 100])(
  'command search obeys management access for role %i',
  (role) => {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    client.setQueryData(['status'], {})
    useAuthStore.getState().auth.setUser({ id: 1, username: 'manager', role })
    render(
      <QueryClientProvider client={client}>
        <CommandMenu />
      </QueryClientProvider>
    )
    if (role === 100) {
      expect(screen.getByText('Channels')).toBeInTheDocument()
      expect(screen.getByText('Models')).toBeInTheDocument()
    } else {
      expect(screen.queryByText('Channels')).not.toBeInTheDocument()
      expect(screen.queryByText('Models')).not.toBeInTheDocument()
    }
    expect(screen.getByText('Users')).toBeInTheDocument()
  }
)
