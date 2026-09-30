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
import {
  Layers,
  MessagesSquare,
  PlugZap,
  type LucideIcon,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'

import { CLIENT_INTEGRATIONS } from '../constants'
import type { ClientIntegration, ClientIntegrationIconId } from '../types'

/**
 * Map of `ClientIntegration.id` → lucide icon. Lives in the component
 * module (not `constants.ts`) so the constants file stays JSX-free and
 * can be imported from non-TSX contexts.
 */
const CLIENT_ICONS: Record<ClientIntegrationIconId, LucideIcon> = {
  codebuddy: PlugZap,
  chatbox: MessagesSquare,
  cherryStudio: Layers,
}

/**
 * Section that explains how to wire New API into a few popular
 * OpenAI-compatible clients. Each card pulls its name, description and
 * onboarding steps from the translation files; only the Base URL is
 * resolved at runtime so the snippet always matches the deployment the
 * reader is reading from.
 */
export function ClientIntegrationSection() {
  const { t } = useTranslation()
  const baseUrl = resolveBaseUrl()

  return (
    <section className='relative z-10 px-4 py-12 md:px-6 md:py-16'>
      <AnimateInView>
        <div className='mx-auto max-w-[1400px]'>
          <header className='mb-8 max-w-2xl'>
            <p className='text-blue-600 dark:text-blue-400 mb-1 text-sm font-semibold tracking-wider uppercase'>
              {t('docs.clients.eyebrow')}
            </p>
            <h2 className='text-foreground text-2xl font-bold tracking-tight sm:text-3xl'>
              {t('docs.clients.title')}
            </h2>
            <p className='text-muted-foreground mt-3 text-base text-balance'>
              {t('docs.clients.subtitle')}
            </p>
          </header>

          <div className='grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3'>
            {CLIENT_INTEGRATIONS.map((integration) => (
              <ClientIntegrationCard
                key={integration.id}
                integration={integration}
                baseUrl={baseUrl}
              />
            ))}
          </div>
        </div>
      </AnimateInView>
    </section>
  )
}

function resolveBaseUrl(): string {
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin
  }
  return 'https://your-new-api-host.example'
}

function ClientIntegrationCard({
  integration,
  baseUrl,
}: {
  integration: ClientIntegration
  baseUrl: string
}) {
  const { t } = useTranslation()
  const apiBaseUrl = `${baseUrl}/v1`
  const Icon = CLIENT_ICONS[integration.id]

  return (
    <article className='border-border/60 bg-card flex flex-col gap-5 rounded-2xl border p-5 shadow-[0_24px_60px_-30px_rgba(15,23,42,0.35)]'>
      <header className='flex items-start gap-3'>
        <span className='bg-blue-600/10 text-blue-600 dark:bg-blue-400/15 dark:text-blue-300 inline-flex size-10 shrink-0 items-center justify-center rounded-xl'>
          <Icon className='size-5' strokeWidth={1.6} />
        </span>
        <div className='min-w-0'>
          <h3 className='text-foreground text-base font-semibold tracking-tight sm:text-lg'>
            {t(integration.nameKey)}
          </h3>
          <p className='text-muted-foreground mt-1 text-sm leading-relaxed'>
            {t(integration.descriptionKey)}
          </p>
        </div>
      </header>

      <ol className='space-y-3 border-t border-blue-100/60 pt-4 dark:border-blue-400/10'>
        {integration.steps.map((step, index) => (
          <li key={step.titleKey} className='flex gap-3'>
            <span className='bg-blue-600/10 text-blue-600 dark:bg-blue-400/15 dark:text-blue-300 mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full text-xs font-semibold'>
              {index + 1}
            </span>
            <div className='min-w-0'>
              <p className='text-foreground text-sm font-medium'>
                {t(step.titleKey)}
              </p>
              <p className='text-muted-foreground mt-1 text-xs leading-relaxed'>
                {t(step.descriptionKey)}
              </p>
            </div>
          </li>
        ))}
      </ol>

      <dl className='grid grid-cols-1 gap-2 rounded-xl border border-blue-100/60 bg-blue-50/30 p-3 text-sm dark:border-blue-400/10 dark:bg-blue-400/5'>
        <Field
          label={t(integration.baseUrlLabelKey)}
          value={apiBaseUrl}
          mono
        />
        <Field
          label={t('docs.clients.field.protocol')}
          value={t(integration.protocolKey)}
        />
        <Field
          label={t('docs.clients.field.defaultModel')}
          value={t(integration.defaultModelKey)}
          mono
        />
      </dl>
    </article>
  )
}

function Field({
  label,
  value,
  mono = false,
}: {
  label: string
  value: string
  mono?: boolean
}) {
  return (
    <div className='flex items-center justify-between gap-3'>
      <dt className='text-muted-foreground text-xs font-medium tracking-wide uppercase'>
        {label}
      </dt>
      <dd
        className={
          mono
            ? 'truncate rounded-md bg-background/70 px-2 py-1 font-mono text-[12px] text-foreground'
            : 'text-foreground truncate text-xs'
        }
      >
        {value}
      </dd>
    </div>
  )
}