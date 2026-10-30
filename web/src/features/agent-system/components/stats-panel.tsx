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
import { useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  XAxis,
  YAxis,
} from 'recharts'

import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { formatQuota, quotaUnitsToDollars } from '@/lib/format'

import { getAgentSystemStats } from '../api'
import { MetricCard } from './metric'

const ranges = [7, 15, 30] as const

type Range = (typeof ranges)[number]

type TrendPoint = {
  label: string
  orders: number
  success: number
  recharge: number
  commission: number
  rate: number
}

export function StatsPanel() {
  const { t } = useTranslation()
  const [range, setRange] = useState<Range>(7)
  const query = useQuery({
    queryKey: ['agent-system', 'stats', range],
    queryFn: () => getAgentSystemStats(range),
    placeholderData: keepPreviousData,
  })
  const stats = query.data
  const data: TrendPoint[] = (stats?.points ?? []).map((point) => ({
    label: point.label,
    orders: point.orders,
    success: point.success,
    recharge: quotaNumber(point.rechargeQuota),
    commission: quotaNumber(point.commissionQuota),
    rate: point.rate,
  }))

  return (
    <div className='space-y-4'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <p className='text-muted-foreground max-w-3xl text-sm'>
          {t(
            'Orders and commission are grouped by day in Beijing time. Only successful recharges earn commission. Each order keeps the rate saved when it succeeded.'
          )}
        </p>
        <div className='flex gap-2'>
          {ranges.map((item) => (
            <Button
              key={item}
              type='button'
              size='sm'
              variant={item === range ? 'default' : 'outline'}
              onClick={() => setRange(item)}
            >
              {t('Last {{count}} days', { count: item })}
            </Button>
          ))}
        </div>
      </div>

      {query.isLoading ? <Skeleton className='h-80 rounded-xl' /> : null}
      {query.isError && !stats ? (
        <p className='text-sm text-red-600 dark:text-red-400'>
          {t('Failed to load statistics')}
        </p>
      ) : null}
      {stats ? (
        <>
          <div className='grid gap-3 md:grid-cols-2 xl:grid-cols-5'>
            <MetricCard
              label={t('Order count')}
              value={formatCount(stats.orderCount)}
              hint={t(
                "Invited users' orders in this period. Successful orders are counted on the completion day."
              )}
            />
            <MetricCard
              label={t('Success rate')}
              value={formatRate(stats.successRate)}
              hint={t('Successful orders divided by total orders')}
            />
            <MetricCard
              label={t('Paid orders')}
              value={formatCount(stats.successCount)}
              hint={t('Paid orders in this period')}
            />
            <MetricCard
              label={t('Recharge amount')}
              value={formatQuota(stats.rechargeQuota)}
              hint={t('Successful recharge amount in this period')}
            />
            <MetricCard
              label={t('Commission amount')}
              value={formatQuota(stats.commissionQuota)}
              hint={t('Commission credited in this period')}
            />
          </div>

          <div className='grid gap-3 lg:grid-cols-2'>
            <ChartCard title={t('Order trend')}>
              <TrendChart
                data={data}
                lines={[
                  { key: 'orders', name: t('Orders'), color: '#94a3b8' },
                  {
                    key: 'success',
                    name: t('Successful orders'),
                    color: '#22c55e',
                  },
                ]}
              />
            </ChartCard>
            <ChartCard title={t('Daily payment success rate')}>
              <TrendChart
                data={data}
                lines={[
                  { key: 'rate', name: t('Success rate'), color: '#2468f2' },
                ]}
                domain={[0, 1]}
                tickFormatter={(value) => `${Math.round(value * 100)}%`}
              />
            </ChartCard>
            <ChartCard title={t('Recharge and commission trend')}>
              <TrendChart
                data={data}
                lines={[
                  {
                    key: 'recharge',
                    name: t('Recharge amount'),
                    color: '#f97316',
                  },
                  {
                    key: 'commission',
                    name: t('Commission amount'),
                    color: '#14b8a6',
                  },
                ]}
                tickFormatter={(value) =>
                  Number.isInteger(value) ? String(value) : value.toFixed(2)
                }
              />
            </ChartCard>
            <ChartCard title={t('Payment method distribution')}>
              {stats.payments.length === 0 ? (
                <div className='text-muted-foreground flex h-56 items-center justify-center text-sm'>
                  {t('No successful recharges in this period')}
                </div>
              ) : (
                <PaymentList
                  payments={stats.payments.map((item) => ({
                    method:
                      item.method === 'unknown' ? t('Unknown') : item.method,
                    count: item.count,
                    rechargeQuota: item.rechargeQuota,
                  }))}
                />
              )}
            </ChartCard>
          </div>

          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
            <BottomStat
              label={t('Successful order count')}
              value={formatCount(stats.successCount)}
            />
            <BottomStat
              label={t('Pending payment orders')}
              value={formatCount(stats.pendingCount)}
            />
            <BottomStat
              label={t('Failed or expired orders')}
              value={formatCount(stats.failedCount)}
            />
            <BottomStat
              label={t('Commission orders')}
              value={formatCount(stats.successCount)}
            />
          </div>
        </>
      ) : null}
    </div>
  )
}

