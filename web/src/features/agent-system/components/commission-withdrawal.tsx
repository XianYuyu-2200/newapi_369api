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
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  formatQuota,
  formatTimestamp,
  parseQuotaFromDollars,
} from '@/lib/format'
import { getServerErrorMessage } from '@/lib/server-error-message'

import {
  createAgentWithdrawal,
  listMyAgentWithdrawals,
  updateAgentAlipay,
} from '../api'
import type { AgentWithdrawal } from '../types'

export function CommissionWithdrawal(props: {
  availableQuota: number
  minimumQuota: number
  alipayAccount: string
  pendingQuota: number
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [account, setAccount] = useState(props.alipayAccount)
  const [amount, setAmount] = useState('')
  const [busy, setBusy] = useState(false)
  const withdrawals = useQuery({
    queryKey: ['agent-withdrawals'],
    queryFn: listMyAgentWithdrawals,
  })

  async function saveAccount() {
    setBusy(true)
    try {
      await updateAgentAlipay(account.trim())
      await queryClient.invalidateQueries({ queryKey: ['agent-system'] })
      toast.success(t('Alipay account saved'))
    } catch (error) {
      toast.error(getServerErrorMessage(error))
    } finally {
      setBusy(false)
    }
  }

  async function request(quota: number) {
    if (!account.trim()) {
      toast.error(t('Bind your Alipay account first'))
      return
    }
    if (!Number.isFinite(quota) || quota < props.minimumQuota) {
      toast.error(t('Withdrawal amount is below the minimum'))
      return
    }
    setBusy(true)
    try {
      await createAgentWithdrawal(Math.round(quota))
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['agent-system'] }),
        queryClient.invalidateQueries({ queryKey: ['agent-withdrawals'] }),
      ])
      setAmount('')
      toast.success(t('Withdrawal request submitted'))
    } catch (error) {
      toast.error(getServerErrorMessage(error))
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className='w-full space-y-3'>
      <div className='flex flex-wrap items-end gap-2'>
        <label className='grid min-w-56 flex-1 gap-1 text-sm'>
          <span>{t('Alipay account')}</span>
          <Input
            value={account}
            maxLength={128}
            onChange={(event) => setAccount(event.target.value)}
            placeholder={t('Enter Alipay account')}
          />
        </label>
        <Button
          size='sm'
          variant='outline'
          disabled={busy || account.trim() === props.alipayAccount}
          onClick={() => void saveAccount()}
        >
          {t('Save account')}
        </Button>
      </div>
      <p className='text-muted-foreground text-sm'>
        {t('Pending withdrawal amount {{amount}}', {
          amount: formatQuota(
            withdrawals.data
              ?.filter((row) => row.status === 'pending')
              .reduce((sum, row) => sum + row.quota, 0) ?? 0
          ),
        })}
      </p>
      <div className='flex flex-wrap items-end gap-2'>
        <label className='grid min-w-40 flex-1 gap-1 text-sm'>
          <span>{t('Withdrawal amount')}</span>
          <Input
            inputMode='decimal'
            value={amount}
            onChange={(event) => setAmount(event.target.value)}
            placeholder={formatQuota(props.availableQuota)}
          />
        </label>
        <Button
          size='sm'
          disabled={busy || props.availableQuota < props.minimumQuota}
          onClick={() => void request(parseQuotaFromDollars(Number(amount)))}
        >
          {t('Request withdrawal')}
        </Button>
        <Button
          size='sm'
          variant='outline'
          disabled={busy || props.availableQuota < props.minimumQuota}
          onClick={() => void request(props.availableQuota)}
        >
          {t('Withdraw all available')}
        </Button>
      </div>
      <div className='divide-y rounded-md border px-3'>
        {withdrawals.data?.length ? (
          withdrawals.data.map((row) => (
            <WithdrawalRow key={row.id} row={row} />
          ))
        ) : (
          <p className='text-muted-foreground py-3 text-sm'>
            {t('No withdrawal requests yet')}
          </p>
        )}
      </div>
    </div>
  )
}

function WithdrawalRow({ row }: { row: AgentWithdrawal }) {
  const { t } = useTranslation()
  const status = withdrawalStatus(t, row.status)
  return (
    <div className='flex flex-wrap justify-between gap-2 py-2 text-sm'>
      <span>
        {formatQuota(row.quota)} · {status}
      </span>
      <span className='text-muted-foreground'>
        {formatTimestamp(row.createdAt)}
        {row.adminNote ? ` · ${row.adminNote}` : ''}
      </span>
    </div>
  )
}

function withdrawalStatus(
  t: ReturnType<typeof useTranslation>['t'],
  status: AgentWithdrawal['status']
) {
  if (status === 'pending') return t('Pending')
  if (status === 'paid') return t('Paid')
  return t('Rejected')
}
