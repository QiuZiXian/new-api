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
import { ChevronDown, Settings2, X } from 'lucide-react'
import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from 'react'
import { useTranslation } from 'react-i18next'

import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface TagInputProps {
  value: string[]
  onChange: (tags: string[]) => void
  placeholder?: string
  className?: string
  disabled?: boolean
  /** Managed dictionary entries offered in the dropdown. Free typing stays enabled. */
  suggestions?: string[]
  /** Opens the tag dictionary manager, when provided. */
  onManage?: () => void
}

export function TagInput({
  value = [],
  onChange,
  placeholder,
  className,
  disabled = false,
  suggestions = [],
  onManage,
}: TagInputProps) {
  const { t } = useTranslation()
  const placeholderText = placeholder ?? t('Add tags...')
  const [inputValue, setInputValue] = useState('')
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)

  const showDropdown = suggestions.length > 0 || Boolean(onManage)

  useEffect(() => {
    if (!open) return
    const handlePointerDown = (event: MouseEvent) => {
      if (!containerRef.current?.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handlePointerDown)
    return () => document.removeEventListener('mousedown', handlePointerDown)
  }, [open])

  const options = useMemo(() => {
    const selected = new Set(value.map((tag) => tag.toLowerCase()))
    const keyword = query.trim().toLowerCase()
    return suggestions.filter(
      (tag) =>
        !selected.has(tag.toLowerCase()) &&
        (!keyword || tag.toLowerCase().includes(keyword))
    )
  }, [suggestions, value, query])

  const addTag = (tag: string) => {
    const trimmed = tag.trim()
    if (trimmed && !value.includes(trimmed)) {
      onChange([...value, trimmed])
    }
    setInputValue('')
  }

  const removeTag = (tagToRemove: string) => {
    onChange(value.filter((tag) => tag !== tagToRemove))
  }

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      addTag(inputValue)
    } else if (e.key === 'Backspace' && !inputValue && value.length > 0) {
      const lastTag = value.at(-1)
      if (lastTag) removeTag(lastTag)
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  const handleBlur = () => {
    if (inputValue.trim()) {
      addTag(inputValue)
    }
  }

  const toggleDropdown = () => {
    setOpen((previous) => {
      const next = !previous
      if (next) {
        setQuery('')
        window.setTimeout(() => searchRef.current?.focus(), 0)
      }
      return next
    })
  }

  return (
    <div ref={containerRef} className='relative'>
      <div
        className={cn(
          'border-input focus-within:border-ring focus-within:ring-ring/50 flex min-h-9 w-full flex-wrap items-center gap-2 rounded-md border bg-transparent px-3 py-2 text-base shadow-xs transition-[color,box-shadow] outline-none focus-within:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
          className
        )}
        onClick={() => inputRef.current?.focus()}
      >
        {value.map((tag) => (
          <Badge key={tag} variant='secondary' className='gap-1 pr-1'>
            {tag}
            {!disabled && (
              <Button
                type='button'
                variant='ghost'
                size='icon-sm'
                aria-label='Remove tag'
                onClick={(e) => {
                  e.stopPropagation()
                  removeTag(tag)
                }}
                className='hover:bg-secondary-foreground/20 size-auto rounded-sm p-0'
              >
                <X className='h-3 w-3' aria-hidden='true' />
              </Button>
            )}
          </Badge>
        ))}
        <input
          ref={inputRef}
          type='text'
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={handleBlur}
          placeholder={value.length === 0 ? placeholderText : ''}
          disabled={disabled}
          className='placeholder:text-muted-foreground min-w-[120px] flex-1 border-0 bg-transparent shadow-none outline-none focus-visible:ring-0'
        />
        {showDropdown && !disabled && (
          <button
            type='button'
            aria-label={t('Select a tag')}
            aria-expanded={open}
            onClick={(e) => {
              e.stopPropagation()
              toggleDropdown()
            }}
            className='text-muted-foreground hover:text-foreground shrink-0 rounded p-1 transition-colors'
          >
            <ChevronDown
              className={cn('size-4 transition-transform', open && 'rotate-180')}
            />
          </button>
        )}
      </div>

      {open && showDropdown && (
        <div className='border-border bg-background absolute inset-x-0 top-full z-50 mt-1 rounded-md border p-1.5 shadow-lg'>
          <input
            ref={searchRef}
            type='text'
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Escape') setOpen(false)
              if (e.key === 'Enter' && options.length > 0) {
                e.preventDefault()
                addTag(options[0])
                setQuery('')
              }
            }}
            placeholder={t('Search tags...')}
            className='border-input bg-background placeholder:text-muted-foreground mb-1.5 h-8 w-full rounded-md border px-2.5 text-sm outline-none focus:border-blue-500/60'
          />
          <div className='max-h-48 overflow-y-auto'>
            {options.length > 0 ? (
              options.map((tag) => (
                <button
                  key={tag}
                  type='button'
                  onClick={() => {
                    addTag(tag)
                    setQuery('')
                    searchRef.current?.focus()
                  }}
                  className='hover:bg-muted text-foreground flex w-full items-center rounded px-2 py-1.5 text-left text-sm transition-colors'
                >
                  {tag}
                </button>
              ))
            ) : (
              <p className='text-muted-foreground px-2 py-3 text-center text-xs'>
                {t('No matching tags')}
              </p>
            )}
          </div>
          {onManage && (
            <button
              type='button'
              onClick={() => {
                setOpen(false)
                onManage()
              }}
              className='text-muted-foreground hover:bg-muted hover:text-foreground mt-1 flex w-full items-center gap-1.5 rounded border-t px-2 pt-2 pb-1 text-left text-xs transition-colors'
            >
              <Settings2 className='size-3.5' />
              {t('Manage Tag Dictionary')}
            </button>
          )}
        </div>
      )}
    </div>
  )
}
