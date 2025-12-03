/**
 * TournamentHeader Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Hero section with tournament banner, logo, name, countdown and join button (Polish labels)
 */

import { memo } from 'react'
import { Badge } from '../../../shared/components/ui/badge'
import { Button } from '../../../shared/components/ui/button'
import { useCountdown } from '../hooks/useCountdown'
import { TournamentSettings } from './TournamentSettings'
import type { Tournament, TournamentStatus } from '../types'
import { TOURNAMENT_STATUS_CONFIG, UI_LABELS, JOIN_BUTTON_MINUTES_BEFORE_START } from '../constants'

interface TournamentHeaderProps {
  tournament: Tournament
  status: TournamentStatus
  /** Whether the current user is the owner/host of this tournament */
  isOwner?: boolean
  onJoin?: () => void
  onConfirmParticipation?: () => void
  onEdit?: () => void
  onPublished?: () => void
  onOpenedRegistration?: () => void
  onStarted?: () => void
  onCancelled?: () => void
}

// Default banner for tournaments without custom image
const DEFAULT_BANNER = 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=1920&q=80'

/**
 * Format date for display
 */
function formatDate(date: Date): string {
  return new Intl.DateTimeFormat('en-US', {
    weekday: 'short',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    timeZoneName: 'short'
  }).format(date)
}

function TournamentHeaderComponent({
  tournament,
  status,
  isOwner = false,
  onJoin,
  onConfirmParticipation,
  onEdit,
  onPublished,
  onOpenedRegistration,
  onStarted,
  onCancelled
}: TournamentHeaderProps) {
  const { formatted: countdown, formattedShort } = useCountdown(tournament.startTime)
  const bannerUrl = tournament.bannerUrl || DEFAULT_BANNER
  const statusConfig = TOURNAMENT_STATUS_CONFIG[status]
  const labels = UI_LABELS.detail

  // Calculate minutes until tournament start
  const minutesUntilStart = Math.floor((tournament.startTime.getTime() - Date.now()) / (1000 * 60))
  
  // Determine if user can confirm participation (take a spot)
  const canConfirmParticipation = (
    (status === 'registration_open' || status === 'scheduled') &&
    (!tournament.userParticipationStatus || tournament.userParticipationStatus === 'registered') &&
    (!tournament.format?.maxSlots || !tournament.participantCount || tournament.participantCount < tournament.format.maxSlots)
  )
  console.log(canConfirmParticipation)
  // Determine if user can join tournament
  const canJoinTournament = (
    (tournament.userParticipationStatus === 'confirmed' || tournament.userParticipationStatus === 'checked_in') &&
    (status === 'registration_open' || status === 'registration_closed' || status === 'scheduled') &&
    minutesUntilStart <= JOIN_BUTTON_MINUTES_BEFORE_START &&
    minutesUntilStart > 0
  )

  return (
    <div className="relative">
      {/* Banner Image */}
      <div className="relative h-48 md:h-64 lg:h-80 overflow-hidden rounded-t-xl">
        <img
          src={bannerUrl}
          alt={tournament.name}
          className="w-full h-full object-cover"
        />
        {/* Gradient overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-zinc-900/90 via-zinc-900/50 to-transparent" />
        
        {/* Settings dropdown for owner */}
        {isOwner && (
          <div className="absolute top-4 right-4">
            <TournamentSettings
              tournament={tournament}
              status={status}
              onEdit={onEdit}
              onPublished={onPublished}
              onOpenedRegistration={onOpenedRegistration}
              onStarted={onStarted}
              onCancelled={onCancelled}
            />
          </div>
        )}
      </div>

      {/* Content overlay */}
      <div className="absolute bottom-0 left-0 right-0 p-6 md:p-8">
        <div className="flex flex-col md:flex-row md:items-end gap-6">
          {/* Game Logo */}
          <div className="flex-shrink-0">
            <div className="w-24 h-24 md:w-32 md:h-32 rounded-xl bg-zinc-800 border-2 border-zinc-700 overflow-hidden shadow-xl">
              {tournament.gameLogoUrl ? (
                <img
                  src={tournament.gameLogoUrl}
                  alt={tournament.game}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-zinc-400">
                  <svg className="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              )}
            </div>
          </div>

          {/* Tournament Info */}
          <div className="flex-1 min-w-0">
            {/* Breadcrumb */}
            {tournament.organizer && (
              <div className="flex items-center gap-2 text-sm text-zinc-400 mb-2">
                {tournament.organizer.logoUrl && (
                  <img
                    src={tournament.organizer.logoUrl}
                    alt={tournament.organizer.name}
                    className="w-5 h-5 rounded"
                  />
                )}
                <span>{tournament.organizer.name}</span>
                <span className="text-zinc-600">›</span>
                <span>Turnieje</span>
              </div>
            )}

            {/* Title */}
            <h1 className="text-2xl md:text-3xl lg:text-4xl font-bold text-white mb-2 truncate">
              {tournament.name}
            </h1>

            {/* Meta info */}
            <div className="flex flex-wrap items-center gap-3 text-sm text-zinc-300">
              <span>{formattedShort}</span>
              <span className="text-zinc-600">•</span>
              <span>{formatDate(tournament.startTime)}</span>
            </div>

            {/* Status badge */}
            <div className="mt-3">
              <Badge color={statusConfig.badgeColor}>
                {statusConfig.label}
              </Badge>
            </div>
          </div>

          {/* Countdown & Action Buttons */}
          <div className="flex flex-col items-end gap-3 flex-shrink-0">
            {(status === 'draft' || status === 'registration_open' || status === 'registration_closed' || status === 'scheduled') && (
              <>
                <div className="text-right">
                  <div className="text-xs text-zinc-400 uppercase tracking-wide">{labels.startsIn}</div>
                  <div className="text-2xl md:text-3xl font-mono font-bold text-white tabular-nums">
                    {countdown}
                  </div>
                </div>
                {canConfirmParticipation && (
                  <Button color="orange" onClick={onConfirmParticipation}>
                    Zajmij miejsce
                  </Button>
                )}
                {canJoinTournament && (
                  <Button color="cyan" onClick={onJoin}>
                    {labels.joinTournament}
                  </Button>
                )}
              </>
            )}
            {status === 'started' && (
              <Badge color="emerald">
                <span className="relative flex h-2 w-2 mr-1">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                </span>
                {labels.tournamentInProgress}
              </Badge>
            )}
            {status === 'completed' && (
              <Badge color="zinc">
                {labels.tournamentEnded}
              </Badge>
            )}
            {status === 'cancelled' && (
              <div className="flex flex-col items-end gap-2">
                <Badge color="red">
                  <svg className="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  Anulowany
                </Badge>
                <span className="text-xs text-zinc-400">Ten turniej został anulowany</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

TournamentHeaderComponent.displayName = 'TournamentHeader'

export const TournamentHeader = memo(TournamentHeaderComponent)

