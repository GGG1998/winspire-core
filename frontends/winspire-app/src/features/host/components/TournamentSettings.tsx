/**
 * TournamentSettings Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Dropdown menu with tournament management actions: Edit, Start, Cancel
 * Visible only to tournament owner
 */

import { useState } from 'react'
import { Cog6ToothIcon, PencilIcon, PlayIcon, XMarkIcon, RocketLaunchIcon, ClipboardDocumentCheckIcon } from '@heroicons/react/20/solid'
import {
  Dropdown,
  DropdownButton,
  DropdownMenu,
  DropdownItem,
  DropdownLabel,
  DropdownDivider
} from '../../../shared/components/ui/dropdown'
import { Dialog, DialogActions, DialogBody, DialogDescription, DialogTitle } from '../../../shared/components/ui/dialog'
import { Button } from '../../../shared/components/ui/button'
import { tournamentApi } from '../api/tournamentApi'
import { UI_LABELS } from '../constants'
import type { Tournament, TournamentStatus } from '../types'

interface TournamentSettingsProps {
  tournament: Tournament
  status: TournamentStatus
  onEdit?: () => void
  onPublished?: () => void
  onOpenedRegistration?: () => void
  onStarted?: () => void
  onCancelled?: () => void
}

type ConfirmAction = 'publish' | 'openRegistration' | 'start' | 'cancel' | null

export function TournamentSettings({
  tournament,
  status,
  onEdit,
  onPublished,
  onOpenedRegistration,
  onStarted,
  onCancelled
}: TournamentSettingsProps) {
  const [confirmAction, setConfirmAction] = useState<ConfirmAction>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const labels = UI_LABELS.settings

  // Can publish if tournament is draft
  const canPublish = status === 'draft'
  // Can open registration if tournament is scheduled
  const canOpenRegistration = status === 'scheduled'
  // Can start if tournament is registration_open or registration_closed (not if already starting)
  const canStart = (status === 'registration_open' || status === 'registration_closed') && status !== 'starting'
  // Can cancel if tournament is not completed
  const canCancel = status !== 'completed'
  // Can edit if tournament is not completed or cancelled
  const canEdit = status !== 'completed' && status !== 'cancelled'

  const handlePublish = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await tournamentApi.publishTournament(tournament.id, tournament.startTime)
      setConfirmAction(null)
      onPublished?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : labels.publishError)
    } finally {
      setIsLoading(false)
    }
  }

  const handleOpenRegistration = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await tournamentApi.openRegistration(tournament.id)
      setConfirmAction(null)
      onOpenedRegistration?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : labels.openRegistrationError)
    } finally {
      setIsLoading(false)
    }
  }

  const handleStart = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await tournamentApi.startTournament(tournament.id)
      setConfirmAction(null)
      // Note: Status will be 'starting' - frontend should show grace period UI
      // Tournament will transition to 'started' after grace period via WebSocket event
      onStarted?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : labels.startError)
    } finally {
      setIsLoading(false)
    }
  }

  const handleCancel = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await tournamentApi.cancelTournament(tournament.id)
      setConfirmAction(null)
      onCancelled?.()
    } catch (err) {
      setError(err instanceof Error ? err.message : labels.cancelError)
    } finally {
      setIsLoading(false)
    }
  }

  const handleConfirmAction = () => {
    if (confirmAction === 'publish') {
      handlePublish()
    } else if (confirmAction === 'openRegistration') {
      handleOpenRegistration()
    } else if (confirmAction === 'start') {
      handleStart()
    } else if (confirmAction === 'cancel') {
      handleCancel()
    }
  }

  // Check if there are any available actions
  const hasActions = canPublish || canOpenRegistration || canEdit || canStart || canCancel

  if (!hasActions) {
    return null
  }

  return (
    <>
      <Dropdown>
        <DropdownButton
          className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-white/10 hover:bg-white/20 text-white text-sm font-medium transition-colors"
        >
          <Cog6ToothIcon className="size-4" />
          <span>{labels.title}</span>
        </DropdownButton>
        <DropdownMenu anchor="bottom end">
          {canPublish && (
            <DropdownItem onClick={() => setConfirmAction('publish')}>
              <RocketLaunchIcon data-slot="icon" className="text-cyan-500" />
              <DropdownLabel>{labels.publish}</DropdownLabel>
            </DropdownItem>
          )}
          {canOpenRegistration && (
            <DropdownItem onClick={() => setConfirmAction('openRegistration')}>
              <ClipboardDocumentCheckIcon data-slot="icon" className="text-emerald-500" />
              <DropdownLabel>{labels.openRegistration}</DropdownLabel>
            </DropdownItem>
          )}
          {canEdit && (
            <DropdownItem onClick={onEdit}>
              <PencilIcon data-slot="icon" />
              <DropdownLabel>{labels.edit}</DropdownLabel>
            </DropdownItem>
          )}
          {canStart && (
            <DropdownItem onClick={() => setConfirmAction('start')}>
              <PlayIcon data-slot="icon" className="text-emerald-500" />
              <DropdownLabel>{labels.start}</DropdownLabel>
            </DropdownItem>
          )}
          {(canPublish || canOpenRegistration || canEdit || canStart) && canCancel && <DropdownDivider />}
          {canCancel && (
            <DropdownItem onClick={() => setConfirmAction('cancel')}>
              <XMarkIcon data-slot="icon" className="text-red-500" />
              <DropdownLabel className="text-red-600 dark:text-red-400">
                {labels.cancel}
              </DropdownLabel>
            </DropdownItem>
          )}
        </DropdownMenu>
      </Dropdown>

      {/* Confirmation Dialog */}
      <Dialog open={confirmAction !== null} onClose={() => setConfirmAction(null)}>
        <DialogTitle>
          {confirmAction === 'publish' ? labels.publish 
            : confirmAction === 'openRegistration' ? labels.openRegistration
            : confirmAction === 'start' ? labels.start 
            : labels.cancel}
        </DialogTitle>
        <DialogDescription>
          {confirmAction === 'publish' ? labels.confirmPublish 
            : confirmAction === 'openRegistration' ? labels.confirmOpenRegistration
            : confirmAction === 'start' ? labels.confirmStart 
            : labels.confirmCancel}
        </DialogDescription>
        <DialogBody>
          {error && (
            <div className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-3 py-2 rounded-lg">
              {error}
            </div>
          )}
        </DialogBody>
        <DialogActions>
          <Button
            plain
            onClick={() => setConfirmAction(null)}
            disabled={isLoading}
          >
            {UI_LABELS.form.cancelButton}
          </Button>
          <Button
            color={confirmAction === 'publish' ? 'cyan' 
              : confirmAction === 'openRegistration' ? 'emerald'
              : confirmAction === 'start' ? 'emerald' 
              : 'red'}
            onClick={handleConfirmAction}
            disabled={isLoading}
          >
            {isLoading ? UI_LABELS.loading 
              : confirmAction === 'publish' ? labels.publish 
              : confirmAction === 'openRegistration' ? labels.openRegistration
              : confirmAction === 'start' ? labels.start 
              : labels.cancel}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

