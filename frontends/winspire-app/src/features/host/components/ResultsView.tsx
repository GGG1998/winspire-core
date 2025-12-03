/**
 * ResultsView Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Displays tournament results and final standings
 */

import { Avatar } from '../../../shared/components/ui/avatar'
import type { Tournament, TournamentParticipant } from '../types'

interface ResultsViewProps {
  tournament: Tournament
  participants: TournamentParticipant[]
}

// Podium medal colors
const MEDAL_COLORS = {
  1: { bg: 'bg-amber-100 dark:bg-amber-500/20', text: 'text-amber-600 dark:text-amber-400', border: 'border-amber-300 dark:border-amber-500/50' },
  2: { bg: 'bg-zinc-100 dark:bg-zinc-400/20', text: 'text-zinc-500 dark:text-zinc-300', border: 'border-zinc-300 dark:border-zinc-500/50' },
  3: { bg: 'bg-orange-100 dark:bg-orange-500/20', text: 'text-orange-600 dark:text-orange-400', border: 'border-orange-300 dark:border-orange-500/50' }
}

function getMedalEmoji(place: number): string {
  switch (place) {
    case 1: return '🥇'
    case 2: return '🥈'
    case 3: return '🥉'
    default: return ''
  }
}

export function ResultsView({ tournament, participants }: ResultsViewProps) {
  const isCompleted = tournament.isCompleted
  
  // Mock results - in real implementation, this would come from tournament data
  const results = participants.slice(0, Math.min(participants.length, 8)).map((p, i) => ({
    ...p,
    place: i + 1,
    wins: Math.max(0, 4 - i),
    losses: i,
    points: Math.max(0, 100 - i * 15)
  }))

  if (!isCompleted) {
    return (
      <div className="py-12 text-center">
        <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
          <svg className="w-10 h-10 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
        </div>
        
        <h3 className="text-xl font-semibold text-zinc-950 dark:text-white mb-2">
          Turniej w toku
        </h3>
        
        <p className="text-zinc-500 dark:text-zinc-400 max-w-md mx-auto">
          Wyniki zostaną opublikowane po zakończeniu wszystkich meczów turniejowych.
        </p>

        {/* Progress indicator */}
        <div className="mt-8 max-w-xs mx-auto">
          <div className="flex items-center justify-between text-sm text-zinc-500 dark:text-zinc-400 mb-2">
            <span>Postęp turnieju</span>
            <span>W trakcie...</span>
          </div>
          <div className="h-2 bg-zinc-200 dark:bg-zinc-700 rounded-full overflow-hidden">
            <div className="h-full bg-cyan-500 rounded-full animate-pulse" style={{ width: '45%' }} />
          </div>
        </div>
      </div>
    )
  }

  if (results.length === 0) {
    return (
      <div className="py-12 text-center">
        <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
          <svg className="w-8 h-8 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-zinc-950 dark:text-white mb-2">Brak wyników</h3>
        <p className="text-zinc-500 dark:text-zinc-400">Wyniki nie są jeszcze dostępne.</p>
      </div>
    )
  }

  const podium = results.slice(0, 3)
  const restOfResults = results.slice(3)

  return (
    <div className="py-6 space-y-8">
      {/* Podium */}
      <div>
        <h3 className="text-lg font-semibold text-zinc-950 dark:text-white mb-6 text-center">
          Podium
        </h3>
        
        <div className="flex items-end justify-center gap-4">
          {/* 2nd place */}
          {podium[1] && (
            <div className="text-center">
              <div className={`w-24 h-32 rounded-t-xl ${MEDAL_COLORS[2].bg} border-2 ${MEDAL_COLORS[2].border} flex flex-col items-center justify-end pb-3`}>
                <Avatar
                  src={podium[1].avatarUrl}
                  initials={podium[1].name.slice(0, 2).toUpperCase()}
                  className="w-12 h-12 mb-2"
                />
                <span className="text-2xl">{getMedalEmoji(2)}</span>
              </div>
              <div className="mt-2">
                <div className="font-medium text-zinc-950 dark:text-white text-sm truncate max-w-24">
                  {podium[1].name}
                </div>
                <div className="text-xs text-zinc-500 dark:text-zinc-400">{podium[1].points} pkt</div>
              </div>
            </div>
          )}
          
          {/* 1st place */}
          {podium[0] && (
            <div className="text-center">
              <div className={`w-28 h-40 rounded-t-xl ${MEDAL_COLORS[1].bg} border-2 ${MEDAL_COLORS[1].border} flex flex-col items-center justify-end pb-3`}>
                <Avatar
                  src={podium[0].avatarUrl}
                  initials={podium[0].name.slice(0, 2).toUpperCase()}
                  className="w-14 h-14 mb-2"
                />
                <span className="text-3xl">{getMedalEmoji(1)}</span>
              </div>
              <div className="mt-2">
                <div className="font-semibold text-zinc-950 dark:text-white truncate max-w-28">
                  {podium[0].name}
                </div>
                <div className="text-sm text-amber-600 dark:text-amber-400 font-medium">{podium[0].points} pkt</div>
              </div>
            </div>
          )}
          
          {/* 3rd place */}
          {podium[2] && (
            <div className="text-center">
              <div className={`w-24 h-28 rounded-t-xl ${MEDAL_COLORS[3].bg} border-2 ${MEDAL_COLORS[3].border} flex flex-col items-center justify-end pb-3`}>
                <Avatar
                  src={podium[2].avatarUrl}
                  initials={podium[2].name.slice(0, 2).toUpperCase()}
                  className="w-12 h-12 mb-2"
                />
                <span className="text-2xl">{getMedalEmoji(3)}</span>
              </div>
              <div className="mt-2">
                <div className="font-medium text-zinc-950 dark:text-white text-sm truncate max-w-24">
                  {podium[2].name}
                </div>
                <div className="text-xs text-zinc-500 dark:text-zinc-400">{podium[2].points} pkt</div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Prize info */}
      {tournament.prize && (
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-700/50 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-500/10 dark:to-orange-500/10 p-6 text-center">
          <div className="text-sm text-zinc-600 dark:text-zinc-400 mb-1">Nagroda</div>
          <div className="text-xl font-semibold text-zinc-950 dark:text-white">
            {tournament.prize.description}
          </div>
        </div>
      )}

      {/* Full standings table */}
      {restOfResults.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-zinc-950 dark:text-white mb-4">
            Pełna klasyfikacja
          </h3>
          
          <div className="rounded-xl border border-zinc-200 dark:border-zinc-700/50 overflow-hidden">
            <table className="w-full">
              <thead className="bg-zinc-50 dark:bg-zinc-800/50">
                <tr>
                  <th className="px-4 py-3 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                    Miejsce
                  </th>
                  <th className="px-4 py-3 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                    Gracz
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                    W
                  </th>
                  <th className="px-4 py-3 text-center text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                    L
                  </th>
                  <th className="px-4 py-3 text-right text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                    Punkty
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-200 dark:divide-zinc-700/50 bg-white dark:bg-zinc-800/30">
                {restOfResults.map((result) => (
                  <tr key={result.id} className="hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors">
                    <td className="px-4 py-3 text-sm font-medium text-zinc-950 dark:text-white">
                      {result.place}.
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <Avatar
                          src={result.avatarUrl}
                          initials={result.name.slice(0, 2).toUpperCase()}
                          className="w-8 h-8"
                        />
                        <span className="font-medium text-zinc-950 dark:text-white">
                          {result.name}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-center text-sm text-emerald-600 dark:text-emerald-400 font-medium">
                      {result.wins}
                    </td>
                    <td className="px-4 py-3 text-center text-sm text-red-600 dark:text-red-400 font-medium">
                      {result.losses}
                    </td>
                    <td className="px-4 py-3 text-right text-sm font-semibold text-zinc-950 dark:text-white">
                      {result.points}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}

