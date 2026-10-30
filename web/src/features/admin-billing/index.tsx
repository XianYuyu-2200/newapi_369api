import {
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  CircleDollarSign,
  ClipboardList,
  Search,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { StatusBadge } from '@/components/status-badge'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { formatNumber } from '@/lib/format'
import { handleServerError } from '@/lib/handle-server-error'
import { completeOrder } from '@/features/wallet/api'
import type { TopupRecord } from '@/features/wallet/types'
import { getPaymentMethodName, getStatusConfig, formatTimestamp } from '@/features/wallet/lib/billing'

import {
  getAdminBillingRecords,
  getAdminBillingSummary,
  type BillingFilters,
  type BillingListResponse,
  type BillingSummary,
} from './api'

const initialFilters: BillingFilters = {
  keyword: '', paymentMethod: '', status: '', startTime: '', endTime: '', page: 1, pageSize: 10,
}

const periods = [
  { key: 'all', label: 'Total top-ups' },
  { key: 'month', label: 'This month' },
  { key: 'week', label: 'This week' },
  { key: 'last_week', label: 'Last week' },
  { key: 'today', label: 'Today' },
  { key: 'yesterday', label: 'Yesterday' },
] as const

function money(value: number) {
  return formatNumber(value)
}

export function AdminBilling() {
  const { t } = useTranslation()
  const [filters, setFilters] = useState(initialFilters)
  const [draft, setDraft] = useState(initialFilters)
  const [list, setList] = useState<BillingListResponse>({ items: [], total: 0 })
  const [summary, setSummary] = useState<BillingSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [busyOrder, setBusyOrder] = useState<string | null>(null)
  const [confirmOrder, setConfirmOrder] = useState<TopupRecord | null>(null)
  const [refresh, setRefresh] = useState(0)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const [recordsResponse, summaryResponse] = await Promise.all([
        getAdminBillingRecords(filters), getAdminBillingSummary(),
      ])
      if (recordsResponse.success) setList(recordsResponse.data)
      if (summaryResponse.success) setSummary(summaryResponse.data)
    } catch (error) {
      handleServerError(error, t('Failed to load billing history'))
    } finally {
      setLoading(false)
    }
  }, [filters, t])

  useEffect(() => { void load() }, [load, refresh])

  const totalPages = Math.max(1, Math.ceil(list.total / filters.pageSize))
  const rows = useMemo(() => list.items, [list.items])

  const applyFilters = () => setFilters({ ...draft, page: 1 })
  const resetFilters = () => { setDraft(initialFilters); setFilters(initialFilters) }
  const changePage = (page: number) => setFilters((current) => ({ ...current, page }))

  const reconcile = async (record: TopupRecord) => {
    if (record.status !== 'pending') return false
    setBusyOrder(record.trade_no)
    try {
      const result = await completeOrder({ trade_no: record.trade_no })
      if (result.success) {
        setRefresh((value) => value + 1)
        return true
      }
      handleServerError(result, t('Failed to complete order'))
    } catch (error) {
      handleServerError(error, t('Failed to complete order'))
    } finally {
      setBusyOrder(null)
    }
    return false
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Billing Management')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='flex flex-col gap-4 p-1'>
          <section aria-label={t('Billing summary')} className='grid gap-3 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-6'>
            {periods.map(({ key, label }) => {
              const value = summary?.[key]
              return (
                <Card key={key} className='gap-2 py-4'>
                  <CardContent className='flex items-start justify-between gap-2 px-4'>
                    <div className='min-w-0'>
                      <p className='text-muted-foreground truncate text-xs'>{t(label)}</p>
                      <p className='mt-2 text-xl font-semibold tabular-nums'>{money(value?.amount ?? 0)}</p>
                      <p className='text-muted-foreground mt-1 text-xs'>{t('{{count}} successful orders', { count: value?.count ?? 0 })}</p>
                    </div>
                    <span className='bg-primary/10 text-primary rounded-lg p-2'><CircleDollarSign className='size-4' /></span>
                  </CardContent>
                </Card>
              )
            })}
          </section>

          <Card className='gap-0 py-0'>
            <CardContent className='p-0'>
              <div className='grid gap-2 border-b p-3 sm:grid-cols-2 xl:grid-cols-[minmax(180px,1.5fr)_minmax(130px,1fr)_minmax(130px,1fr)_minmax(145px,1fr)_minmax(145px,1fr)_auto]'>
                <div className='relative'>
                  <Search className='text-muted-foreground absolute top-1/2 left-3 size-4 -translate-y-1/2' />
                  <Input aria-label={t('Search orders')} placeholder={t('Order number, username or user ID')} className='pl-9' value={draft.keyword} onChange={(event) => setDraft((current) => ({ ...current, keyword: event.target.value }))} onKeyDown={(event) => event.key === 'Enter' && applyFilters()} />
                </div>
                <Select value={draft.paymentMethod || 'all'} onValueChange={(value) => setDraft((current) => ({ ...current, paymentMethod: value === 'all' ? '' : value ?? '' }))}>
                  <SelectTrigger aria-label={t('Payment method')}><SelectValue placeholder={t('All payment methods')} /></SelectTrigger>
                  <SelectContent><SelectItem value='all'>{t('All payment methods')}</SelectItem><SelectItem value='alipay'>{t('Alipay')}</SelectItem><SelectItem value='wxpay'>{t('WeChat Pay')}</SelectItem><SelectItem value='stripe'>Stripe</SelectItem><SelectItem value='creem'>Creem</SelectItem><SelectItem value='waffo'>Waffo</SelectItem><SelectItem value='balance'>{t('Balance')}</SelectItem></SelectContent>
                </Select>
                <Input aria-label={t('Start date')} type='date' value={draft.startTime} onChange={(event) => setDraft((current) => ({ ...current, startTime: event.target.value }))} />
                <Input aria-label={t('End date')} type='date' value={draft.endTime} onChange={(event) => setDraft((current) => ({ ...current, endTime: event.target.value }))} />
                <Select value={draft.status || 'all'} onValueChange={(value) => setDraft((current) => ({ ...current, status: value === 'all' ? '' : value ?? '' }))}>
                  <SelectTrigger aria-label={t('Order status')}><SelectValue placeholder={t('All statuses')} /></SelectTrigger>
                  <SelectContent><SelectItem value='all'>{t('All statuses')}</SelectItem><SelectItem value='success'>{t('Success')}</SelectItem><SelectItem value='pending'>{t('Pending')}</SelectItem><SelectItem value='expired'>{t('Expired')}</SelectItem></SelectContent>
                </Select>
                <div className='flex gap-2'>
                  <Button onClick={applyFilters}><Search className='mr-1 size-4' />{t('Filter')}</Button>
                  <Button variant='outline' onClick={resetFilters}><X className='mr-1 size-4' />{t('Reset')}</Button>
                </div>
              </div>

              <div className='overflow-x-auto'>
                <table className='w-full min-w-[1120px] text-sm'>
                  <thead className='bg-muted/40 text-muted-foreground text-left text-xs'>
                    <tr>{['Username', 'Bound agent', 'User ID', 'Payment method', 'Top-up amount', 'Payment amount', 'Status', 'Created at', 'Action', 'Order number'].map((label) => <th key={label} className='px-3 py-3 font-medium'>{t(label)}</th>)}</tr>
                  </thead>
                  <tbody className='divide-y'>
                    {loading && <tr><td colSpan={10} className='text-muted-foreground h-36 text-center'>{t('Loading...')}</td></tr>}
                    {!loading && rows.length === 0 && <tr><td colSpan={10} className='text-muted-foreground h-36 text-center'>{t('No billing records found')}</td></tr>}
                    {!loading && rows.map((record) => {
                      const status = getStatusConfig(record.status)
                      return <tr key={record.id} className='hover:bg-muted/30'>
                        <td className='px-3 py-3 font-medium'>{record.username || '—'}</td>
                        <td className='px-3 py-3'>{record.agent_name || '—'}</td>
                        <td className='px-3 py-3 tabular-nums'>{record.user_id}</td>
                        <td className='px-3 py-3'>{getPaymentMethodName(record.payment_method, t)}</td>
                        <td className='px-3 py-3 font-medium'>{money(record.amount)}</td>
                        <td className='text-destructive px-3 py-3 font-medium'>{money(record.money)}</td>
                        <td className='px-3 py-3'><StatusBadge label={t(status.label)} variant={status.variant} showDot /></td>
                        <td className='px-3 py-3 whitespace-nowrap'>{formatTimestamp(record.create_time)}</td>
                        <td className='px-3 py-3'>{record.status === 'pending' ? <Button size='sm' variant='outline' disabled={busyOrder === record.trade_no} onClick={() => setConfirmOrder(record)}><CheckCircle2 className='mr-1 size-3.5' />{busyOrder === record.trade_no ? t('Processing...') : t('Reconcile')}</Button> : '—'}</td>
                        <td className='px-3 py-3'><code className='max-w-56 truncate text-xs' title={record.trade_no}>{record.trade_no}</code></td>
                      </tr>
                    })}
                  </tbody>
                </table>
              </div>
              <div className='text-muted-foreground flex flex-wrap items-center justify-between gap-3 border-t px-3 py-3 text-xs'>
                <span>{t('Showing {{from}}–{{to}} of {{total}} orders', { from: list.total ? (filters.page - 1) * filters.pageSize + 1 : 0, to: Math.min(filters.page * filters.pageSize, list.total), total: list.total })}</span>
                <div className='flex items-center gap-2'>
                  <span>{t('Page {{page}} of {{total}}', { page: filters.page, total: totalPages })}</span>
                  <Button aria-label={t('Previous page')} variant='outline' size='icon' className='size-8' disabled={filters.page <= 1 || loading} onClick={() => changePage(filters.page - 1)}><ChevronLeft className='size-4' /></Button>
                  <Button aria-label={t('Next page')} variant='outline' size='icon' className='size-8' disabled={filters.page >= totalPages || loading} onClick={() => changePage(filters.page + 1)}><ChevronRight className='size-4' /></Button>
                  <Select value={String(filters.pageSize)} onValueChange={(value) => value && setFilters((current) => ({ ...current, page: 1, pageSize: Number(value) }))}>
                    <SelectTrigger className='h-8 w-28'><SelectValue /></SelectTrigger><SelectContent><SelectItem value='10'>{t('10 / page')}</SelectItem><SelectItem value='20'>{t('20 / page')}</SelectItem><SelectItem value='50'>{t('50 / page')}</SelectItem><SelectItem value='100'>{t('100 / page')}</SelectItem></SelectContent>
                  </Select>
                </div>
              </div>
            </CardContent>
          </Card>

          <div className='bg-muted/50 text-muted-foreground flex items-center gap-3 rounded-lg border p-4 text-sm'>
            <ClipboardList className='text-primary size-5 shrink-0' />
            <div><p className='text-foreground font-medium'>{t('Top-up reconciliation')}</p><p>{t('Pending payment orders can be reconciled from the Action column after verifying payment.')}</p></div>
            
          </div>
        </div>
      </SectionPageLayout.Content>
      <AlertDialog open={confirmOrder !== null} onOpenChange={(open) => !open && setConfirmOrder(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('Confirm top-up reconciliation')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('Confirm that payment has been received for order {{order}}. This will credit the user account.', { order: confirmOrder?.trade_no ?? '' })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={busyOrder !== null}>{t('Cancel')}</AlertDialogCancel>
            <AlertDialogAction disabled={!confirmOrder || busyOrder !== null} onClick={async (event) => {
              event.preventDefault()
              if (confirmOrder && await reconcile(confirmOrder)) setConfirmOrder(null)
            }}>{busyOrder ? t('Processing...') : t('Confirm top-up')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </SectionPageLayout>
  )
}
