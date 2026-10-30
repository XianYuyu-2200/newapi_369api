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

import { useQuery } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import dayjs from '@/lib/dayjs'
import { cn } from '@/lib/utils'

import { getGroupStatus } from './api'
import type {
  GroupStatusBucket,
  GroupStatusGroup,
  GroupStatusReport,
} from './types'

const badgeColors = [
  'bg-violet-100 text-violet-700 dark:bg-violet-500/15 dark:text-violet-300',
  'bg-sky-100 text-sky-700 dark:bg-sky-500/15 dark:text-sky-300',
  'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-300',
  'bg-amber-100 text-amber-700 dark:bg-amber-500/15 dark:text-amber-300',
  'bg-rose-100 text-rose-700 dark:bg-rose-500/15 dark:text-rose-300',
]

function formatRatio(ratio: number) {
  if (!Number.isFinite(ratio)) return '—'
  return `${Number(ratio.toFixed(4))}x`
}

function formatPercent(value: number | null) {
  if (value == null || !Number.isFinite(value)) return '—'
  return `${(value * 100).toFixed(2)}%`
}

function formatSeconds(value: number | null) {
  if (value == null || !Number.isFinite(value)) return '—'
  return `${value.toFixed(2)}s`
}

function percentClass(value: number | null) {
  if (value == null || !Number.isFinite(value)) return 'text-muted-foreground'
  if (value >= 0.99) return 'text-emerald-600 dark:text-emerald-400'
  if (value >= 0.95) return 'text-amber-600 dark:text-amber-400'
  return 'text-red-600 dark:text-red-400'
}

function badgeClass(name: string) {
  let hash = 0
  for (const char of name) hash += char.codePointAt(0) ?? 0
  return badgeColors[hash % badgeColors.length]
}

function barClass(bucket: GroupStatusBucket) {
  if (bucket.success === 0 && bucket.error === 0) {
    return 'bg-muted-foreground/25'
  }
  if (bucket.error === 0) return 'bg-emerald-500'
  if (bucket.success === 0) return 'bg-red-500'
  return 'bg-amber-500'
}

function bucketMinute(generatedAt: number, index: number, count: number) {
  const currentMinute = Math.floor(generatedAt / 60) * 60
  return currentMinute - (count - 1 - index) * 60
}

export function GroupStatusPage() {
  const { t } = useTranslation()
  const [view, setView] = useState<'compact' | 'detailed'>('compact')
  const query = useQuery({
    queryKey: ['group-status'],
    queryFn: getGroupStatus,
    refetchInterval: 60_000,
  })
  const report = query.data
  const updatedAt = report
    ? dayjs.unix(report.generatedAt).format('HH:mm:ss')
    : null

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Group status')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <Button
          type='button'
          variant='outline'
          size='sm'
          onClick={() => void query.refetch()}
          disabled={query.isFetching}
        >
          <RefreshCw className={cn('size-4', query.isFetching && 'animate-spin')} />
          {t('Refresh')}
        </Button>
        <span className='text-muted-foreground text-xs'>
          {updatedAt ? t('Updated at {{time}}', { time: updatedAt }) : '—'}
        </span>
        <div className='bg-muted inline-flex rounded-lg p-1'>
          <ViewButton
            active={view === 'compact'}
            onClick={() => setView('compact')}
            label={t('Simple view')}
          />
          <ViewButton
            active={view === 'detailed'}
            onClick={() => setView('detailed')}
            label={t('Detailed view')}
          />
        </div>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='space-y-4'>
          <div className='text-muted-foreground space-y-1 text-sm'>
            <p>{t('View recent availability and response time for each group.')}</p>
            <p>
              {t(
                'These figures come from real requests in the last 30 minutes. A dash means there were no requests, not that the group is healthy.'
              )}
            </p>
            <p>
              {t(
                "Counts include every user's requests in that group and never include request content."
              )}
            </p>
          </div>
          <StatusBody
            loading={query.isLoading}
            error={query.isError && !report}
            report={report}
            detailed={view === 'detailed'}
            onRetry={() => void query.refetch()}
          />
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}


function StatusBody(props: {
  loading: boolean
  error: boolean
  report?: GroupStatusReport
  detailed: boolean
  onRetry: () => void
}) {
  const { t } = useTranslation()
  if (props.loading) return <StatusSkeleton />
  if (props.error || !props.report) {
    return (
      <div className='bg-card rounded-xl border p-6 text-sm'>
        <p>{t('Failed to load group status')}</p>
        <Button
          type='button'
          variant='outline'
          size='sm'
          className='mt-3'
          onClick={props.onRetry}
        >
          {t('Refresh')}
        </Button>
      </div>
    )
  }
  if (props.report.groups.length === 0) {
    return (
      <div className='bg-card text-muted-foreground rounded-xl border p-6 text-sm'>
        {t('No groups are available for this account')}
      </div>
    )
  }
  const report = props.report
  return (
    <div className='grid grid-cols-1 gap-4 xl:grid-cols-2 2xl:grid-cols-3'>
      {report.groups.map((group) => (
        <GroupCard
          key={group.name}
          group={group}
          detailed={props.detailed}
          generatedAt={report.generatedAt}
        />
      ))}
    </div>
  )
}

function ViewButton(props: {
  active: boolean
  label: string
  onClick: () => void
}) {
  return (
    <button
      type='button'
      onClick={props.onClick}
      className={cn(
        'rounded-md px-3 py-1 text-sm',
        props.active
          ? 'bg-background text-foreground shadow-sm'
          : 'text-muted-foreground'
      )}
    >
      {props.label}
    </button>
  )
}

