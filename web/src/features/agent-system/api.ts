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

import { api } from '@/lib/api'
import { requireServerSuccess } from '@/lib/server-error-message'

import type {
  ApprovedAgentPage,
  AgentCommissionAdminAccount,
  AgentCommissionDeductResult,
  AgentCommissionDeduction,
  AgentInviteePage,
  AgentRateUpdate,
  AgentRecordPage,
  AgentStatsReport,
  AgentSystemReport,
  AgentWithdrawal,
} from './types'

type AgentSystemResponse = {
  success: boolean
  data?: AgentSystemReport
}

export async function getAgentSystem() {
  const res = await api.get('/api/agent-system')
  const body = requireServerSuccess(res.data) as AgentSystemResponse
  if (!body.data) {
    throw new Error('Failed to load agent system')
  }
  return {
    ...body.data,
    users: body.data.users ?? [],
  }
}

export async function updateAgentAlipay(account: string) {
  const res = await api.put('/api/agent-system/alipay', { account })
  return requireServerSuccess(res.data)
}

export async function createAgentWithdrawal(quota: number) {
  const res = await api.post('/api/agent-system/withdrawals', { quota })
  return requireServerSuccess(res.data)
}

export async function listMyAgentWithdrawals() {
  const res = await api.get('/api/agent-system/withdrawals')
  const body = requireServerSuccess(res.data) as { data?: AgentWithdrawal[] }
  return body.data ?? []
}

export async function listAgentWithdrawals() {
  const res = await api.get('/api/agent-system/admin/withdrawals')
  const body = requireServerSuccess(res.data) as { data?: AgentWithdrawal[] }
  return body.data ?? []
}

export async function processAgentWithdrawal(
  id: number,
  paid: boolean,
  note: string
) {
  const res = await api.post(
    `/api/agent-system/admin/withdrawals/${id}/process`,
    { paid, note }
  )
  return requireServerSuccess(res.data)
}

export async function searchAgentCommissionAccounts(keyword: string) {
  const res = await api.get('/api/agent-system/admin/accounts', {
    params: { keyword },
  })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentCommissionAdminAccount[]
  }
  return body.data ?? []
}

export async function listAgentCommissionDeductions() {
  const res = await api.get('/api/agent-system/admin/deductions')
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentCommissionDeduction[]
  }
  return body.data ?? []
}

export async function deductAgentCommission(payload: {
  userId: number
  quota: number
  note: string
}) {
  const res = await api.post('/api/agent-system/admin/deduct', {
    user_id: payload.userId,
    quota: payload.quota,
    note: payload.note,
  })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentCommissionDeductResult
  }
  if (!body.data) {
    throw new Error('Failed to deduct commission')
  }
  return body.data
}

export async function listApprovedAgents(page: number) {
  const res = await api.get('/api/agent-system/admin/agents', {
    params: { page, pageSize: 20 },
  })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: ApprovedAgentPage
  }
  return {
    page: body.data?.page ?? page,
    pageSize: body.data?.pageSize ?? 20,
    total: body.data?.total ?? 0,
    items: body.data?.items ?? [],
  }
}

export async function setAgentCommissionRates(payload: {
  userId: number
  level1Rate: number
  level2Rate: number
}) {
  const res = await api.post('/api/agent-system/admin/rates', {
    user_id: payload.userId,
    level1_rate: payload.level1Rate,
    level2_rate: payload.level2Rate,
  })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentRateUpdate
  }
  if (!body.data) {
    throw new Error('Failed to save commission rates')
  }
  return body.data
}

export async function getAgentSystemStats(days: number) {
  const res = await api.get('/api/agent-system/stats', { params: { days } })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentStatsReport
  }
  if (!body.data) {
    throw new Error('Failed to load statistics')
  }
  return {
    ...body.data,
    points: body.data.points ?? [],
    payments: body.data.payments ?? [],
  }
}

export async function getAgentSystemRecords(params: {
  kind: 'recharge' | 'commission'
  keyword: string
  page: number
  pageSize: number
}) {
  const res = await api.get('/api/agent-system/records', { params })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentRecordPage
  }
  if (!body.data) {
    throw new Error('Failed to load records')
  }
  return {
    ...body.data,
    records: body.data.records ?? [],
  }
}

export async function getAgentSystemUsers(params: {
  keyword: string
  page: number
  pageSize: number
}) {
  const res = await api.get('/api/agent-system/users', { params })
  const body = requireServerSuccess(res.data) as {
    success: boolean
    data?: AgentInviteePage
  }
  if (!body.data) {
    throw new Error('Failed to load invited users')
  }
  return {
    ...body.data,
    users: body.data.users ?? [],
  }
}
