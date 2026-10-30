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

export type AgentInvitee = {
  id: number
  username: string
  displayName: string
  email: string
  status: number
  level: number
  createdAt: number
  rechargeCount?: number
  rechargeQuota?: number
  commissionQuota?: number
}

export type AgentStatus = 'approved'

export type AgentSystemReport = {
  commissionEnabled: boolean
  agentStatus: AgentStatus
  inviteCode: string
  level1Rate: number
  level2Rate: number
  level1RateCustom: boolean
  level2RateCustom: boolean
  level1Count: number
  level2Count: number
  users: AgentInvitee[]
  userLimit: number
  usersTruncated: boolean
  totalQuota?: number
  transferredQuota?: number
  deductedQuota?: number
  availableQuota?: number
  minimumTransferQuota?: number
  level1RechargeQuota?: number
  level2RechargeQuota?: number
  level1CommissionQuota?: number
  level2CommissionQuota?: number
  level1OrderCount?: number
  level2OrderCount?: number
}

export type AgentCommissionTransferResult = {
  quota: number
  availableQuota: number
  transferredQuota: number
  balanceQuota: number
}

export type AgentCommissionAdminAccount = {
  id: number
  username: string
  displayName: string
  agentStatus: AgentStatus
  totalQuota: number
  transferredQuota: number
  deductedQuota: number
  availableQuota: number
  level1Rate: number
  level2Rate: number
  level1Custom: boolean
  level2Custom: boolean
}

export type ApprovedAgent = {
  id: number
  username: string
  displayName: string
  inviteCode: string
  quota: number
  usedQuota: number
  level1Count: number
  level2Count: number
  totalQuota: number
  transferredQuota: number
  deductedQuota: number
  availableQuota: number
  level1CommissionQuota: number
  level2CommissionQuota: number
  level1Rate: number
  level2Rate: number
  level1Custom: boolean
  level2Custom: boolean
  appliedAt: number
  reviewedAt: number
}

export type ApprovedAgentPage = {
  page: number
  pageSize: number
  total: number
  items: ApprovedAgent[]
}

export type AgentRateUpdate = {
  userId: number
  level1Rate: number
  level2Rate: number
  level1Custom: boolean
  level2Custom: boolean
}

export type AgentCommissionDeduction = {
  id: number
  userId: number
  username: string
  adminId: number
  adminUsername: string
  quota: number
  note: string
  createdAt: number
}

export type AgentCommissionDeductResult = {
  quota: number
  availableQuota: number
  transferredQuota: number
  deductedQuota: number
  totalQuota: number
  balanceQuota: number
}

export type AgentStatsPoint = {
  date: string
  label: string
  orders: number
  success: number
  pending: number
  failed: number
  rechargeQuota: number
  commissionQuota: number
  rate: number
}

export type AgentPaymentSlice = {
  method: string
  count: number
  rechargeQuota: number
}

export type AgentStatsReport = {
  days: number
  orderCount: number
  successCount: number
  pendingCount: number
  failedCount: number
  successRate: number
  rechargeQuota: number
  commissionQuota: number
  points: AgentStatsPoint[]
  payments: AgentPaymentSlice[]
}

export type AgentOrderRecord = {
  id: number
  tradeNo: string
  userId: number
  username: string
  displayName: string
  email: string
  level: number
  status: string
  paymentMethod: string
  amount: number
  rechargeQuota: number
  commissionQuota: number
  createdAt: number
  completedAt: number
}

export type AgentRecordPage = {
  kind: 'recharge' | 'commission'
  page: number
  pageSize: number
  total: number
  todayCommissionQuota: number
  yesterdayCommissionQuota: number
  weekCommissionQuota: number
  records: AgentOrderRecord[]
}

export type AgentInviteePage = {
  page: number
  pageSize: number
  total: number
  users: AgentInvitee[]
}
