/**
 * PlayersPanel Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Sidebar panel showing registered players, ready count, and slots (Polish labels)
 */

import { Avatar } from '../../../shared/components/ui/avatar'
import type { TournamentParticipant } from '../types'
import { UI_LABELS } from '../constants'

interface PlayersPanelProps {
  participants: TournamentParticipant[]
  maxSlots: number
  className?: string
}

export function PlayersPanel({ participants, maxSlots, className }: PlayersPanelProps) {
  const registeredCount = participants.length
  const readyCount = participants.filter(p => p.isReady).length
  const displayParticipants = participants.slice(0, 5)
  const labels = UI_LABELS.players

  return (
    <div className={className}>
      <h2 className="text-xl font-semibold text-zinc-950 dark:text-white mb-4">{labels.title}</h2>
      
      <div className="rounded-xl border border-zinc-200 bg-white p-5 dark:border-zinc-700/50 dark:bg-zinc-800/50">
        {/* Stats */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          <div className="text-center">
            <div className="text-xs text-zinc-500 dark:text-zinc-400 uppercase tracking-wide mb-1">{labels.registered}</div>
            <div className="text-2xl font-bold text-zinc-950 dark:text-white">{registeredCount}</div>
          </div>
          <div className="text-center border-x border-zinc-200 dark:border-zinc-700/50">
            <div className="text-xs text-zinc-500 dark:text-zinc-400 uppercase tracking-wide mb-1 flex items-center justify-center gap-1">
              {labels.ready}
              <svg className="w-3.5 h-3.5 text-zinc-400 dark:text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div className="text-2xl font-bold text-zinc-950 dark:text-white">{readyCount}</div>
          </div>
          <div className="text-center">
            <div className="text-xs text-zinc-500 dark:text-zinc-400 uppercase tracking-wide mb-1">{labels.slots}</div>
            <div className="text-2xl font-bold text-zinc-950 dark:text-white">{maxSlots}</div>
          </div>
        </div>

        {/* Avatar stack */}
        {displayParticipants.length > 0 && (
          <div className="flex items-center mb-4">
            <div className="flex -space-x-2">
              {displayParticipants.map((participant, index) => (
                <Avatar
                  key={participant.id}
                  src={participant.avatarUrl}
                  initials={participant.name.slice(0, 2).toUpperCase()}
                  alt={participant.name}
                  className="w-10 h-10 ring-2 ring-white bg-zinc-200 dark:ring-zinc-800 dark:bg-zinc-700"
                  style={{ zIndex: displayParticipants.length - index }}
                />
              ))}
              {participants.length > 5 && (
                <div 
                  className="w-10 h-10 rounded-full bg-zinc-200 ring-2 ring-white flex items-center justify-center text-xs font-medium text-zinc-600 dark:bg-zinc-700 dark:ring-zinc-800 dark:text-zinc-300"
                  style={{ zIndex: 0 }}
                >
                  +{participants.length - 5}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Participant text */}
        <p className="text-sm text-zinc-600 dark:text-zinc-300">
          {participants.length > 0 ? (
            <>
              <span className="font-medium text-zinc-950 dark:text-white">
                {participants.slice(0, 2).map(p => p.name).join(', ')}
              </span>
              {participants.length > 2 && (
                <>
                  {` ${labels.and} `}
                  <span className="font-medium text-zinc-950 dark:text-white">
                    {participants.length - 2} {labels.others}
                  </span>
                </>
              )}
              {` ${labels.areRegistered}`}
            </>
          ) : (
            <span className="text-zinc-400">{labels.noParticipants}</span>
          )}
        </p>

        {/* Progress bar */}
        <div className="mt-4">
          <div className="flex justify-between text-xs text-zinc-500 dark:text-zinc-400 mb-1">
            <span>{registeredCount} {labels.registered.toLowerCase()}</span>
            <span>{maxSlots - registeredCount} {labels.slotsLeft}</span>
          </div>
          <div className="h-2 bg-zinc-200 dark:bg-zinc-700 rounded-full overflow-hidden">
            <div
              className="h-full bg-cyan-500 rounded-full transition-all duration-300"
              style={{ width: `${Math.min((registeredCount / maxSlots) * 100, 100)}%` }}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

