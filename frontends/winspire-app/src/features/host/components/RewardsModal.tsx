/**
 * RewardsModal Component
 *
 * Modal for configuring tournament rewards for 1st, 2nd, and 3rd place.
 * Allows streamer to set public title, description, and hidden value (e.g., activation code).
 */

import { useState } from 'react'
import { TrophyIcon } from '@heroicons/react/24/outline'
import { Dialog, DialogActions, DialogBody, DialogDescription, DialogTitle } from '../../../shared/components/ui/dialog'
import { Button } from '../../../shared/components/ui/button'
import { Input } from '../../../shared/components/ui/input'
import { Textarea } from '../../../shared/components/ui/textarea'
import { Field, Label, Description } from '../../../shared/components/ui/fieldset'
import { UI_LABELS } from '../constants'
import type { Tournament, RewardFormData, TournamentReward } from '../types'

interface RewardsModalProps {
  isOpen: boolean
  onClose: () => void
  tournament: Tournament
  existingRewards?: TournamentReward[]
  onSave: (rewards: RewardFormData) => Promise<void>
}

interface PlacementSectionProps {
  placement: 'first' | 'second' | 'third'
  label: string
  medalColor: string
  data: {
    title: string
    description: string
    hiddenValue: string
  }
  onChange: (field: 'title' | 'description' | 'hiddenValue', value: string) => void
  labels: typeof UI_LABELS.rewards
}

function PlacementSection({ placement, label, medalColor, data, onChange, labels }: PlacementSectionProps) {
  return (
    <div className="rounded-lg border border-zinc-200 dark:border-zinc-700 p-4">
      <div className="flex items-center gap-2 mb-4">
        <div className={`w-8 h-8 rounded-full flex items-center justify-center ${medalColor}`}>
          <TrophyIcon className="w-5 h-5 text-white" />
        </div>
        <h3 className="text-base font-semibold text-zinc-900 dark:text-white">{label}</h3>
      </div>

      <div className="space-y-4">
        <Field>
          <Label>{labels.titleLabel}</Label>
          <Input
            type="text"
            placeholder={labels.titlePlaceholder}
            value={data.title}
            onChange={(e) => onChange('title', e.target.value)}
          />
        </Field>

        <Field>
          <Label>{labels.descriptionLabel}</Label>
          <Textarea
            placeholder={labels.descriptionPlaceholder}
            value={data.description}
            onChange={(e) => onChange('description', e.target.value)}
            rows={2}
            resizable={false}
          />
        </Field>

        <Field>
          <Label>{labels.hiddenValueLabel}</Label>
          <Input
            type="text"
            placeholder={labels.hiddenValuePlaceholder}
            value={data.hiddenValue}
            onChange={(e) => onChange('hiddenValue', e.target.value)}
          />
          <Description>{labels.hiddenValueHint}</Description>
        </Field>
      </div>
    </div>
  )
}

const initialRewardData = {
  title: '',
  description: '',
  hiddenValue: ''
}

export function RewardsModal({
  isOpen,
  onClose,
  tournament: _tournament,
  existingRewards = [],
  onSave
}: RewardsModalProps) {
  const labels = UI_LABELS.rewards

  // Initialize form data from existing rewards or empty
  const getInitialData = (placement: 1 | 2 | 3) => {
    const existing = existingRewards.find(r => r.placement === placement)
    if (existing) {
      return {
        title: existing.title,
        description: existing.description,
        hiddenValue: existing.hiddenValue
      }
    }
    return { ...initialRewardData }
  }

  const [formData, setFormData] = useState<RewardFormData>({
    first: getInitialData(1),
    second: getInitialData(2),
    third: getInitialData(3)
  })

  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleChange = (
    placement: 'first' | 'second' | 'third',
    field: 'title' | 'description' | 'hiddenValue',
    value: string
  ) => {
    setFormData(prev => ({
      ...prev,
      [placement]: {
        ...prev[placement],
        [field]: value
      }
    }))
  }

  const handleSave = async () => {
    setIsLoading(true)
    setError(null)
    try {
      await onSave(formData)
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : labels.saveError)
    } finally {
      setIsLoading(false)
    }
  }

  const handleClose = () => {
    if (!isLoading) {
      onClose()
    }
  }

  return (
    <Dialog open={isOpen} onClose={handleClose} size="2xl">
      <DialogTitle>{labels.modalTitle}</DialogTitle>
      <DialogDescription>{labels.modalDescription}</DialogDescription>

      <DialogBody>
        <div className="space-y-6">
          {error && (
            <div className="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 px-3 py-2 rounded-lg">
              {error}
            </div>
          )}

          <PlacementSection
            placement="first"
            label={labels.firstPlace}
            medalColor="bg-amber-500"
            data={formData.first}
            onChange={(field, value) => handleChange('first', field, value)}
            labels={labels}
          />

          <PlacementSection
            placement="second"
            label={labels.secondPlace}
            medalColor="bg-zinc-400"
            data={formData.second}
            onChange={(field, value) => handleChange('second', field, value)}
            labels={labels}
          />

          <PlacementSection
            placement="third"
            label={labels.thirdPlace}
            medalColor="bg-amber-700"
            data={formData.third}
            onChange={(field, value) => handleChange('third', field, value)}
            labels={labels}
          />
        </div>
      </DialogBody>

      <DialogActions>
        <Button plain onClick={handleClose} disabled={isLoading}>
          {UI_LABELS.form.cancelButton}
        </Button>
        <Button color="cyan" onClick={handleSave} disabled={isLoading}>
          {isLoading ? UI_LABELS.loading : labels.saveButton}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
