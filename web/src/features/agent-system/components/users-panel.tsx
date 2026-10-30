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

import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Search } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { NativeSelect, NativeSelectOption } from '@/components/ui/native-select'
import { Skeleton } from '@/components/ui/skeleton'
import dayjs from '@/lib/dayjs'
import { formatQuota } from '@/lib/format'

import { getAgentSystemUsers } from '../api'
import type { AgentInvitee } from '../types'

const pageSizes = [10, 20, 50]

export function UsersPanel() {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState('')
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const result = useQuery({
    queryKey: ['agent-system', 'users', query, page, pageSize],
    queryFn: () => getAgentSystemUsers({ keyword: query, page, pageSize }),
    placeholderData: keepPreviousData,
  })
  const data = result.data
  const pageCount = Math.max(1, Math.ceil((data?.total ?? 0) / pageSize))

  return (
    <div className='space-y-4'>
      <form
        className='flex flex-wrap items-center gap-2'
        onSubmit={(event) => {
          event.preventDefault()
          setQuery(keyword.trim())
          setPage(1)
        }}
      >
        <div className='relative min-w-[16rem] flex-1'>
          <Search className='text-muted-foreground pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2' />
          <Input
            value={keyword}
            className='pl-8'
            maxLength={64}
            placeholder={t('Search username, nickname, or email')}
            onChange={(event) => setKeyword(event.target.value)}
          />
        </div>
        <Button type='submit'>{t('Search')}</Button>
        <Button
          type='button'
          variant='outline'
          onClick={() => {
            setKeyword('')
            setQuery('')
            setPage(1)
          }}
        >
          {t('Reset')}
        </Button>
      </form>

      {data ? (
        <p className='text-muted-foreground text-sm'>
          {t('{{count}} invited users', { count: data.total })}
        </p>
      ) : null}

      {result.isLoading ? <Skeleton className='h-64 rounded-xl' /> : null}
      {!result.isLoading && result.isError ? (
        <p className='text-sm text-red-600 dark:text-red-400'>
          {t('Failed to load invited users')}
        </p>
      ) : null}
      {data && data.users.length === 0 ? (
        <div className='text-muted-foreground rounded-lg border border-dashed px-4 py-10 text-center text-sm'>
          {query
            ? t('No invited users match this search')
            : t('No invited users yet')}
        </div>
      ) : null}
      {data && data.users.length > 0 ? (
        <div className='space-y-3'>
          {data.users.map((user) => (
            <UserCard key={user.id} user={user} />
          ))}
        </div>
      ) : null}

      <Pager
        page={Math.min(page, pageCount)}
        pageCount={pageCount}
        pageSize={pageSize}
        onPage={setPage}
        onPageSize={(size) => {
          setPageSize(size)
          setPage(1)
        }}
      />
    </div>
  )
}

function UserCard(props: { user: AgentInvitee }) {
  const { t } = useTranslation()
  const user = props.user
  const name =
    user.displayName && user.displayName !== user.username
      ? user.displayName
      : user.username
  const registered =
    user.createdAt > 0
      ? dayjs.unix(user.createdAt).format('YYYY-MM-DD HH:mm:ss')
      : '—'

  return (
    <article className='rounded-lg border p-4'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div className='min-w-0'>
          <div className='font-medium'>{name}</div>
          <div className='text-muted-foreground mt-1 truncate text-sm'>
            @{user.username}
            {user.email ? ` · ${user.email}` : ''}
          </div>
        </div>
        <div className='flex items-center gap-2'>
          <Badge variant='secondary'>
            {user.level === 1 ? t('Level 1') : t('Level 2')}
          </Badge>
          <Badge variant={user.status === 1 ? 'outline' : 'warning'}>
            {statusLabel(user.status, t)}
          </Badge>
        </div>
      </div>
      <dl className='mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
        <Field
          label={t('Successful recharge count')}
          value={String(user.rechargeCount ?? 0)}
        />
        <Field
          label={t('Recharge amount')}
          value={formatQuota(user.rechargeQuota ?? 0)}
        />
        <Field
          label={t('Contributed commission')}
          value={formatQuota(user.commissionQuota ?? 0)}
        />
        <Field label={t('Registration time')} value={registered} />
      </dl>
    </article>
  )
}

function statusLabel(status: number, t: (key: string) => string) {
  if (status === 1) return t('Enabled')
  if (status === 2) return t('Disabled')
  return t('Unknown')
}

function Field(props: { label: string; value: string; hint?: string }) {
  return (
    <div>
      <dt className='text-muted-foreground text-xs'>{props.label}</dt>
      <dd className='mt-1 text-sm font-medium'>{props.value}</dd>
      {props.hint ? (
        <p className='text-muted-foreground mt-1 text-xs'>{props.hint}</p>
      ) : null}
    </div>
  )
}

export function Pager(props: {
  page: number
  pageCount: number
  pageSize: number
  onPage: (page: number) => void
  onPageSize: (size: number) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='flex flex-wrap items-center justify-end gap-2'>
      <Button
        type='button'
        size='sm'
        variant='outline'
        disabled={props.page <= 1}
        onClick={() => props.onPage(props.page - 1)}
      >
        {t('Previous page')}
      </Button>
      <span className='text-sm tabular-nums'>{props.page}</span>
      <Button
        type='button'
        size='sm'
        variant='outline'
        disabled={props.page >= props.pageCount}
        onClick={() => props.onPage(props.page + 1)}
      >
        {t('Next page')}
      </Button>
      <NativeSelect
        size='sm'
        value={String(props.pageSize)}
        aria-label={t('Rows per page')}
        onChange={(event) => props.onPageSize(Number(event.target.value))}
      >
        {pageSizes.map((size) => (
          <NativeSelectOption key={size} value={String(size)}>
            {t('{{count}} per page', { count: size })}
          </NativeSelectOption>
        ))}
      </NativeSelect>
    </div>
  )
}
