/**
 * RewardCard Component
 *
 * Displays a tournament reward with placement medal, title, description.
 * Shows locked state for non-winners, unlocked state with hidden value for winners.
 */

import { useState } from 'react'
import { TrophyIcon, LockClosedIcon, ClipboardDocumentIcon, CheckIcon } from '@heroicons/react/24/outline'
import { UI_LABELS, COPY_FEEDBACK_DURATION_MS } from '../constants'
import type { TournamentReward, RewardPlacement } from '../types'

interface RewardCardProps {
  reward: TournamentReward
  /** Whether current user is the winner who can see hidden value */
  isWinner?: boolean
  /** Whether to show the hidden value (unlocked) */
  isUnlocked?: boolean
}

const PLACEMENT_CONFIG: Record<RewardPlacement, {
  label: string
  medalColor: string
  bgGradient: string
}> = {
  1: {
    label: '1. Miejsce',
    medalColor: 'bg-amber-500',
    bgGradient: 'from-amber-500/10 to-amber-600/5 dark:from-amber-500/20 dark:to-amber-600/10'
  },
  2: {
    label: '2. Miejsce',
    medalColor: 'bg-zinc-400',
    bgGradient: 'from-zinc-400/10 to-zinc-500/5 dark:from-zinc-400/20 dark:to-zinc-500/10'
  },
  3: {
    label: '3. Miejsce',
    medalColor: 'bg-amber-700',
    bgGradient: 'from-amber-700/10 to-amber-800/5 dark:from-amber-700/20 dark:to-amber-800/10'
  }
}

export function RewardCard({ reward, isWinner = false, isUnlocked = false }: RewardCardProps) {
  const [isCopied, setIsCopied] = useState(false)
  const labels = UI_LABELS.rewards
  const config = PLACEMENT_CONFIG[reward.placement]

  const handleCopyCode = async () => {
    if (!reward.hiddenValue) return

    try {
      await navigator.clipboard.writeText(reward.hiddenValue)
      setIsCopied(true)
      setTimeout(() => setIsCopied(false), COPY_FEEDBACK_DURATION_MS)
    } catch (err) {
      console.error('Failed to copy code:', err)
    }
  }

  const showHiddenValue = isWinner && isUnlocked

  return (
    <div className={`
      relative overflow-hidden rounded-xl border
      ${showHiddenValue
        ? 'border-emerald-500/50 dark:border-emerald-400/30'
        : 'border-zinc-200 dark:border-zinc-700'
      }
    `}>
      {/* Background gradient based on placement */}
      <div className={`absolute inset-0 bg-gradient-to-br ${config.bgGradient}`} />

      <div className="relative p-4">
        {/* Header with medal and placement */}
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <div className={`w-8 h-8 rounded-full flex items-center justify-center ${config.medalColor}`}>
              <TrophyIcon className="w-5 h-5 text-white" />
            </div>
            <span className="text-sm font-semibold text-zinc-700 dark:text-zinc-300">
              {config.label}
            </span>
          </div>

          {/* Status indicator */}
          {showHiddenValue ? (
            <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-600 dark:text-emerald-400">
              <CheckIcon className="w-4 h-4" />
              {reward.isClaimed ? labels.cardClaimed : labels.cardPending}
            </span>
          ) : (
            <span className="inline-flex items-center gap-1 text-xs font-medium text-zinc-500 dark:text-zinc-400">
              <LockClosedIcon className="w-4 h-4" />
              {labels.cardLocked}
            </span>
          )}
        </div>

        {/* Reward title */}
        <h3 className="text-base font-semibold text-zinc-900 dark:text-white mb-1">
          {reward.title}
        </h3>

        {/* Reward description */}
        <p className="text-sm text-zinc-600 dark:text-zinc-400 mb-3">
          {reward.description}
        </p>

        {/* Hidden value section - only shown to winners */}
        {showHiddenValue ? (
          <div className="mt-3 p-3 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800">
            <div className="flex items-center justify-between gap-2">
              <div className="flex-1 min-w-0">
                <p className="text-xs text-emerald-600 dark:text-emerald-400 mb-1">
                  {labels.cardUnlocked}
                </p>
                <p className="text-sm font-mono text-emerald-800 dark:text-emerald-300 break-all">
                  {reward.hiddenValue}
                </p>
              </div>
              <button
                onClick={handleCopyCode}
                className="flex-shrink-0 p-2 rounded-lg hover:bg-emerald-100 dark:hover:bg-emerald-800/30 transition-colors"
                title={labels.copyCode}
              >
                {isCopied ? (
                  <CheckIcon className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
                ) : (
                  <ClipboardDocumentIcon className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
                )}
              </button>
            </div>
            {isCopied && (
              <p className="text-xs text-emerald-600 dark:text-emerald-400 mt-1">
                {labels.codeCopied}
              </p>
            )}
          </div>
        ) : (
          <div className="mt-3 p-3 rounded-lg bg-zinc-100 dark:bg-zinc-800/50 border border-zinc-200 dark:border-zinc-700">
            <div className="flex items-center gap-2 text-zinc-500 dark:text-zinc-400">
              <LockClosedIcon className="w-4 h-4" />
              <p className="text-sm">
                {labels.cardLocked}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

/**
 * RewardCardsGrid Component
 *
 * Displays a grid of reward cards for 1st, 2nd, and 3rd place.
 */
interface RewardCardsGridProps {
  rewards: TournamentReward[]
  /** Current user's placement (1, 2, or 3) if they won, undefined otherwise */
  userPlacement?: RewardPlacement
}

export function RewardCardsGrid({ rewards, userPlacement }: RewardCardsGridProps) {
  // Sort rewards by placement
  const sortedRewards = [...rewards].sort((a, b) => a.placement - b.placement)

  if (sortedRewards.length === 0) {
    return null
  }

  return (
    <div className="grid gap-4 md:grid-cols-3">
      {sortedRewards.map(reward => (
        <RewardCard
          key={reward.id || `reward-${reward.placement}`}
          reward={reward}
          isWinner={userPlacement === reward.placement}
          isUnlocked={userPlacement === reward.placement}
        />
      ))}
    </div>
  )
}
