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
import { useEffect, useRef, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'

import { fetchTokenKey, getApiKeys } from '@/features/keys/api'
import { API_KEY_STATUS } from '@/features/keys/constants'
import { getUserModels } from '@/features/playground/api'

const CANVAS_SRC = '/canvas-app/index.html'

type CanvasEmbedConfig = {
  type: '369api-embed-config'
  baseUrl: string
  apiKey: string
  models: string[]
}

export function InfiniteCanvas() {
  const { t } = useTranslation()
  const iframeRef = useRef<HTMLIFrameElement>(null)
  const configRef = useRef<CanvasEmbedConfig | null>(null)
  const [missingKey, setMissingKey] = useState(false)

  useEffect(() => {
    let cancelled = false

    void (async () => {
      const config = await loadCanvasEmbedConfig()
      if (cancelled) return
      configRef.current = config
      setMissingKey(config == null)
      postConfig(iframeRef.current, config)
    })()

    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className='flex h-full min-h-0 flex-1 flex-col'>
      {missingKey ? (
        <div className='border-b border-amber-500/30 bg-amber-500/10 px-4 py-2 text-sm text-amber-100'>
          {t(
            'No enabled API key yet. Create one under API Keys before the canvas can call this site.'
          )}{' '}
          <Link to='/keys' className='underline'>
            {t('API Keys')}
          </Link>
        </div>
      ) : null}
      <iframe
        ref={iframeRef}
        title={t('Infinite canvas')}
        src={CANVAS_SRC}
        className='h-full min-h-0 w-full flex-1 border-0 bg-[#0a0a0a]'
        onLoad={() => postConfig(iframeRef.current, configRef.current)}
      />
    </div>
  )
}

function postConfig(
  frame: HTMLIFrameElement | null,
  config: CanvasEmbedConfig | null
) {
  if (!frame?.contentWindow || !config) return
  frame.contentWindow.postMessage(config, window.location.origin)
}

async function loadCanvasEmbedConfig(): Promise<CanvasEmbedConfig | null> {
  try {
    const list = await getApiKeys({ p: 1, size: 100 })
    const now = Math.floor(Date.now() / 1000)
    const token = (list.data?.items ?? []).find((item) => {
      if (item.status !== API_KEY_STATUS.ENABLED) return false
      return !(item.expired_time > 0 && item.expired_time < now)
    })
    if (!token) return null

    const keyRes = await fetchTokenKey(token.id)
    const rawKey = keyRes.data?.key?.trim()
    if (!keyRes.success || !rawKey) return null
    const apiKey = rawKey.startsWith('sk-') ? rawKey : `sk-${rawKey}`

    return {
      type: '369api-embed-config',
      baseUrl: window.location.origin,
      apiKey,
      models: await loadModelNames(),
    }
  } catch {
    return null
  }
}

async function loadModelNames(): Promise<string[]> {
  try {
    // Empty group asks the relay for every model this account can use.
    const models = await getUserModels('')
    return models.map((model) => model.value).filter(Boolean)
  } catch {
    return []
  }
}
