import { useNavigate, useRouter } from '@tanstack/react-router'
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
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const FEEDBACK_URL = 'https://github.com/QuantumNous/new-api/issues'

type GeneralErrorProps = React.HTMLAttributes<HTMLDivElement> & {
  minimal?: boolean
  error?: unknown
}

const CHUNK_RELOAD_KEY = 'newapi-chunk-reload-at'
const CHUNK_RELOAD_WINDOW_MS = 60_000

function isChunkLoadError(error: unknown): boolean {
  if (typeof error !== 'object' || error === null) return false
  const record = error as {
    name?: unknown
    message?: unknown
    code?: unknown
    cause?: unknown
  }
  const name = typeof record.name === 'string' ? record.name : ''
  const message = typeof record.message === 'string' ? record.message : ''
  const code = typeof record.code === 'string' ? record.code : ''
  if (
    record.cause &&
    record.cause !== error &&
    isChunkLoadError(record.cause)
  ) {
    return true
  }
  return (
    name === 'ChunkLoadError' ||
    code === 'CSS_CHUNK_LOAD_FAILED' ||
    /Loading chunk \d+ failed|Failed to fetch dynamically imported module|Importing a module script failed|error loading dynamically imported module|Loading CSS chunk/i.test(
      message
    )
  )
}

function reloadOnceForChunkError(): boolean {
  try {
    const last = Number(sessionStorage.getItem(CHUNK_RELOAD_KEY) || 0)
    if (Number.isFinite(last) && Date.now() - last < CHUNK_RELOAD_WINDOW_MS) {
      return false
    }
    sessionStorage.setItem(CHUNK_RELOAD_KEY, String(Date.now()))
  } catch {
    return false
  }
  window.location.reload()
  return true
}

function getHttpStatus(error: unknown): number | undefined {
  if (typeof error !== 'object' || error === null) return undefined
  const response = (error as Record<string, unknown>).response
  if (typeof response !== 'object' || response === null) return undefined
  const status = (response as Record<string, unknown>).status
  return typeof status === 'number' ? status : undefined
}

export function GeneralError({
  className,
  minimal = false,
  error,
}: GeneralErrorProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { history } = useRouter()
  const status = getHttpStatus(error)
  const chunkFailed = isChunkLoadError(error)
  const isRateLimited = status === 429
  let title = `${t('Oops! Something went wrong')} ${`:')`}`
  let description = t('Please try again later.')
  if (chunkFailed) {
    title = t('Loading...')
    description = t('Please try again later.')
  } else if (isRateLimited) {
    title = t('Too many requests')
    description = t('Please wait a moment before trying again.')
  }

  useEffect(() => {
    if (!chunkFailed) return
    reloadOnceForChunkError()
  }, [chunkFailed])

  return (
    <div className={cn('h-svh w-full', className)}>
      <div className='m-auto flex h-full w-full flex-col items-center justify-center gap-2'>
        {!minimal && !chunkFailed && (
          <h1 className='text-[7rem] leading-tight font-bold'>
            {status ?? 500}
          </h1>
        )}
        <span className='font-medium'>{title}</span>
        <p className='text-muted-foreground text-center'>
          {t('We apologize for the inconvenience.')} <br /> {description}
        </p>
        {!minimal && (
          <p className='text-muted-foreground text-center text-sm'>
            {t('If this keeps happening, please report it on GitHub Issues.')}
          </p>
        )}
        {!minimal && (
          <div className='mt-6 flex flex-wrap justify-center gap-4'>
            <Button variant='outline' onClick={() => history.go(-1)}>
              {t('Go Back')}
            </Button>
            {chunkFailed && (
              <Button
                variant='outline'
                onClick={() => window.location.reload()}
              >
                {t('Retry')}
              </Button>
            )}
            <Button
              variant='outline'
              render={
                <a
                  href={FEEDBACK_URL}
                  target='_blank'
                  rel='noopener noreferrer'
                />
              }
            >
              {t('Report an issue')}
            </Button>
            <Button onClick={() => navigate({ to: '/' })}>
              {t('Back to Home')}
            </Button>
          </div>
        )}
      </div>
    </div>
  )
}
