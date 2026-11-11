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
import { ArrowUpRight, Code2, Copy } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useCopyToClipboard } from '@/hooks/use-copy-to-clipboard'
import { cn } from '@/lib/utils'

const CHAT_COMPLETIONS_PATH = '/v1/chat/completions'

interface PricingApiBarProps {
  docsLink?: string
  className?: string
}

export function PricingApiBar(props: PricingApiBarProps) {
  const { t } = useTranslation()
  const { copyToClipboard } = useCopyToClipboard()
  const fullUrl =
    typeof window === 'undefined'
      ? CHAT_COMPLETIONS_PATH
      : `${window.location.origin}${CHAT_COMPLETIONS_PATH}`
  const docsHref = props.docsLink || '/docs'

  return (
    <div
      className={cn(
        'mt-6 flex flex-wrap items-center gap-x-3 gap-y-2 rounded-xl border border-blue-100 bg-[linear-gradient(90deg,#f2f7ff,#f8fbff)] px-4 py-3 dark:border-white/10 dark:bg-[linear-gradient(90deg,rgba(30,58,138,0.18),rgba(15,23,42,0.12))]',
        props.className
      )}
    >
      <div className='flex shrink-0 items-center gap-2'>
        <Code2 className='size-4 text-blue-600 dark:text-blue-400' />
        <span className='text-foreground text-sm font-semibold'>
          {t('Unified API Access')}
        </span>
      </div>

      <div className='border-border/70 bg-background/80 flex min-w-0 flex-1 items-center gap-2 rounded-md border px-2.5 py-1.5'>
        <span className='shrink-0 rounded bg-blue-600/10 px-1.5 py-0.5 text-[10px] font-semibold text-blue-700 dark:text-blue-300'>
          POST
        </span>
        <code className='text-muted-foreground truncate font-mono text-xs'>
          {fullUrl}
        </code>
        <button
          type='button'
          aria-label={t('Copy')}
          title={t('Copy')}
          onClick={() => copyToClipboard(fullUrl)}
          className='text-muted-foreground hover:text-foreground ml-auto shrink-0 rounded p-1 transition-colors'
        >
          <Copy className='size-3.5' />
        </button>
      </div>

      <a
        href={docsHref}
        target='_blank'
        rel='noopener noreferrer'
        className='inline-flex shrink-0 items-center gap-1 text-xs font-medium text-blue-600 hover:underline dark:text-blue-400'
      >
        {t('View full developer docs')}
        <ArrowUpRight className='size-3.5' />
      </a>
    </div>
  )
}
