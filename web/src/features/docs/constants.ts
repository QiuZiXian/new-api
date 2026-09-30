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
import type {
  ClientIntegration,
  CodeSample,
  FeatureGridItem,
  QuickStartStep,
} from './types'

/**
 * Hero badge text shown above the page title.
 */
export const HERO_BADGE_KEY = 'docs.hero.badge'

/**
 * Three-step onboarding shown in the Quick Start section. Copy lives in the
 * translation files; this module only owns the shape and ordering.
 */
export const QUICK_START_STEPS: QuickStartStep[] = [
  {
    number: 1,
    titleKey: 'docs.quickStart.step1.title',
    descriptionKey: 'docs.quickStart.step1.description',
  },
  {
    number: 2,
    titleKey: 'docs.quickStart.step2.title',
    descriptionKey: 'docs.quickStart.step2.description',
  },
  {
    number: 3,
    titleKey: 'docs.quickStart.step3.title',
    descriptionKey: 'docs.quickStart.step3.description',
  },
]

/**
 * Code samples for the unified `/v1/chat/completions` endpoint. Each entry
 * pairs an i18n tab label with a snippet identifier; the renderer resolves
 * the identifier to a concrete template at runtime so translation files
 * stay free of large, escape-heavy code blocks.
 */
export const CODE_SAMPLES: CodeSample[] = [
  {
    id: 'python',
    labelKey: 'docs.codeSamples.tabs.python',
    language: 'python',
    snippetId: 'python',
  },
  {
    id: 'javascript',
    labelKey: 'docs.codeSamples.tabs.javascript',
    language: 'javascript',
    snippetId: 'javascript',
  },
  {
    id: 'bash',
    labelKey: 'docs.codeSamples.tabs.bash',
    language: 'bash',
    snippetId: 'bash',
  },
]

/**
 * Third-party clients we test against. The order in this array is the
 * display order. Add a new entry here + a matching icon map entry in the
 * section component + matching translation block under
 * `docs.clients.<id>.*` and it will render automatically.
 */
export const CLIENT_INTEGRATIONS: ClientIntegration[] = [
  {
    id: 'codebuddy',
    nameKey: 'docs.clients.codebuddy.name',
    descriptionKey: 'docs.clients.codebuddy.description',
    baseUrlLabelKey: 'docs.clients.field.baseUrl',
    baseUrlKey: 'docs.clients.codebuddy.baseUrl',
    protocolKey: 'docs.clients.codebuddy.protocol',
    defaultModelKey: 'docs.clients.codebuddy.defaultModel',
    steps: [
      {
        titleKey: 'docs.clients.codebuddy.steps.1.title',
        descriptionKey: 'docs.clients.codebuddy.steps.1.description',
      },
      {
        titleKey: 'docs.clients.codebuddy.steps.2.title',
        descriptionKey: 'docs.clients.codebuddy.steps.2.description',
      },
      {
        titleKey: 'docs.clients.codebuddy.steps.3.title',
        descriptionKey: 'docs.clients.codebuddy.steps.3.description',
      },
      {
        titleKey: 'docs.clients.codebuddy.steps.4.title',
        descriptionKey: 'docs.clients.codebuddy.steps.4.description',
      },
    ],
  },
  {
    id: 'chatbox',
    nameKey: 'docs.clients.chatbox.name',
    descriptionKey: 'docs.clients.chatbox.description',
    baseUrlLabelKey: 'docs.clients.field.baseUrl',
    baseUrlKey: 'docs.clients.chatbox.baseUrl',
    protocolKey: 'docs.clients.chatbox.protocol',
    defaultModelKey: 'docs.clients.chatbox.defaultModel',
    steps: [
      {
        titleKey: 'docs.clients.chatbox.steps.1.title',
        descriptionKey: 'docs.clients.chatbox.steps.1.description',
      },
      {
        titleKey: 'docs.clients.chatbox.steps.2.title',
        descriptionKey: 'docs.clients.chatbox.steps.2.description',
      },
      {
        titleKey: 'docs.clients.chatbox.steps.3.title',
        descriptionKey: 'docs.clients.chatbox.steps.3.description',
      },
      {
        titleKey: 'docs.clients.chatbox.steps.4.title',
        descriptionKey: 'docs.clients.chatbox.steps.4.description',
      },
    ],
  },
  {
    id: 'cherryStudio',
    nameKey: 'docs.clients.cherryStudio.name',
    descriptionKey: 'docs.clients.cherryStudio.description',
    baseUrlLabelKey: 'docs.clients.field.baseUrl',
    baseUrlKey: 'docs.clients.cherryStudio.baseUrl',
    protocolKey: 'docs.clients.cherryStudio.protocol',
    defaultModelKey: 'docs.clients.cherryStudio.defaultModel',
    steps: [
      {
        titleKey: 'docs.clients.cherryStudio.steps.1.title',
        descriptionKey: 'docs.clients.cherryStudio.steps.1.description',
      },
      {
        titleKey: 'docs.clients.cherryStudio.steps.2.title',
        descriptionKey: 'docs.clients.cherryStudio.steps.2.description',
      },
      {
        titleKey: 'docs.clients.cherryStudio.steps.3.title',
        descriptionKey: 'docs.clients.cherryStudio.steps.3.description',
      },
      {
        titleKey: 'docs.clients.cherryStudio.steps.4.title',
        descriptionKey: 'docs.clients.cherryStudio.steps.4.description',
      },
    ],
  },
]

/**
 * Six headline capabilities that anchor the docs page. Each item's titleKey
 * / descriptionKey must exist in the translation files. Icons are looked
 * up in the section component via `FEATURE_GRID_ICON_MAP`.
 */
export const FEATURE_GRID_ITEMS: FeatureGridItem[] = [
  {
    id: 'multiModel',
    titleKey: 'docs.features.items.multiModel.title',
    descriptionKey: 'docs.features.items.multiModel.description',
  },
  {
    id: 'openAICompat',
    titleKey: 'docs.features.items.openAICompat.title',
    descriptionKey: 'docs.features.items.openAICompat.description',
  },
  {
    id: 'billing',
    titleKey: 'docs.features.items.billing.title',
    descriptionKey: 'docs.features.items.billing.description',
  },
  {
    id: 'routing',
    titleKey: 'docs.features.items.routing.title',
    descriptionKey: 'docs.features.items.routing.description',
  },
  {
    id: 'formatConvert',
    titleKey: 'docs.features.items.formatConvert.title',
    descriptionKey: 'docs.features.items.formatConvert.description',
  },
  {
    id: 'permission',
    titleKey: 'docs.features.items.permission.title',
    descriptionKey: 'docs.features.items.permission.description',
  },
]