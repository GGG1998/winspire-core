/**
 * TournamentOverview Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Overview tab content showing format details and players panel
 */

import { FormatCards } from './FormatCard'
import { PlayersPanel } from './PlayersPanel'
import type { Tournament, TournamentParticipant } from '../types'

interface TournamentOverviewProps {
  tournament: Tournament
  participants: TournamentParticipant[]
}

/**
 * Format team size for display (e.g., "1vs1", "2vs2")
 */
function formatTeamSize(size: number): string {
  return `${size}vs${size}`
}

/**
 * Format tournament format type for display
 */
function formatFormatType(type: string): string {
  const labels: Record<string, string> = {
    single_elimination: 'Single Elimination',
    double_elimination: 'Double Elimination',
    round_robin: 'Round Robin',
    swiss: 'Swiss'
  }
  return labels[type] || type
}

/**
 * Format ready window for display
 */
function formatReadyWindow(startsAt: Date, _endsAt: Date): string {
  const options: Intl.DateTimeFormatOptions = {
    weekday: 'short',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }
  
  const start = new Intl.DateTimeFormat('en-US', options).format(startsAt)
  // Truncate if too long
  return start.length > 25 ? start.slice(0, 22) + '...' : start
}

export function TournamentOverview({ tournament, participants }: TournamentOverviewProps) {
  const maxSlots = tournament.format?.maxSlots || 16

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 py-6">
      {/* Format section - takes 2 columns */}
      <div className="lg:col-span-2">
        <FormatCards
          game={tournament.game}
          readyWindow={
            tournament.readyWindow
              ? formatReadyWindow(tournament.readyWindow.startsAt, tournament.readyWindow.endsAt)
              : undefined
          }
          teamSize={tournament.format ? formatTeamSize(tournament.format.teamSize) : undefined}
          prize={tournament.prize?.description}
          prizeWarning={tournament.prize?.type === 'custom'}
          format={tournament.format ? formatFormatType(tournament.format.type) : undefined}
        />
      </div>

      {/* Players panel - takes 1 column */}
      <div className="lg:col-span-1">
        <PlayersPanel participants={participants} maxSlots={maxSlots} />
      </div>
    </div>
  )
}


