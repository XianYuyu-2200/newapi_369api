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
import { Skeleton } from '@/components/ui/skeleton'
import dayjs from '@/lib/dayjs'
import { formatQuota } from '@/lib/format'

import { getAgentSystemRecords } from '../api'
import type { AgentOrderRecord } from '../types'
import { MetricCard } from './metric'
import { Pager } from './users-panel'

export function RecordsPanel(props: { kind: 'recharge' | 'commission' }) {
  const { t } = useTranslation()
  const [keyword, setKeyword] = useState('')
  const [query, setQuery] = useState('')
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const result = useQuery({
    queryKey: ['agent-system', 'records', props.kind, query, page, pageSize],
    queryFn: () =>
      getAgentSystemRecords({
        kind: props.kind,
        keyword: query,
        page,
        pageSize,
      }),
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
            placeholder={t('Search username, email, or order number')}
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

      {result.isLoading ? <Skeleton className='h-24 rounded-xl' /> : null}
      {data ? (
        <div className='grid gap-3 md:grid-cols-3'>
          <MetricCard
            label={t("Today's recharge commission")}
            value={formatQuota(data.todayCommissionQuota)}
            hint={t('Commission from successful recharges today')}
          />
          <MetricCard
            label={t("Yesterday's recharge commission")}
            value={formatQuota(data.yesterdayCommissionQuota)}
            hint={t('Commission from successful recharges yesterday')}
          />
          <MetricCard
            label={t('Recharge commission in the last 7 days')}
            value={formatQuota(data.weekCommissionQuota)}
            hint={t('Commission from successful recharges in the last 7 days')}
          />
        </div>
      ) : null}

      <section>
        <div className='mb-3 flex items-center justify-between gap-3'>
          <h3 className='text-sm font-medium'>{t('Record details')}</h3>
          {data ? (
            <span className='text-muted-foreground text-xs'>
              {t('{{count}} records', { count: data.total })}
            </span>
          ) : null}
        </div>
        {result.isLoading ? <Skeleton className='h-40 rounded-xl' /> : null}
        {!result.isLoading && result.isError ? (
          <p className='text-sm text-red-600 dark:text-red-400'>
            {t('Failed to load records')}
          </p>
        ) : null}
        {data && data.records.length === 0 ? (
          <div className='text-muted-foreground rounded-lg border border-dashed px-4 py-10 text-center text-sm'>
            {emptyRecordsLabel(props.kind, query, t)}
          </div>
        ) : null}
        {data && data.records.length > 0 ? (
          <div className='space-y-3'>
            {data.records.map((record) => (
              <RecordCard key={record.id} record={record} />
            ))}
          </div>
        ) : null}
      </section>
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

function emptyRecordsLabel(
  kind: 'recharge' | 'commission',
  query: string,
  t: (key: string) => string
) {
  if (query) return t('No records match this search')
  if (kind === 'recharge') return t('No recharge records')
  return t('No commission records')
}

function RecordCard(props: { record: AgentOrderRecord }) {
  const { t } = useTranslation()
  const record = props.record
  const name =
    record.displayName && record.displayName !== record.username
      ? record.displayName
      : record.username
  const stamp =
    record.status === 'success' && record.completedAt > 0
      ? record.completedAt
      : record.createdAt
  const when = stamp > 0 ? dayjs.unix(stamp).format('YYYY-MM-DD HH:mm:ss') : '—'
  const method =
    record.paymentMethod && record.paymentMethod !== 'unknown'
      ? record.paymentMethod
      : t('Unknown')

  return (
    <article className='rounded-lg border p-4'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div className='min-w-0'>
          <div className='font-medium'>{name}</div>
          <div className='text-muted-foreground mt-1 truncate text-sm'>
            @{record.username}
            {record.email ? ` · ${record.email}` : ''}
          </div>
        </div>
        <div className='flex items-center gap-2'>
          <Badge variant='secondary'>
            {record.level === 1 ? t('Level 1') : t('Level 2')}
          </Badge>
          <Badge variant={record.status === 'success' ? 'outline' : 'warning'}>
            {orderStatusLabel(record.status, t)}
          </Badge>
        </div>
      </div>
      <dl className='mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-5'>
        <Field label={t('Order number')} value={record.tradeNo || '—'} />
        <Field label={t('Payment method')} value={method} />
        <Field
          label={t('Recharge amount')}
          value={formatQuota(record.rechargeQuota)}
        />
        <Field
          label={t('Commission amount')}
          value={formatQuota(record.commissionQuota)}
        />
        <Field label={t('Time')} value={when} />
      </dl>
    </article>
  )
}

function orderStatusLabel(status: string, t: (key: string) => string) {
  if (status === 'success') return t('Success')
  if (status === 'pending') return t('Pending')
  if (status === 'failed') return t('Failed')
  if (status === 'expired') return t('Expired')
  return t('Unknown')
}

function Field(props: { label: string; value: string }) {
  return (
    <div>
      <dt className='text-muted-foreground text-xs'>{props.label}</dt>
      <dd className='mt-1 text-sm font-medium break-all'>{props.value}</dd>
    </div>
  )
}
