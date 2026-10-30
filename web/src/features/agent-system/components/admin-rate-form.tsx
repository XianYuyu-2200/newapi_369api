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

import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { getServerErrorMessage } from '@/lib/server-error-message'

import { setAgentCommissionRates } from '../api'
import type { AgentCommissionAdminAccount, AgentRateUpdate } from '../types'

export function AdminRateForm(props: {
  account: AgentCommissionAdminAccount
  onSaved: (update: AgentRateUpdate) => void
}) {
  const { t } = useTranslation()
  const { account } = props
  const [level1, setLevel1] = useState(String(account.level1Rate))
  const [level2, setLevel2] = useState(String(account.level2Rate))
  const [level1Default, setLevel1Default] = useState(!account.level1Custom)
  const [level2Default, setLevel2Default] = useState(!account.level2Custom)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setLevel1(String(account.level1Rate))
    setLevel2(String(account.level2Rate))
    setLevel1Default(!account.level1Custom)
    setLevel2Default(!account.level2Custom)
  }, [
    account.id,
    account.level1Rate,
    account.level2Rate,
    account.level1Custom,
    account.level2Custom,
  ])

  function rateValue(raw: string, useDefault: boolean) {
    if (useDefault) return -1
    if (!/^\d+$/.test(raw.trim())) return null
    const parsed = Number.parseInt(raw.trim(), 10)
    if (parsed < 0 || parsed > 100) return null
    return parsed
  }

  async function save() {
    const level1Rate = rateValue(level1, level1Default)
    const level2Rate = rateValue(level2, level2Default)
    if (level1Rate == null || level2Rate == null) {
      toast.error(t('Enter a whole number from 0 to 100'))
      return
    }
    setSaving(true)
    try {
      const update = await setAgentCommissionRates({
        userId: account.id,
        level1Rate,
        level2Rate,
      })
      props.onSaved(update)
      toast.success(t('Commission rates saved'))
    } catch (error) {
      toast.error(getServerErrorMessage(error))
    } finally {
      setSaving(false)
    }
  }

  return (
    <form
      className='space-y-3 border-t pt-4'
      onSubmit={(event) => {
        event.preventDefault()
        void save()
      }}
    >
      <div>
        <h4 className='text-sm font-medium'>{t('Custom commission rates')}</h4>
        <p className='text-muted-foreground mt-1 text-sm'>
          {t('Rates apply only to recharges completed after they are saved.')}
        </p>
      </div>
      <div className='grid gap-3 md:grid-cols-2'>
        <RateField
          id='agent-level1-rate'
          label={t('Level 1 rate')}
          value={level1}
          useDefault={level1Default}
          custom={account.level1Custom}
          onValueChange={setLevel1}
          onDefaultChange={setLevel1Default}
        />
        <RateField
          id='agent-level2-rate'
          label={t('Level 2 rate')}
          value={level2}
          useDefault={level2Default}
          custom={account.level2Custom}
          onValueChange={setLevel2}
          onDefaultChange={setLevel2Default}
        />
      </div>
      <Button type='submit' disabled={saving}>
        {saving ? t('Processing...') : t('Save commission rates')}
      </Button>
    </form>
  )
}

function RateField(props: {
  id: string
  label: string
  value: string
  useDefault: boolean
  custom: boolean
  onValueChange: (value: string) => void
  onDefaultChange: (value: boolean) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='space-y-2'>
      <Label htmlFor={props.id}>
        {props.label}
        <span className='text-muted-foreground ml-2 text-xs font-normal'>
          {props.custom ? t('Custom rate') : t('Default rate')}
        </span>
      </Label>
      <Input
        id={props.id}
        inputMode='numeric'
        value={props.useDefault ? '' : props.value}
        disabled={props.useDefault}
        placeholder={props.useDefault ? t('Default rate') : '0-100'}
        onChange={(event) => props.onValueChange(event.target.value)}
      />
      <label className='text-muted-foreground flex items-center gap-2 text-sm'>
        <Checkbox
          checked={props.useDefault}
          onCheckedChange={(checked) => props.onDefaultChange(checked === true)}
        />
        {t('Use the default rate')}
      </label>
    </div>
  )
}
