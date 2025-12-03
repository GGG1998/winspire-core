/**
 * CreateTournamentModal Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Modal form for creating a new tournament
 * Uses: Zod validation, React Hook Form, shared UI components
 */

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Dialog, DialogTitle, DialogBody, DialogActions } from '../../../shared/components/ui/dialog'
import { Button } from '../../../shared/components/ui/button'
import { Input } from '../../../shared/components/ui/input'
import { Select } from '../../../shared/components/ui/select'
import { Field, FieldGroup, Label, ErrorMessage } from '../../../shared/components/ui/fieldset'
import { Text } from '../../../shared/components/ui/text'
import { useTournaments } from '../hooks/useTournaments'
import { formToApiInput } from '../api/tournamentApi'
import { createTournamentSchema, type CreateTournamentFormData } from '../schemas'
import { 
  UI_LABELS, 
  DEFAULT_GAME,
  DEFAULT_TEAM_MODE,
  DEFAULT_MAX_PLAYERS,
  TEAM_MODE_OPTIONS,
  TOURNAMENT_VALIDATION,
  VALIDATION_MESSAGES
} from '../constants'
import type { CreateTournamentModalProps } from '../types'

/**
 * Get default start time (now + 1 hour)
 */
function getDefaultStartTime(): string {
  const now = new Date()
  now.setHours(now.getHours() + 1)
  return now.toISOString().slice(0, 16)
}

/**
 * Get minimum datetime value (now + 1 minute)
 */
function getMinDateTime(): string {
  const now = new Date()
  now.setMinutes(now.getMinutes() + 1)
  return now.toISOString().slice(0, 16)
}

/**
 * Modal component for tournament creation
 */
export function CreateTournamentModal({
  isOpen,
  onClose,
  onSuccess
}: CreateTournamentModalProps) {
  const { createTournament } = useTournaments()
  
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors, isSubmitting }
  } = useForm<CreateTournamentFormData>({
    resolver: zodResolver(createTournamentSchema),
    defaultValues: {
      name: '',
      startTime: getDefaultStartTime(),
      game: DEFAULT_GAME,
      teamMode: DEFAULT_TEAM_MODE,
      maxPlayers: DEFAULT_MAX_PLAYERS
    }
  })

  // Reset form when modal opens
  useEffect(() => {
    if (isOpen) {
      reset({
        name: '',
        startTime: getDefaultStartTime(),
        game: DEFAULT_GAME,
        teamMode: DEFAULT_TEAM_MODE,
        maxPlayers: DEFAULT_MAX_PLAYERS
      })
    }
  }, [isOpen, reset])

  /**
   * Handle form submission
   */
  const onSubmit = async (data: CreateTournamentFormData) => {
    try {
      const input = formToApiInput({
        name: data.name,
        startTime: data.startTime,
        game: data.game,
        teamMode: data.teamMode,
        maxPlayers: data.maxPlayers
      })
      const tournament = await createTournament(input)
      onSuccess(tournament)
    } catch (err) {
      console.error('Failed to create tournament:', err)
      setError('root', { 
        type: 'server',
        message: VALIDATION_MESSAGES.general.server 
      })
    }
  }

  return (
    <Dialog open={isOpen} onClose={() => {}}>
      <DialogTitle>{UI_LABELS.page.createButton}</DialogTitle>
      
      <DialogBody>
        <form id="create-tournament-form" onSubmit={handleSubmit(onSubmit)}>
          <FieldGroup>
            {/* General error */}
            {errors.root && (
              <div className="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 dark:bg-red-900/20 dark:ring-red-800">
                <Text className="text-red-800 dark:text-red-200">
                  {errors.root.message}
                </Text>
              </div>
            )}

            {/* Name field */}
            <Field>
              <Label htmlFor="tournament-name">{UI_LABELS.form.nameLabel}</Label>
              <Input
                id="tournament-name"
                type="text"
                placeholder={UI_LABELS.form.namePlaceholder}
                maxLength={TOURNAMENT_VALIDATION.name.maxLength}
                autoFocus
                {...register('name')}
                data-invalid={errors.name ? '' : undefined}
              />
              {errors.name && (
                <ErrorMessage>{errors.name.message}</ErrorMessage>
              )}
            </Field>

            {/* Start time field */}
            <Field>
              <Label htmlFor="tournament-start-time">{UI_LABELS.form.startTimeLabel}</Label>
              <Input
                id="tournament-start-time"
                type="datetime-local"
                min={getMinDateTime()}
                {...register('startTime')}
                data-invalid={errors.startTime ? '' : undefined}
              />
              {errors.startTime && (
                <ErrorMessage>{errors.startTime.message}</ErrorMessage>
              )}
            </Field>

            {/* Game field (disabled) */}
            <Field>
              <Label htmlFor="tournament-game">{UI_LABELS.form.gameLabel}</Label>
              <Input
                id="tournament-game"
                type="text"
                disabled
                {...register('game')}
              />
            </Field>

            {/* Team mode field (disabled) */}
            <Field>
              <Label htmlFor="tournament-team-mode">{UI_LABELS.form.teamModeLabel}</Label>
              <Select
                id="tournament-team-mode"
                disabled
                {...register('teamMode')}
              >
                {TEAM_MODE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Select>
            </Field>

            {/* Max players field (disabled) */}
            <Field>
              <Label htmlFor="tournament-max-players">{UI_LABELS.form.maxPlayersLabel}</Label>
              <Input
                id="tournament-max-players"
                type="number"
                disabled
                min={1}
                {...register('maxPlayers', { valueAsNumber: true })}
              />
            </Field>
          </FieldGroup>
        </form>
      </DialogBody>

      <DialogActions>
        <Button outline onClick={onClose}>
          {UI_LABELS.form.cancelButton}
        </Button>
        <Button 
          type="submit" 
          form="create-tournament-form"
          color="blue"
          disabled={isSubmitting}
        >
          {isSubmitting ? (
            <span className="flex items-center gap-2">
              <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              {UI_LABELS.loading}
            </span>
          ) : (
            UI_LABELS.form.createButton
          )}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
