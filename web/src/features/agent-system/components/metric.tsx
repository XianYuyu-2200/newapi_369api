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

import { cn } from '@/lib/utils'

export function MetricCard(props: {
  label: string
  value: string
  hint?: string
  valueClassName?: string
}) {
  return (
    <div className='bg-card rounded-lg border px-4 py-3'>
      <div className='text-muted-foreground text-sm'>{props.label}</div>
      <div
        className={cn(
          'mt-2 text-2xl font-semibold tracking-tight tabular-nums',
          props.valueClassName
        )}
      >
        {props.value}
      </div>
      {props.hint ? (
        <p className='text-muted-foreground mt-2 text-xs'>{props.hint}</p>
      ) : null}
    </div>
  )
}

export function UnavailableFigure(props: { label: string }) {
  return <span className='text-muted-foreground'>{props.label}</span>
}
