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

import { Copy } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { generateAffiliateLink } from '@/features/wallet/lib/affiliate'
import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { formatQuota } from '@/lib/format'

import type { AgentSystemReport } from '../types'
import { CommissionWithdrawal } from './commission-withdrawal'
import { MetricCard } from './metric'

export function OverviewPanel(props: { report?: AgentSystemReport }) {
  const { t } = useTranslation()
  const { copyToClipboard } = useCopyToClipboard()
  const report = props.report
  const inviteCode = report?.inviteCode ?? ''
  const inviteLink = inviteCode ? generateAffiliateLink(inviteCode) : ''
  const level1 = report == null ? '—' : String(report.level1Count)
  const level2 = report == null ? '—' : String(report.level2Count)
  const money = (quota?: number) =>
    report == null || quota == null ? '—' : formatQuota(quota)
  const count = (value?: number) =>
    report == null || value == null ? '—' : String(value)

  return (
    <div className='space-y-4'>
      <div className='grid gap-6 lg:grid-cols-2'>
        <section className='space-y-3'>
          <h3 className='text-sm font-medium'>{t('Invite code')}</h3>
          <p className='text-2xl font-semibold tracking-wide break-all'>
            {inviteCode || t('Invite code is unavailable')}
          </p>
          <p className='text-muted-foreground text-sm break-all'>
            {inviteLink || '—'}
          </p>
          <div className='flex flex-wrap gap-2'>
            <Button
              type='button'
              size='sm'
              disabled={!inviteCode}
              onClick={() => copyToClipboard(inviteCode)}
            >
              <Copy />
              {t('Copy invite code')}
            </Button>
            <Button
              type='button'
              size='sm'
              variant='outline'
              disabled={!inviteLink}
              onClick={() => copyToClipboard(inviteLink)}
            >
              <Copy />
              {t('Copy invite link')}
            </Button>
          </div>
        </section>

        <section>
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <h3 className='text-sm font-medium'>
              {t('Commission management')}
            </h3>
          </div>
          <dl className='mt-4 grid grid-cols-2 gap-x-6 gap-y-4'>
            <Figure
              label={t('Available commission')}
              value={money(report?.availableQuota)}
            />
            <Figure
              label={t('Total commission')}
              value={money(report?.totalQuota)}
            />
            <Figure
              label={t('Transferred to balance')}
              value={money(report?.transferredQuota)}
            />
            <Figure
              label={t('Settled by support')}
              value={money(report?.deductedQuota)}
            />
            <Figure label={t('Level 1 users')} value={level1} />
            <Figure label={t('Level 2 users')} value={level2} />
          </dl>
          <div className='mt-4 flex flex-wrap items-center justify-between gap-3 border-t pt-4'>
            <p className='text-muted-foreground text-sm'>
              {t(
                'Request a withdrawal to your bound Alipay account. The amount is reserved while the request is pending and released if it is rejected.'
              )}
            </p>
            <div className='flex flex-wrap gap-2'>
              <CommissionWithdrawal
                availableQuota={report?.availableQuota ?? 0}
                minimumQuota={report?.minimumTransferQuota ?? 0}
                alipayAccount={report?.alipayAccount ?? ''}
                pendingQuota={report?.pendingWithdrawalQuota ?? 0}
              />
            </div>
          </div>
        </section>
      </div>

      <div className='grid gap-3 md:grid-cols-3'>
        <MetricCard
          label={t('Level 1 users')}
          value={level1}
          hint={t('Users who registered with your invite code.')}
        />
        <MetricCard
          label={t('Level 1 recharge amount')}
          value={money(report?.level1RechargeQuota)}
        />
        <MetricCard
          label={t('Level 1 commission orders')}
          value={count(report?.level1OrderCount)}
        />
        <MetricCard
          label={t('Level 2 users')}
          value={level2}
          hint={t('Users invited by your level 1 users.')}
        />
        <MetricCard
          label={t('Level 2 recharge amount')}
          value={money(report?.level2RechargeQuota)}
        />
        <MetricCard
          label={t('Level 2 commission orders')}
          value={count(report?.level2OrderCount)}
        />
      </div>

      <div className='grid gap-3 md:grid-cols-2'>
        <MetricCard
          label={t('Level 1 rate')}
          value={report == null ? '—' : `${report.level1Rate.toFixed(2)}%`}
          hint={report?.level1RateCustom ? t('Custom rate') : t('Default rate')}
        />
        <MetricCard
          label={t('Level 2 rate')}
          value={report == null ? '—' : `${report.level2Rate.toFixed(2)}%`}
          hint={report?.level2RateCustom ? t('Custom rate') : t('Default rate')}
        />
      </div>
      <p className='text-muted-foreground text-sm'>
        {t(
          'Level 1 commission is {{level1}}% of successful recharges from directly invited users. Level 2 commission is {{level2}}%. The old invitation quota is not included. A later rate change does not rewrite older orders.',
          {
            level1: report?.level1Rate ?? 15,
            level2: report?.level2Rate ?? 5,
          }
        )}
      </p>
    </div>
  )
}

function Figure(props: { label: string; value: string; hint?: string }) {
  return (
    <div>
      <dt className='text-muted-foreground text-sm'>{props.label}</dt>
      <dd className='mt-1 text-xl font-semibold tabular-nums'>{props.value}</dd>
      {props.hint ? (
        <p className='text-muted-foreground mt-1 text-xs'>{props.hint}</p>
      ) : null}
    </div>
  )
}