function GroupCard(props: {
  group: GroupStatusGroup
  detailed: boolean
  generatedAt: number
}) {
  const { t } = useTranslation()
  const { group } = props
  const title =
    group.description && group.description !== group.name
      ? t(group.description)
      : group.name
  const initial = title.trim().charAt(0).toUpperCase() || 'G'

  return (
    <article className='bg-card flex flex-col gap-4 rounded-xl border p-4'>
      <div className='flex items-start justify-between gap-3'>
        <div className='flex min-w-0 items-center gap-3'>
          <span
            className={cn(
              'flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold',
              badgeClass(group.name)
            )}
          >
            {initial}
          </span>
          <div className='min-w-0'>
            <div className='flex min-w-0 items-baseline gap-2'>
              <h3 className='truncate text-base font-semibold'>{title}</h3>
              <span className='text-muted-foreground shrink-0 text-sm'>
                {formatRatio(group.ratio)}
              </span>
            </div>
            {title !== group.name && (
              <p className='text-muted-foreground truncate text-xs'>{group.name}</p>
            )}
          </div>
        </div>
        <span
          className={cn(
            'inline-flex shrink-0 items-center gap-1.5 text-sm',
            group.enabled
              ? 'text-emerald-600 dark:text-emerald-400'
              : 'text-muted-foreground'
          )}
        >
          <span
            className={cn(
              'size-2 rounded-full',
              group.enabled ? 'bg-emerald-500' : 'bg-muted-foreground/50'
            )}
          />
          {group.enabled ? t('Available') : t('No enabled models')}
        </span>
      </div>

      <div className='text-muted-foreground flex items-center justify-between gap-3 text-sm'>
        <span>{t('{{count}} models', { count: group.modelCount })}</span>
        <span>
          {t('Cache rate')} {formatPercent(group.cacheRate)}
        </span>
      </div>

      <div className='grid grid-cols-3 gap-2'>
        <Metric
          label={t('Availability')}
          value={formatPercent(group.availability)}
          valueClass={percentClass(group.availability)}
        />
        <Metric
          label={t('Average response time')}
          value={formatSeconds(group.avgUseTime)}
        />
        <Metric
          label={t('Average first token latency')}
          value={formatSeconds(group.avgFirstToken)}
        />
      </div>

      <div>
        <div className='text-muted-foreground mb-2 flex items-center justify-between gap-3 text-xs'>
          <span>{t('Request history, last 30 minutes')}</span>
          <span>
            {group.latestAt
              ? t('Latest sample {{time}}', {
                  time: dayjs.unix(group.latestAt).format('HH:mm:ss'),
                })
              : t('No requests in the last 30 minutes')}
          </span>
        </div>
        <div className='flex h-7 items-end gap-0.5' aria-hidden='true'>
          {group.history.map((bucket, index) => {
            const minute = bucketMinute(
              props.generatedAt,
              index,
              group.history.length
            )
            const label =
              bucket.success + bucket.error === 0
                ? `${dayjs.unix(minute).format('HH:mm')} · ${t('No requests in this minute')}`
                : `${dayjs.unix(minute).format('HH:mm')} · ${t(
                    'Success {{success}}, failed {{failed}}',
                    { success: bucket.success, failed: bucket.error }
                  )}`
            return (
              <span
                key={`${group.name}-${minute}`}
                title={label}
                className={cn('h-full min-w-0 flex-1 rounded-[2px]', barClass(bucket))}
              />
            )
          })}
        </div>
      </div>

      {props.detailed && (
        <div className='space-y-2 border-t pt-3'>
          {group.models.length === 0 ? (
            <p className='text-muted-foreground text-xs'>
              {t('No requests in the last 30 minutes')}
            </p>
          ) : (
            group.models.map((model) => (
              <div
                key={model.model || 'unknown'}
                className='grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 text-xs'
              >
                <span className='truncate font-medium'>
                  {model.model || t('Unknown model')}
                </span>
                <span className='text-muted-foreground shrink-0'>
                  <span className={percentClass(model.availability)}>
                    {formatPercent(model.availability)}
                  </span>
                  {' · '}
                  {formatSeconds(model.avgUseTime)}
                  {' · '}
                  {formatSeconds(model.avgFirstToken)}
                  {' · '}
                  {t('{{count}} requests', { count: model.requests })}
                </span>
              </div>
            ))
          )}
          {group.modelsWithoutRequests > 0 && (
            <p className='text-muted-foreground text-xs'>
              {t(
                '{{count}} configured models had no requests in this window',
                { count: group.modelsWithoutRequests }
              )}
            </p>
          )}
        </div>
      )}
    </article>
  )
}

function Metric(props: { label: string; value: string; valueClass?: string }) {
  return (
    <div className='bg-muted/60 rounded-lg px-3 py-2'>
      <div className='text-muted-foreground truncate text-xs'>{props.label}</div>
      <div className={cn('mt-1 text-sm font-semibold', props.valueClass)}>
        {props.value}
      </div>
    </div>
  )
}

function StatusSkeleton() {
  return (
    <div className='grid grid-cols-1 gap-4 xl:grid-cols-2 2xl:grid-cols-3'>
      {Array.from({ length: 3 }, (_, index) => (
        <Skeleton key={index} className='h-64 rounded-xl' />
      ))}
    </div>
  )
}