function PaymentList(props: {
  payments: Array<{ method: string; count: number; rechargeQuota: number }>
}) {
  const max = Math.max(...props.payments.map((item) => item.rechargeQuota), 1)
  return (
    <div className='flex h-56 flex-col justify-center gap-3 overflow-auto pr-1'>
      {props.payments.map((item) => (
        <div key={item.method}>
          <div className='mb-1 flex items-center justify-between gap-3 text-xs'>
            <span className='truncate font-medium'>{item.method}</span>
            <span className='text-muted-foreground shrink-0 tabular-nums'>
              {formatCount(item.count)} · {formatQuota(item.rechargeQuota)}
            </span>
          </div>
          <div className='bg-muted h-2 overflow-hidden rounded'>
            <div
              className='h-2 rounded bg-[#2468f2]'
              style={{
                width: `${Math.max(4, (item.rechargeQuota / max) * 100)}%`,
              }}
            />
          </div>
        </div>
      ))}
    </div>
  )
}

function ChartCard(props: { title: string; children: ReactNode }) {
  return (
    <section className='rounded-lg border p-4'>
      <h3 className='mb-3 text-sm font-medium'>{props.title}</h3>
      {props.children}
    </section>
  )
}

function TrendChart(props: {
  data: TrendPoint[]
  lines: Array<{ key: string; name: string; color: string }>
  domain?: [number, number]
  tickFormatter?: (value: number) => string
}) {
  return (
    <div className='text-muted-foreground h-56 w-full'>
      <ResponsiveContainer width='100%' height='100%'>
        <LineChart data={props.data} margin={{ left: 0, right: 8, top: 8 }}>
          <CartesianGrid
            stroke='currentColor'
            strokeOpacity={0.2}
            vertical={false}
          />
          <XAxis dataKey='label' tick={{ fontSize: 12 }} tickLine={false} />
          <YAxis
            domain={props.domain}
            tick={{ fontSize: 12 }}
            tickFormatter={props.tickFormatter}
            tickLine={false}
            width={56}
          />
          <Legend />
          {props.lines.map((line) => (
            <Line
              key={line.key}
              dataKey={line.key}
              name={line.name}
              stroke={line.color}
              strokeWidth={2}
              dot={{ r: 3 }}
              isAnimationActive={false}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}

function BottomStat(props: { label: string; value: string }) {
  return (
    <div className='flex items-center justify-between gap-3 rounded-lg border px-4 py-3'>
      <span className='text-muted-foreground text-sm'>{props.label}</span>
      <span className='text-lg font-semibold tabular-nums'>{props.value}</span>
    </div>
  )
}

function quotaNumber(quota: number) {
  const value = quotaUnitsToDollars(quota)
  return Number.isFinite(value) ? value : 0
}

function formatCount(value: number) {
  return new Intl.NumberFormat().format(value)
}

function formatRate(rate: number) {
  if (!Number.isFinite(rate)) return '0.00%'
  return `${(rate * 100).toFixed(2)}%`
}
