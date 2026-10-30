import { api } from '@/lib/api'
import type { TopupRecord } from '@/features/wallet/types'

export interface BillingListResponse {
  items: (TopupRecord & { username: string; agent_name: string })[]
  total: number
}

export interface BillingPeriodSummary {
  amount: number
  count: number
}

export type BillingSummary = Record<
  'all' | 'month' | 'week' | 'last_week' | 'today' | 'yesterday',
  BillingPeriodSummary
>

export interface BillingFilters {
  keyword: string
  paymentMethod: string
  status: string
  startTime: string
  endTime: string
  page: number
  pageSize: number
}

export async function getAdminBillingRecords(filters: BillingFilters) {
  const params = new URLSearchParams({
    p: String(filters.page),
    page_size: String(filters.pageSize),
  })
  if (filters.keyword) params.set('keyword', filters.keyword)
  if (filters.paymentMethod) params.set('payment_method', filters.paymentMethod)
  if (filters.status) params.set('status', filters.status)
  if (filters.startTime) params.set('start_time', localDateStart(filters.startTime).toISOString())
  if (filters.endTime) {
    const end = localDateStart(filters.endTime)
    end.setDate(end.getDate() + 1)
    params.set('end_time', end.toISOString())
  }
  const response = await api.get(`/api/billing/admin/records?${params.toString()}`)
  return response.data as { success: boolean; data: BillingListResponse }
}

export async function getAdminBillingSummary() {
  const response = await api.get('/api/billing/admin/summary')
  return response.data as { success: boolean; data: BillingSummary }
}

function localDateStart(value: string): Date {
  const [year, month, day] = value.split('-').map(Number)
  return new Date(year, month - 1, day)
}
