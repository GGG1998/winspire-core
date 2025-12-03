/**
 * PlayersListView Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Full list of tournament participants with details
 */

import { Avatar } from '../../../shared/components/ui/avatar'
import { Badge } from '../../../shared/components/ui/badge'
import type { Tournament, TournamentParticipant } from '../types'

interface PlayersListViewProps {
  tournament: Tournament
  participants: TournamentParticipant[]
}

function formatRegistrationDate(date: Date): string {
  return new Intl.DateTimeFormat('pl-PL', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

export function PlayersListView({ tournament, participants }: PlayersListViewProps) {
  const maxSlots = tournament.format?.maxSlots || 16
  const availableSlots = maxSlots - participants.length

  if (participants.length === 0) {
    return (
      <div className="py-12 text-center">
        <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
          <svg className="w-8 h-8 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-zinc-950 dark:text-white mb-2">Brak uczestników</h3>
        <p className="text-zinc-500 dark:text-zinc-400">Nikt jeszcze nie dołączył do turnieju.</p>
        <p className="text-sm text-zinc-400 dark:text-zinc-500 mt-2">
          Dostępne miejsca: {maxSlots}
        </p>
      </div>
    )
  }

  return (
    <div className="py-6">
      {/* Header stats */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold text-zinc-950 dark:text-white">
            Lista uczestników
          </h3>
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            {participants.length} zarejestrowanych • {availableSlots} wolnych miejsc
          </p>
        </div>
        <div className="flex items-center gap-4 text-sm">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-emerald-500" />
            <span className="text-zinc-600 dark:text-zinc-300">
              Gotowi: {participants.filter(p => p.isReady).length}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 rounded-full bg-amber-500" />
            <span className="text-zinc-600 dark:text-zinc-300">
              Oczekują: {participants.filter(p => !p.isReady).length}
            </span>
          </div>
        </div>
      </div>

      {/* Participants table */}
      <div className="rounded-xl border border-zinc-200 dark:border-zinc-700/50 overflow-hidden">
        <table className="w-full">
          <thead className="bg-zinc-50 dark:bg-zinc-800/50">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                #
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                Gracz
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                Status
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-zinc-500 dark:text-zinc-400 uppercase tracking-wide">
                Data rejestracji
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-200 dark:divide-zinc-700/50 bg-white dark:bg-zinc-800/30">
            {participants.map((participant, index) => (
              <tr key={participant.id} className="hover:bg-zinc-50 dark:hover:bg-zinc-800/50 transition-colors">
                <td className="px-4 py-4 text-sm text-zinc-500 dark:text-zinc-400">
                  {index + 1}
                </td>
                <td className="px-4 py-4">
                  <div className="flex items-center gap-3">
                    <Avatar
                      src={participant.avatarUrl}
                      initials={participant.name.slice(0, 2).toUpperCase()}
                      alt={participant.name}
                      className="w-10 h-10"
                    />
                    <div>
                      <div className="font-medium text-zinc-950 dark:text-white">
                        {participant.name}
                      </div>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-4">
                  {participant.isReady ? (
                    <Badge color="emerald">Gotowy</Badge>
                  ) : (
                    <Badge color="amber">Oczekuje</Badge>
                  )}
                </td>
                <td className="px-4 py-4 text-sm text-zinc-500 dark:text-zinc-400">
                  {formatRegistrationDate(participant.registeredAt)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Slots progress */}
      <div className="mt-6">
        <div className="flex justify-between text-sm text-zinc-500 dark:text-zinc-400 mb-2">
          <span>{participants.length} / {maxSlots} miejsc zajętych</span>
          <span>{Math.round((participants.length / maxSlots) * 100)}%</span>
        </div>
        <div className="h-2 bg-zinc-200 dark:bg-zinc-700 rounded-full overflow-hidden">
          <div
            className="h-full bg-cyan-500 rounded-full transition-all duration-300"
            style={{ width: `${Math.min((participants.length / maxSlots) * 100, 100)}%` }}
          />
        </div>
      </div>
    </div>
  )
}

