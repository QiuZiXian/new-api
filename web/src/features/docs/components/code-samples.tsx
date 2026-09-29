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
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { AnimateInView } from '@/components/animate-in-view'
import {
  CodeBlock,
  CodeBlockCopyButton,
} from '@/components/ai-elements/code-block'
import { cn } from '@/lib/utils'

import { CODE_SAMPLES } from '../constants'

/**
 * Snippet templates keyed by `snippetId`. Each template is a function so
 * the caller can splice in the deployment-specific Base URL. The snippets
 * intentionally avoid library-specific defaults (timeouts, retry counts)
 * so they stay copy-pasteable.
 */
const SNIPPET_TEMPLATES: Record<'python' | 'javascript' | 'bash', string> = {
  python: `# pip install openai
from openai import OpenAI

BASE_URL = "{BASE_URL}"
client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url=f"{BASE_URL}/v1",
)

response = client.chat.completions.create(
    model="gpt-4o-mini",
    messages=[
        {"role": "user", "content": "Hello, who are you?"},
    ],
)

print(response.choices[0].message.content)
`,
  javascript: `// npm install openai
import OpenAI from 'openai';

const BASE_URL = '{BASE_URL}';

const client = new OpenAI({
  apiKey: 'YOUR_API_KEY',
  baseURL: \`\${BASE_URL}/v1\`,
});

const response = await client.chat.completions.create({
  model: 'gpt-4o-mini',
  messages: [
    { role: 'user', content: 'Hello, who are you?' },
  ],
});

console.log(response.choices[0].message.content);
`,
  bash: `BASE_URL="{BASE_URL}"

curl -X POST "$BASE_URL/v1/chat/completions" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4o-mini",
    "messages": [
      {"role": "user", "content": "Hello, who are you?"}
    ]
  }'
`,
}

function resolveBaseUrl(): string {
  if (typeof window !== 'undefined' && window.location?.origin) {
    return window.location.origin
  }
  return 'https://your-new-api-host.example'
}

/**
 * Code-samples block under Quick Start. Tabs swap between Python /
 * JavaScript / cURL; each tab renders the same request against the local
 * deployment's Base URL via the `CodeBlock` component.
 */
export function CodeSamples() {
  const { t } = useTranslation()
  const [activeId, setActiveId] = useState(CODE_SAMPLES[0].id)
  const active = CODE_SAMPLES.find((sample) => sample.id === activeId) ?? CODE_SAMPLES[0]

  const renderedCode = useMemo(() => {
    const template = SNIPPET_TEMPLATES[active.snippetId]
    return template.replaceAll('{BASE_URL}', resolveBaseUrl())
  }, [active.snippetId])

  return (
    <section className='relative z-10 px-4 py-12 md:px-6 md:py-16'>
      <AnimateInView>
        <div className='mx-auto max-w-[1400px]'>
          <header className='mb-6 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between'>
            <div className='max-w-2xl'>
              <p className='text-blue-600 dark:text-blue-400 mb-1 text-sm font-semibold tracking-wider uppercase'>
                {t('docs.codeSamples.eyebrow')}
              </p>
              <h2 className='text-foreground text-2xl font-bold tracking-tight sm:text-3xl'>
                {t('docs.codeSamples.title')}
              </h2>
              <p className='text-muted-foreground mt-2 text-sm sm:text-base'>
                {t('docs.codeSamples.subtitle')}
              </p>
            </div>
            <p className='text-muted-foreground/70 hidden text-xs sm:block'>
              {t('docs.codeSamples.copyHint')}
            </p>
          </header>

          <div className='border-border/60 bg-card overflow-hidden rounded-2xl border shadow-[0_30px_80px_-50px_rgba(15,23,42,0.45)]'>
            <div
              role='tablist'
              aria-label={t('docs.codeSamples.title')}
              className='border-border/60 bg-muted/30 flex items-center gap-1 border-b px-2 py-2'
            >
              {CODE_SAMPLES.map((sample) => {
                const isActive = sample.id === active.id
                return (
                  <button
                    key={sample.id}
                    role='tab'
                    type='button'
                    aria-selected={isActive}
                    aria-controls={`docs-code-panel-${sample.id}`}
                    id={`docs-code-tab-${sample.id}`}
                    onClick={() => setActiveId(sample.id)}
                    className={cn(
                      'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                      isActive
                        ? 'bg-background text-foreground shadow-xs'
                        : 'text-muted-foreground hover:text-foreground hover:bg-background/60'
                    )}
                  >
                    {t(sample.labelKey)}
                  </button>
                )
              })}
            </div>

            <div
              role='tabpanel'
              id={`docs-code-panel-${active.id}`}
              aria-labelledby={`docs-code-tab-${active.id}`}
              className='bg-muted/10 p-2 sm:p-3'
            >
              <CodeBlock
                code={renderedCode}
                language={active.language}
                showToolbar
              >
                <CodeBlockCopyButton />
              </CodeBlock>
            </div>
          </div>
        </div>
      </AnimateInView>
    </section>
  )
}