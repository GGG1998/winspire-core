/**
 * FormatCard Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Card displaying a single format detail (game, ready window, team size, etc.)
 * Supports both light and dark themes.
 */

import clsx from 'clsx'

interface FormatCardProps {
  icon: React.ReactNode
  label: string
  value: string
  warning?: boolean
  className?: string
}

export function FormatCard({ icon, label, value, warning, className }: FormatCardProps) {
  return (
    <div
      className={clsx(
        'relative rounded-xl border p-4 transition-colors',
        'border-zinc-200 bg-white hover:border-zinc-300',
        'dark:border-zinc-700/50 dark:bg-zinc-800/50 dark:hover:border-zinc-600',
        className
      )}
    >
      {/* Warning indicator */}
      {warning && (
        <div className="absolute top-3 right-3 text-amber-500">
          <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
            />
          </svg>
        </div>
      )}

      {/* Icon */}
      <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-3 bg-zinc-100 text-zinc-500 dark:bg-zinc-700/50 dark:text-zinc-400">
        {icon}
      </div>

      {/* Label */}
      <div className="text-sm text-zinc-500 dark:text-zinc-400 mb-1">{label}</div>

      {/* Value */}
      <div className="text-base font-medium text-zinc-950 dark:text-white truncate">{value}</div>
    </div>
  )
}

// ============================================================================
// Preset Icons
// ============================================================================

export const FormatIcons = {
  game: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.5}
        d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"
      />
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.5}
        d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  ),
  calendar: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.5}
        d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
      />
    </svg>
  ),
  team: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.5}
        d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
      />
    </svg>
  ),
  trophy: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.5}
        d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
      />
    </svg>
  ),
  bracket: (
    <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={1.5}
        d="M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5z"
      />
    </svg>
  )
}

// ============================================================================
// Format Cards Grid
// ============================================================================

import { UI_LABELS } from '../constants'

interface FormatCardsProps {
  game: string
  readyWindow?: string
  teamSize?: string
  prize?: string
  prizeWarning?: boolean
  format?: string
}

export function FormatCards({
  game,
  readyWindow,
  teamSize,
  prize,
  prizeWarning,
  format
}: FormatCardsProps) {
  const labels = UI_LABELS.format
  console.log('formatCards', {
    game,
    readyWindow,
    teamSize,
    prize,
    prizeWarning,
    format
  })
  return (
    <div>
      <h2 className="text-xl font-semibold text-zinc-950 dark:text-white mb-4">{labels.title}</h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <FormatCard icon={FormatIcons.game} label={labels.game} value={game} />
        {readyWindow && (
          <FormatCard icon={FormatIcons.calendar} label={labels.readyWindow} value={readyWindow} />
        )}
        {teamSize && (
          <FormatCard icon={FormatIcons.team} label={labels.teamSize} value={teamSize} />
        )}
        {prize && (
          <FormatCard
            icon={FormatIcons.trophy}
            label={labels.customPrize}
            value={prize}
            warning={prizeWarning}
          />
        )}
        {format && (
          <FormatCard icon={FormatIcons.bracket} label={labels.formatType} value={format} />
        )}
      </div>
    </div>
  )
}

