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
import { PublicLayout } from '@/components/layout'

import {
  ClientIntegrationSection,
  CodeSamples,
  DocsHero,
  FeatureGrid,
  QuickStart,
} from './components'

/**
 * Public-facing developer docs page. Section ordering mirrors the standard
 * OpenAI-style onboarding flow that deployment guides tend to expect:
 * Hero → Quick Start → Code Samples → Client integrations → Feature grid.
 * Uses `showMainContainer={false}` so the hero's gradient banner can span
 * the full viewport width — the same trick the home page applies.
 */
export function DocsPage() {
  return (
    <PublicLayout showMainContainer={false}>
      <DocsHero />
      <QuickStart />
      <CodeSamples />
      <ClientIntegrationSection />
      <FeatureGrid />
    </PublicLayout>
  )
}