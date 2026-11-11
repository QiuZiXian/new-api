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
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ArrowDown,
  ArrowUp,
  Check,
  Download,
  Loader2,
  Pencil,
  Plus,
  RefreshCcw,
  Trash2,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { ConfirmDialog } from '@/components/confirm-dialog'
import { Dialog } from '@/components/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

import {
  createModelTag,
  deleteModelTag,
  getModelTags,
  importModelTags,
  updateModelTag,
} from '../../api'
import { modelTagsQueryKeys } from '../../lib'
import type { ModelTag } from '../../types'

interface ModelTagDictionaryDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelTagDictionaryDialog(
  props: ModelTagDictionaryDialogProps
) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [newName, setNewName] = useState('')
  const [editingId, setEditingId] = useState<number | null>(null)
  const [editingName, setEditingName] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<ModelTag | null>(null)
  const [busy, setBusy] = useState(false)

  const { data, isLoading, refetch } = useQuery({
    queryKey: modelTagsQueryKeys.lists(),
    queryFn: getModelTags,
    enabled: props.open,
  })
  const tags = data?.data ?? []

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setNewName('')
      setEditingId(null)
      setDeleteTarget(null)
    }
    props.onOpenChange(next)
  }

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: modelTagsQueryKeys.lists() })

  const run = async (action: () => Promise<{ success: boolean; message?: string }>) => {
    setBusy(true)
    try {
      const response = await action()
      if (!response.success) {
        toast.error(response.message || t('Operation failed'))
        return
      }
      await refresh()
      return response
    } catch (error) {
      toast.error((error as Error)?.message || t('Operation failed'))
    } finally {
      setBusy(false)
    }
    return undefined
  }

  const handleAdd = async () => {
    const name = newName.trim()
    if (!name) return
    const result = await run(() => createModelTag({ name }))
    if (result?.success) {
      setNewName('')
      toast.success(t('Tag added'))
    }
  }

  const handleSaveEdit = async (tag: ModelTag) => {
    const name = editingName.trim()
    if (!name || name === tag.name) {
      setEditingId(null)
      return
    }
    const result = await run(() =>
      updateModelTag({ id: tag.id, name, sort: tag.sort })
    )
    if (result?.success) setEditingId(null)
  }

  const handleDelete = async () => {
    if (!deleteTarget) return
    const result = await run(() => deleteModelTag(deleteTarget.id))
    if (result?.success) setDeleteTarget(null)
  }

  const handleImport = async () => {
    const response = await importModelTags()
    if (!response.success) {
      toast.error(response.message || t('Operation failed'))
      return
    }
    await refresh()
    const added = response.data?.added ?? 0
    toast.success(t('{{count}} tags imported', { count: added }))
  }

  const handleMove = async (index: number, direction: -1 | 1) => {
    const current = tags[index]
    const neighbor = tags[index + direction]
    if (!current || !neighbor) return
    setBusy(true)
    try {
      const [a, b] = await Promise.all([
        updateModelTag({ id: current.id, name: current.name, sort: neighbor.sort }),
        updateModelTag({ id: neighbor.id, name: neighbor.name, sort: current.sort }),
      ])
      if (!a.success || !b.success) {
        toast.error(a.message || b.message || t('Operation failed'))
        return
      }
      await refresh()
    } catch (error) {
      toast.error((error as Error)?.message || t('Operation failed'))
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <Dialog
        open={props.open}
        onOpenChange={handleOpenChange}
        title={t('Model Tag Dictionary')}
        description={t(
          'Manage the selectable tags used on models. The model square filter and card badges still read the tags saved on each model.'
        )}
        contentClassName='sm:max-w-xl'
      >
        <div className='space-y-3'>
          <div className='flex flex-wrap items-center gap-2'>
            <Input
              value={newName}
              onChange={(event) => setNewName(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter') void handleAdd()
              }}
              placeholder={t('Tag Name')}
              className='h-8 min-w-40 flex-1'
              disabled={busy}
            />
            <Button size='sm' onClick={() => void handleAdd()} disabled={busy}>
              <Plus className='h-4 w-4' />
              {t('Add Tag')}
            </Button>
            <Button
              size='sm'
              variant='outline'
              onClick={() => void handleImport()}
              disabled={busy}
            >
              <Download className='h-4 w-4' />
              {t('Import Existing Tags')}
            </Button>
            <Button
              size='icon-sm'
              variant='ghost'
              onClick={() => void refetch()}
              aria-label={t('Refresh')}
            >
              <RefreshCcw
                className={cn('size-3.5', isLoading && 'animate-spin')}
              />
            </Button>
          </div>

          <p className='text-muted-foreground text-xs'>
            {t(
              'Removing a dictionary entry does not remove the tag from models that already use it.'
            )}
          </p>

          <div className='divide-y rounded-md border'>
            {isLoading && (
              <div className='text-muted-foreground flex items-center justify-center gap-2 py-10 text-sm'>
                <Loader2 className='size-4 animate-spin' />
                {t('Loading...')}
              </div>
            )}
            {!isLoading && tags.length === 0 && (
              <p className='text-muted-foreground py-10 text-center text-sm'>
                {t('No tags yet')}
              </p>
            )}
            {!isLoading && tags.length > 0 && (
              tags.map((tag, index) => (
                <div
                  key={tag.id}
                  className='flex items-center gap-2 px-2.5 py-1.5'
                >
                  <span className='text-muted-foreground w-6 shrink-0 text-center font-mono text-xs'>
                    {index + 1}
                  </span>
                  {editingId === tag.id ? (
                    <Input
                      autoFocus
                      value={editingName}
                      onChange={(event) => setEditingName(event.target.value)}
                      onKeyDown={(event) => {
                        if (event.key === 'Enter') void handleSaveEdit(tag)
                        if (event.key === 'Escape') setEditingId(null)
                      }}
                      className='h-7 flex-1'
                    />
                  ) : (
                    <span className='min-w-0 flex-1 truncate text-sm font-medium'>
                      {tag.name}
                    </span>
                  )}

                  {editingId === tag.id ? (
                    <>
                      <Button
                        size='icon-sm'
                        variant='ghost'
                        onClick={() => void handleSaveEdit(tag)}
                        aria-label={t('Save')}
                      >
                        <Check className='size-3.5' />
                      </Button>
                      <Button
                        size='icon-sm'
                        variant='ghost'
                        onClick={() => setEditingId(null)}
                        aria-label={t('Cancel')}
                      >
                        <X className='size-3.5' />
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button
                        size='icon-sm'
                        variant='ghost'
                        disabled={busy || index === 0}
                        onClick={() => void handleMove(index, -1)}
                        aria-label={t('Move up')}
                      >
                        <ArrowUp className='size-3.5' />
                      </Button>
                      <Button
                        size='icon-sm'
                        variant='ghost'
                        disabled={busy || index === tags.length - 1}
                        onClick={() => void handleMove(index, 1)}
                        aria-label={t('Move down')}
                      >
                        <ArrowDown className='size-3.5' />
                      </Button>
                      <Button
                        size='icon-sm'
                        variant='ghost'
                        disabled={busy}
                        onClick={() => {
                          setEditingId(tag.id)
                          setEditingName(tag.name)
                        }}
                        aria-label={t('Edit')}
                      >
                        <Pencil className='size-3.5' />
                      </Button>
                      <Button
                        size='icon-sm'
                        variant='ghost'
                        disabled={busy}
                        className='text-destructive hover:text-destructive'
                        onClick={() => setDeleteTarget(tag)}
                        aria-label={t('Delete')}
                      >
                        <Trash2 className='size-3.5' />
                      </Button>
                    </>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      </Dialog>

      <ConfirmDialog
        open={Boolean(deleteTarget)}
        onOpenChange={(next) => {
          if (!next) setDeleteTarget(null)
        }}
        title={t('Delete tag')}
        desc={t('Are you sure you want to delete "{{name}}"?', {
          name: deleteTarget?.name ?? '',
        })}
        destructive
        confirmText={busy ? t('Deleting...') : t('Delete')}
        isLoading={busy}
        handleConfirm={() => void handleDelete()}
      />
    </>
  )
}
