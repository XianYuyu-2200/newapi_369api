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

import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { formatQuota, parseQuotaFromDollars, quotaUnitsToDollars } from '@/lib/format'
import { getServerErrorMessage } from '@/lib/server-error-message'
import { useAuthStore } from '@/stores/auth-store'

import { transferAgentCommission } from '../api'

export function CommissionTransfer(props: {
  availableQuota: number
  minimumQuota: number
}) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)
  const [amount, setAmount] = useState('')
  const [pending, setPending] = useState(false)
  const canTransfer = props.availableQuota >= props.minimumQuota && props.minimumQuota > 0

  async function submit(quota: number) {
    if (!Number.isFinite(quota) || quota <= 0) {
      toast.error(t('Enter a transfer amount'))
      return
    }
    setPending(true)
    try {
      const result = await transferAgentCommission(Math.round(quota))
      const auth = useAuthStore.getState().auth
      if (auth.user) {
        auth.setUser({
          ...auth.user,
          quota: result.balanceQuota,
        })
      }
      await queryClient.invalidateQueries({ queryKey: ['agent-system'] })
      toast.success(t('Commission transferred to your balance'))
      setOpen(false)
    } catch (error) {
      toast.error(error instanceof Error ? error.message : getServerErrorMessage(error))
    } finally {
      setPending(false)
    }
  }

  return (
    <>
      <Button
        type='button'
        size='sm'
        disabled={!canTransfer}
        onClick={() => {
          setAmount(formatTransferInput(props.availableQuota))
          setOpen(true)
        }}
      >
        {t('Transfer commission to balance')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('Transfer commission to balance')}</DialogTitle>
            <DialogDescription>
              {t('Available commission {{amount}}', {
                amount: formatQuota(props.availableQuota),
              })}
              {' · '}
              {t('Minimum transfer {{amount}}', {
                amount: formatQuota(props.minimumQuota),
              })}
            </DialogDescription>
          </DialogHeader>
          <label className='grid gap-2'>
            <span className='text-sm font-medium'>{t('Amount to transfer')}</span>
            <Input
              inputMode='decimal'
              value={amount}
              onChange={(event) => setAmount(event.target.value)}
            />
          </label>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              disabled={pending || !canTransfer}
              onClick={() => void submit(props.availableQuota)}
            >
              {t('Transfer all available commission')}
            </Button>
            <Button
              type='button'
              disabled={pending}
              onClick={() => void submit(parseQuotaFromDollars(Number(amount)))}
            >
              {t('Transfer commission to balance')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

function formatTransferInput(quota: number) {
  const value = quotaUnitsToDollars(quota)
  if (!Number.isFinite(value) || value <= 0) return ''
  return String(Math.round(value * 10000) / 10000)
}
