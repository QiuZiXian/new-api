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
export type QuickStartStep = {
  number: number
  titleKey: string
  descriptionKey: string
}

export type CodeSample = {
  id: string
  labelKey: string
  language: 'python' | 'javascript' | 'bash'
  /**
   * Identifier the renderer uses to pick a snippet template from
   * `SNIPPET_TEMPLATES` inside `code-samples.tsx`. Keeping the template
   * out of i18n avoids large, escape-heavy translation strings.
   */
  snippetId: 'python' | 'javascript' | 'bash'
}

export type ClientIntegrationStep = {
  titleKey: string
  descriptionKey: string
}

export type ClientIntegrationIconId = 'codebuddy' | 'chatbox' | 'cherryStudio'

export type ClientIntegration = {
  id: ClientIntegrationIconId
  nameKey: string
  descriptionKey: string
  baseUrlLabelKey: string
  /**
   * Translation key for the Base URL the user is told to paste into the
   * third-party client. Keep it under `docs.clients.<id>.baseUrl` so the
   * deploy-time host can be swapped without touching this file.
   */
  baseUrlKey: string
  protocolKey: string
  defaultModelKey: string
  /** Ordered onboarding steps shown inside the card body. */
  steps: ClientIntegrationStep[]
}

export type FeatureGridIconId =
  | 'multiModel'
  | 'openAICompat'
  | 'billing'
  | 'routing'
  | 'formatConvert'
  | 'permission'

export type FeatureGridItem = {
  id: FeatureGridIconId
  titleKey: string
  descriptionKey: string
}