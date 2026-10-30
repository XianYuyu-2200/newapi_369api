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
import { BadgePercent } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  formatQuota,
  formatTimestamp,
  parseQuotaFromDollars,
  quotaUnitsToEditableAmount,
} from '@/lib/format'
import { getServerErrorMessage } from '@/lib/server-error-message'
import { cn } from '@/lib/utils'

import {
  deductAgentCommission,
  listAgentCommissionDeductions,
  searchAgentCommissionAccounts,
} from './api'
import { AdminApprovedAgents } from './components/admin-approved-agents'
import { AdminRateForm } from './components/admin-rate-form'
import type { AgentCommissionAdminAccount } from './types'

export function AgentCommissionAdminPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [keyword, setKeyword] = useState('')
  const [accounts, setAccounts] = useState<AgentCommissionAdminAccount[]>([])
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [searched, setSearched] = useState(false)
  const [searching, setSearching] = useState(false)
  const [amount, setAmount] = useState('')
  const [note, setNote] = useState('')
  const [deducting, setDeducting] = useState(false)
  const deductions = useQuery({
    queryKey: ['agent-commission-deductions'],
    queryFn: listAgentCommissionDeductions,
  })
  const selected = accounts.find((account) => account.id === selectedId) ?? null

  async function search() {
    const value = keyword.trim()
    if (!value) {
      toast.error(t('Enter a username or user ID'))
      return
    }
    setSearching(true)
    try {
      const rows = await searchAgentCommissionAccounts(value)
      setAccounts(rows)
      setSearched(true)
      setSelectedId(rows.length === 1 ? rows[0].id : null)
      setAmount('')
    } catch (error) {
      toast.error(getServerErrorMessage(error))
    } finally {
      setSearching(false)
    }
  }

  async function deduct(quota: number) {
    if (!selected) {
      toast.error(t('Select a user before deducting commission'))
      return
    }
    const trimmedNote = note.trim()
    if (!trimmedNote) {
      toast.error(t('Enter a deduction note'))
      return
    }
    if (!Number.isFinite(quota) || quota <= 0) {
      toast.error(t('Deduct amount must be greater than 0'))
      return
    }
    if (quota > selected.availableQuota) {
      toast.error(t('Available commission is insufficient'))
      return
    }
    setDeducting(true)
    try {
      const result = await deductAgentCommission({
        userId: selected.id,
        quota: Math.round(quota),
        note: trimmedNote,
      })
      setAccounts((current) =>
        current.map((account) =>
          account.id === selected.id
            ? {
                ...account,
                totalQuota: result.totalQuota,
                transferredQuota: result.transferredQuota,
                deductedQuota: result.deductedQuota,
                availableQuota: result.availableQuota,
              }
            : account
        )
      )
      setAmount('')
      await queryClient.invalidateQueries({
        queryKey: ['agent-commission-deductions'],
      })
      await queryClient.invalidateQueries({
        queryKey: ['approved-agents'],
      })
      toast.success(t('Commission deducted'))
    } catch (error) {
      toast.error(getServerErrorMessage(error))
    } finally {
      setDeducting(false)
    }
  }

  function quotaFromInput() {
    if (!selected) return 0
    if (
      amount === String(quotaUnitsToEditableAmount(selected.availableQuota))
    ) {
      return selected.availableQuota
    }
    return parseQuotaFromDollars(Number.parseFloat(amount))
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>
        {t('Settle commission')}
      </SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <AdminApprovedAgents />
        <section className='bg-card space-y-5 rounded-xl border p-4 sm:p-5'>
          <div className='flex items-start gap-3'>
            <div className='flex size-9 shrink-0 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
              <BadgePercent className='size-4' />
            </div>
            <div>
              <h3 className='font-semibold'>{t('Settle commission')}</h3>
              <p className='text-muted-foreground mt-1 text-sm'>
                {t(
                  "After customer service pays by WeChat or Alipay, deduct that commission here. Deducted commission is not added to the user's balance."
                )}
              </p>
            </div>
          </div>

          <form
            className='flex flex-wrap items-end gap-2'
            onSubmit={(event) => {
              event.preventDefault()
              void search()
            }}
          >
            <div className='min-w-64 flex-1 space-y-2'>
              <Label htmlFor='agent-commission-keyword'>
                {t('Username or user ID')}
              </Label>
              <Input
                id='agent-commission-keyword'
                value={keyword}
                onChange={(event) => setKeyword(event.target.value)}
                placeholder={t('Username or user ID')}
              />
            </div>
            <Button type='submit' disabled={searching}>
              {searching ? t('Processing...') : t('Search')}
            </Button>
          </form>

          {searched && accounts.length === 0 ? (
            <p className='text-muted-foreground text-sm'>
              {t('No matching users')}
            </p>
          ) : null}

          {accounts.length > 0 ? (
            <div className='grid gap-2'>
              {accounts.map((account) => {
                const active = account.id === selectedId
                return (
                  <button
                    key={account.id}
                    type='button'
                    className={cn(
                      'rounded-lg border px-3 py-3 text-left',
                      active && 'border-primary'
                    )}
                    onClick={() => {
                      setSelectedId(account.id)
                      setAmount('')
                    }}
                  >
                    <div className='flex flex-wrap items-baseline justify-between gap-2'>
                      <span className='font-medium'>
                        {account.displayName || account.username}
                        <span className='text-muted-foreground ml-2 text-sm font-normal'>
                          {account.username} · #{account.id}
                        </span>
                      </span>
                      <span className='text-sm tabular-nums'>
                        {t('Available commission')}{' '}
                        {formatQuota(account.availableQuota)}
                      </span>
                    </div>
                    <p className='text-muted-foreground mt-1 text-xs tabular-nums'>
                      {t('Total commission')} {formatQuota(account.totalQuota)}
                      {' · '}
                      {t('Transferred to balance')}{' '}
                      {formatQuota(account.transferredQuota)}
                      {' · '}
                      {t('Settled by support')}{' '}
                      {formatQuota(account.deductedQuota)}
                      {' · '}
                      {t('Level 1 rate')} {account.level1Rate}%
                      {account.level1Custom ? ` (${t('Custom rate')})` : ''}
                      {' · '}
                      {t('Level 2 rate')} {account.level2Rate}%
                      {account.level2Custom ? ` (${t('Custom rate')})` : ''}
                    </p>
                  </button>
                )
              })}
            </div>
          ) : null}

          {selected ? (
            <>
              <form
                className='space-y-3 border-t pt-4'
                onSubmit={(event) => {
                  event.preventDefault()
                  void deduct(quotaFromInput())
                }}
              >
                <p className='text-sm'>
                  {t('This does not add quota to the user.')}{' '}
                  {t('Available commission')}{' '}
                  {formatQuota(selected.availableQuota)}
                </p>
                <div className='grid gap-3 md:grid-cols-2'>
                  <div className='space-y-2'>
                    <Label htmlFor='agent-commission-amount'>
                      {t('Amount to deduct')}
                    </Label>
                    <Input
                      id='agent-commission-amount'
                      type='number'
                      min='0'
                      step='any'
                      value={amount}
                      onChange={(event) => setAmount(event.target.value)}
                    />
                  </div>
                  <div className='space-y-2'>
                    <Label htmlFor='agent-commission-note'>
                      {t('Deduction note')}
                    </Label>
                    <Textarea
                      id='agent-commission-note'
                      value={note}
                      maxLength={200}
                      placeholder={t('For example, paid by WeChat')}
                      onChange={(event) => setNote(event.target.value)}
                    />
                  </div>
                </div>
                <div className='flex flex-wrap gap-2'>
                  <Button
                    type='submit'
                    disabled={deducting || selected.availableQuota <= 0}
                  >
                    {deducting ? t('Processing...') : t('Deduct commission')}
                  </Button>
                  <Button
                    type='button'
                    variant='outline'
                    disabled={deducting || selected.availableQuota <= 0}
                    onClick={() => void deduct(selected.availableQuota)}
                  >
                    {t('Deduct all available commission')}
                  </Button>
                </div>
              </form>
              <AdminRateForm
                account={selected}
                onSaved={(update) => {
                  setAccounts((current) =>
                    current.map((account) =>
                      account.id === update.userId
                        ? {
                            ...account,
                            level1Rate: update.level1Rate,
                            level2Rate: update.level2Rate,
                            level1Custom: update.level1Custom,
                            level2Custom: update.level2Custom,
                          }
                        : account
                    )
                  )
                  void queryClient.invalidateQueries({
                    queryKey: ['approved-agents'],
                  })
                }}
              />
            </>
          ) : null}
        </section>

        <section className='bg-card mt-4 rounded-xl border p-4 sm:p-5'>
          <h3 className='font-semibold'>{t('Recent commission deductions')}</h3>
          {deductions.isError ? (
            <p className='mt-3 text-sm text-red-600 dark:text-red-400'>
              {t('Failed to load commission deductions')}
            </p>
          ) : null}
          {!deductions.isLoading && (deductions.data?.length ?? 0) === 0 ? (
            <p className='text-muted-foreground mt-3 text-sm'>
              {t('No commission deductions yet')}
            </p>
          ) : null}
          <div className='mt-3 divide-y'>
            {deductions.data?.map((row) => (
              <div key={row.id} className='py-3 text-sm'>
                <div className='flex flex-wrap items-baseline justify-between gap-2'>
                  <span className='font-medium'>
                    {row.username || `#${row.userId}`}
                    <span className='text-muted-foreground ml-2 font-normal'>
                      {formatQuota(row.quota)}
                    </span>
                  </span>
                  <span className='text-muted-foreground text-xs'>
                    {formatTimestamp(row.createdAt)}
                  </span>
                </div>
                <p className='mt-1'>{row.note}</p>
                <p className='text-muted-foreground mt-1 text-xs'>
                  {t('Settled by')} {row.adminUsername || `#${row.adminId}`}
                </p>
              </div>
            ))}
          </div>
        </section>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
