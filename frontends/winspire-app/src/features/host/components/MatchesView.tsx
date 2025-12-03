/**
 * MatchesView Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Displays list of all matches in the tournament
 */

import { Badge } from '../../../shared/components/ui/badge'
import { UI_LABELS } from '../constants'
import type { Tournament } from '../types'

interface MatchesViewProps {
  tournament: Tournament
}

// Mock match data for demo purposes
interface Match {
  id: string
  round: number
  matchNumber: number
  player1?: { name: string; score?: number }
  player2?: { name: string; score?: number }
  status: 'pending' | 'in_progress' | 'completed'
  scheduledTime?: Date
}

function generateMockMatches(tournament: Tournament): Match[] {
  const participants = tournament.participants || []
  const maxSlots = tournament.format?.maxSlots || 8
  
  // Generate matches based on single elimination
  const rounds = Math.ceil(Math.log2(maxSlots))
  const matches: Match[] = []
  
  let matchNumber = 1
  for (let round = 1; round <= rounds; round++) {
    const matchesInRound = Math.pow(2, rounds - round)
    for (let i = 0; i < matchesInRound; i++) {
      const isFirstRound = round === 1
      matches.push({
        id: `match-${matchNumber}`,
        round,
        matchNumber: matchNumber++,
        player1: isFirstRound && participants[i * 2] 
          ? { name: participants[i * 2].name }
          : undefined,
        player2: isFirstRound && participants[i * 2 + 1]
          ? { name: participants[i * 2 + 1].name }
          : undefined,
        status: 'pending',
        scheduledTime: tournament.startTime
      })
    }
  }
  
  return matches
}

function getRoundName(round: number, totalRounds: number): string {
  if (round === totalRounds) return 'Finał'
  if (round === totalRounds - 1) return 'Półfinały'
  if (round === totalRounds - 2) return 'Ćwierćfinały'
  return `Runda ${round}`
}

function getMatchStatusBadge(status: Match['status']) {
  switch (status) {
    case 'in_progress':
      return <Badge color="emerald">Na żywo</Badge>
    case 'completed':
      return <Badge color="zinc">Zakończony</Badge>
    default:
      return <Badge color="cyan">Oczekuje</Badge>
  }
}

export function MatchesView({ tournament }: MatchesViewProps) {
  const matches = generateMockMatches(tournament)
  const maxSlots = tournament.format?.maxSlots || 8
  const totalRounds = Math.ceil(Math.log2(maxSlots))
  
  // Group matches by round
  const matchesByRound = matches.reduce((acc, match) => {
    if (!acc[match.round]) acc[match.round] = []
    acc[match.round].push(match)
    return acc
  }, {} as Record<number, Match[]>)

  if (matches.length === 0) {
    return (
      <div className="py-12 text-center">
        <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
          <svg className="w-8 h-8 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-zinc-950 dark:text-white mb-2">Brak meczów</h3>
        <p className="text-zinc-500 dark:text-zinc-400">Mecze pojawią się po rozpoczęciu turnieju.</p>
      </div>
    )
  }

  return (
    <div className="py-6 space-y-8">
      {Object.entries(matchesByRound).map(([round, roundMatches]) => (
        <div key={round}>
          {/* Round header */}
          <h3 className="text-lg font-semibold text-zinc-950 dark:text-white mb-4">
            {getRoundName(parseInt(round), totalRounds)}
          </h3>
          
          {/* Matches list */}
          <div className="space-y-3">
            {roundMatches.map(match => (
              <div
                key={match.id}
                className="rounded-xl border border-zinc-200 dark:border-zinc-700/50 bg-white dark:bg-zinc-800/50 p-4"
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-sm text-zinc-500 dark:text-zinc-400">
                    Mecz #{match.matchNumber}
                  </span>
                  {getMatchStatusBadge(match.status)}
                </div>
                
                {/* Players */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-zinc-100 dark:bg-zinc-700 flex items-center justify-center text-xs font-medium text-zinc-600 dark:text-zinc-300">
                        {match.player1?.name?.slice(0, 2).toUpperCase() || '?'}
                      </div>
                      <span className="font-medium text-zinc-950 dark:text-white">
                        {match.player1?.name || 'TBD'}
                      </span>
                    </div>
                    {match.player1?.score !== undefined && (
                      <span className="text-lg font-bold text-zinc-950 dark:text-white">
                        {match.player1.score}
                      </span>
                    )}
                  </div>
                  
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-zinc-100 dark:bg-zinc-700 flex items-center justify-center text-xs font-medium text-zinc-600 dark:text-zinc-300">
                        {match.player2?.name?.slice(0, 2).toUpperCase() || '?'}
                      </div>
                      <span className="font-medium text-zinc-950 dark:text-white">
                        {match.player2?.name || 'TBD'}
                      </span>
                    </div>
                    {match.player2?.score !== undefined && (
                      <span className="text-lg font-bold text-zinc-950 dark:text-white">
                        {match.player2.score}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

