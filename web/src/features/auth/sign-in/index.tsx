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
import { Link, useSearch } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useStatus } from '@/hooks/use-status'

import { AuthLayout } from '../auth-layout'
import { TermsFooter } from '../components/terms-footer'
import { UserAuthForm } from './components/user-auth-form'

export function SignIn() {
  const { t } = useTranslation()
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const { status } = useStatus()

  return (
    <AuthLayout>
      <div className='w-full space-y-8'>
        <div className='space-y-2'>
          {/* Small brand tagline above the sign-in heading */}
          <p className='text-muted-foreground/70 mb-4 flex items-center gap-2 text-xs font-medium tracking-[0.15em] uppercase sm:text-left'>
            <span className='relative flex size-1.5'>
              <span className='absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-400 opacity-75' />
              <span className='relative inline-flex size-1.5 rounded-full bg-blue-500 dark:bg-blue-400' />
            </span>
            {t('AI Application Infrastructure Foundation')}
          </p>
          <h2 className='text-center text-2xl font-semibold tracking-tight sm:text-left'>
            {t('Sign in')}
          </h2>
          {!status?.self_use_mode_enabled &&
            status?.register_enabled !== false && (
              <div className='text-muted-foreground flex flex-wrap items-center gap-2.5 text-left text-sm sm:text-base'>
                <span>{t("Don't have an account?")}</span>
                <Link
                  to='/sign-up'
                  className='inline-flex h-10 items-center gap-1.5 rounded-lg bg-gradient-to-r from-blue-600 to-blue-500 px-5 text-base font-semibold text-white shadow-[0_8px_20px_-6px_rgba(37,99,235,0.6)] transition-all hover:from-blue-700 hover:to-blue-600'
                >
                  {t('Sign up')}
                  <ArrowRight className='size-4' />
                </Link>
              </div>
            )}
        </div>

        <UserAuthForm redirectTo={redirect} />

        <TermsFooter
          variant='sign-in'
          status={status}
          className='text-center'
        />
      </div>
    </AuthLayout>
  )
}
