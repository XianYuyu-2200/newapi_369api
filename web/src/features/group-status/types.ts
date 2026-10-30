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

export type GroupStatusBucket = {
  success: number
  error: number
}

export type GroupStatusModel = {
  model: string
  requests: number
  success: number
  error: number
  availability: number | null
  avgUseTime: number | null
  avgFirstToken: number | null
  cacheRate: number | null
}

export type GroupStatusGroup = {
  name: string
  description: string
  ratio: number
  modelCount: number
  enabled: boolean
  requests: number
  success: number
  error: number
  availability: number | null
  avgUseTime: number | null
  avgFirstToken: number | null
  cacheRate: number | null
  latestAt: number | null
  modelsWithoutRequests: number
  history: GroupStatusBucket[]
  models: GroupStatusModel[]
}

export type GroupStatusReport = {
  windowSeconds: number
  bucketSeconds: number
  generatedAt: number
  groups: GroupStatusGroup[]
}
