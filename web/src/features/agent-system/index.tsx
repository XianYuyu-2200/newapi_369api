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

import { useIsFetching, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  BadgePercent,
  BarChart3,
  Receipt,
  RefreshCw,
  Share2,
  Users,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { SectionPageLayout } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'

import { getAgentSystem } from './api'
import { OverviewPanel } from './components/overview-panel'
import { RecordsPanel } from './components/records-panel'
import { StatsPanel } from './components/stats-panel'
import { UsersPanel } from './components/users-panel'

const tabs = [
  { id: 'overview', icon: Share2, label: 'Agent system' },
  { id: 'stats', icon: BarChart3, label: 'Statistics' },
  { id: 'users', icon: Users, label: 'Invited users' },
  { id: 'recharge', icon: Receipt, label: 'Recharge records' },
  { id: 'commission', icon: BadgePercent, label: 'Commission records' },
] as const

export function AgentSystemPage() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const fetching = useIsFetching({ queryKey: ['agent-system'] }) > 0
  const [tab, setTab] = useState<(typeof tabs)[number]['id']>('overview')
  const query = useQuery({
    queryKey: ['agent-system'],
    queryFn: getAgentSystem,
  })

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Agent system')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <Button
          type='button'
          variant='outline'
          size='sm'
          onClick={() =>
            void queryClient.invalidateQueries({ queryKey: ['agent-system'] })
          }
          disabled={fetching}
        >
          <RefreshCw className={cn('size-4', fetching && 'animate-spin')} />
          {t('Refresh')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <section className='bg-card rounded-xl border'>
          <div className='flex items-start gap-3 px-4 pt-4 sm:px-5'>
            <div className='flex size-9 shrink-0 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-600 dark:text-emerald-400'>
              <Share2 className='size-4' />
            </div>
            <div className='min-w-0'>
              <h3 className='font-semibold'>{t('Agent system')}</h3>
              <p className='text-muted-foreground mt-1 text-sm'>
                {t(
                  'View your invite code and level 1 and level 2 users. Level 1 commission is {{level1}}% and level 2 commission is {{level2}}%. Available commission can be transferred to your balance.',
                  {
                    level1: query.data?.level1Rate ?? 15,
                    level2: query.data?.level2Rate ?? 5,
                  }
                )}
              </p>
            </div>
          </div>
          {query.data ? (
            <Tabs
              value={tab}
              onValueChange={(value) =>
                setTab(value as (typeof tabs)[number]['id'])
              }
              className='gap-4 px-4 pt-4 pb-5 sm:px-5'
            >
              <div className='overflow-x-auto'>
                <TabsList variant='line' className='min-w-full justify-start'>
                  {tabs.map((item) => (
                    <TabsTrigger key={item.id} value={item.id} className='px-3'>
                      <item.icon className='size-4' />
                      {t(item.label)}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </div>
              <TabsContent value='overview'>
                {query.isLoading ? (
                  <Skeleton className='h-80 rounded-xl' />
                ) : (
                  <div className='space-y-3'>
                    {query.isError ? (
                      <p className='text-sm text-red-600 dark:text-red-400'>
                        {t('Failed to load agent system')}
                      </p>
                    ) : null}
                    <OverviewPanel report={query.data} />
                  </div>
                )}
              </TabsContent>
              <TabsContent value='stats'>
                <StatsPanel />
              </TabsContent>
              <TabsContent value='users'>
                <UsersPanel />
              </TabsContent>
              <TabsContent value='recharge'>
                <RecordsPanel kind='recharge' />
              </TabsContent>
              <TabsContent value='commission'>
                <RecordsPanel kind='commission' />
              </TabsContent>
            </Tabs>
          ) : null}
        </section>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
