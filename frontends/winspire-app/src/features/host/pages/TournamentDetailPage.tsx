/**
 * TournamentDetailPage Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Detail view for a single tournament with tabs (Przegląd, Drabinka, Mecze, Gracze, Wyniki)
 */

import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { TournamentHeader } from '../components/TournamentHeader'
import { TournamentTabs } from '../components/TournamentTabs'
import { TournamentOverview } from '../components/TournamentOverview'
import { BracketView } from '../components/BracketView'
import { MatchesView } from '../components/MatchesView'
import { PlayersListView } from '../components/PlayersListView'
import { ResultsView } from '../components/ResultsView'
import { HostLayout } from '../layouts/HostLayout'
import { tournamentApi } from '../api/tournamentApi'
import { useAuth } from '../../auth/context/AuthContext'
import { UI_LABELS } from '../constants'
import type { Tournament, TournamentTab, TournamentError, TournamentParticipant } from '../types'

export function TournamentDetailPage() {
  const { streamerId, tournamentId } = useParams<{ streamerId: string; tournamentId: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  
  const [tournament, setTournament] = useState<Tournament | null>(null)
  const [participants, setParticipants] = useState<TournamentParticipant[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<TournamentError | null>(null)
  const [activeTab, setActiveTab] = useState<TournamentTab>('overview')

  const labels = UI_LABELS.detail
  
  // Check if current user is the owner of this tournament
  const isOwner = Boolean(user && tournament && (
    user.id === tournament.creatorId || 
    user.id === streamerId
  ))

  // Fetch tournament data
  useEffect(() => {
    if (!tournamentId || !streamerId) {
      setError({ type: 'validation', message: 'Tournament ID and Streamer ID are required' })
      setIsLoading(false)
      return
    }

    const fetchTournament = async () => {
      try {
        setIsLoading(true)
        setError(null)
        const data = await tournamentApi.getTournament(tournamentId, streamerId)
        setTournament(data)
      } catch (err) {
        // Check if it's a 403 Forbidden error
        const errorMessage = err instanceof Error ? err.message : 'Failed to load tournament'
        const isForbidden = errorMessage.includes('403') || errorMessage.includes('permission') || errorMessage.includes('Forbidden')
        
        setError({
          type: isForbidden ? 'authorization' : 'unknown',
          message: isForbidden 
            ? 'Nie masz uprawnień do przeglądania tego turnieju' 
            : errorMessage,
          details: err
        })
      } finally {
        setIsLoading(false)
      }
    }

    fetchTournament()
  }, [tournamentId, streamerId])

  // Fetch participants list
  useEffect(() => {
    if (!tournamentId || !streamerId) return

    const fetchParticipants = async () => {
      try {
        const data = await tournamentApi.getParticipants(tournamentId, streamerId)
        setParticipants(data)
      } catch (err) {
        console.error('Failed to fetch participants:', err)
        // Don't show error to user, just use empty array
        setParticipants([])
      }
    }

    fetchParticipants()
  }, [tournamentId, streamerId, tournament?.participantCount]) // Reload when count changes

  // Refetch tournament data
  const refetchTournament = useCallback(async () => {
    if (!tournamentId || !streamerId) return
    try {
      const data = await tournamentApi.getTournament(tournamentId, streamerId)
      setTournament(data)
    } catch (err) {
      console.error('Failed to refetch tournament:', err)
    }
  }, [tournamentId, streamerId])

  // Handle registration for tournament (take a spot)
  const handleConfirmParticipation = async () => {
    if (!tournamentId || !streamerId) return

    try {
      await tournamentApi.confirmParticipation(tournamentId, streamerId)
      // Refetch tournament to update participation status (will be 'registered')
      await refetchTournament()
    } catch (err) {
      console.error('Failed to register for tournament:', err)
      setError({
        type: 'unknown',
        message: err instanceof Error ? err.message : 'Nie udało się zarejestrować do turnieju',
        details: err
      })
    }
  }

  // Handle join tournament
  const handleJoin = async () => {
    if (!tournamentId || !streamerId) return

    try {
      const roomLink = await tournamentApi.joinTournament(tournamentId, streamerId)
      // Navigate to room
      if (roomLink.startsWith('http')) {
        window.location.href = roomLink
      } else {
        navigate(roomLink)
      }
    } catch (err) {
      console.error('Failed to join tournament:', err)
      setError({
        type: 'unknown',
        message: err instanceof Error ? err.message : 'Nie udało się dołączyć do turnieju',
        details: err
      })
    }
  }

  // Handle edit tournament
  const handleEdit = () => {
    // TODO: Open edit modal or navigate to edit page
    console.log('Edit tournament:', tournament?.id)
  }

  // Handle tournament published
  const handlePublished = () => {
    refetchTournament()
  }

  // Handle registration opened
  const handleOpenedRegistration = () => {
    refetchTournament()
  }

  // Handle tournament started
  const handleStarted = () => {
    refetchTournament()
  }

  // Handle tournament cancelled
  const handleCancelled = () => {
    // Navigate back to tournaments list after cancellation
    navigate(`/h/${streamerId}/tournaments`)
  }

  // Loading state
  if (isLoading) {
    return (
      <HostLayout>
        <div className="flex items-center justify-center py-20">
          <div className="flex items-center gap-3">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-cyan-500" />
            <span className="text-zinc-500 dark:text-zinc-400">{labels.loadingTournament}</span>
          </div>
        </div>
      </HostLayout>
    )
  }

  // Error state
  if (error || !tournament) {
    return (
      <HostLayout>
        <div className="flex items-center justify-center py-20">
          <div className="text-center">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-red-100 dark:bg-red-500/10 flex items-center justify-center">
              <svg className="w-8 h-8 text-red-500 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-zinc-950 dark:text-white mb-2">
              {labels.tournamentNotFound}
            </h2>
            <p className="text-zinc-500 dark:text-zinc-400 mb-4">
              {error?.message || labels.tournamentNotFoundDesc}
            </p>
            <button
              onClick={() => navigate(`/h/${streamerId}/tournaments`)}
              className="text-cyan-600 hover:text-cyan-500 dark:text-cyan-400 dark:hover:text-cyan-300 font-medium"
            >
              {labels.backToTournaments}
            </button>
          </div>
        </div>
      </HostLayout>
    )
  }

  const status = tournament.status

  // Render tab content
  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return <TournamentOverview tournament={tournament} participants={participants} />
      case 'bracket':
        return <BracketView tournament={tournament} />
      case 'matches':
        return <MatchesView tournament={tournament} />
      case 'players':
        return <PlayersListView tournament={tournament} participants={participants} />
      case 'results':
        return <ResultsView tournament={tournament} participants={participants} />
      default:
        return <TournamentOverview tournament={tournament} participants={participants} />
    }
  }

  return (
    <HostLayout>
      <div className="space-y-6">
        {/* Header with banner */}
        <div className="-mx-6 -mt-6 lg:-mx-10 lg:-mt-10">
          <TournamentHeader
            tournament={tournament}
            status={status}
            isOwner={isOwner}
            onJoin={handleJoin}
            onConfirmParticipation={handleConfirmParticipation}
            onEdit={handleEdit}
            onPublished={handlePublished}
            onOpenedRegistration={handleOpenedRegistration}
            onStarted={handleStarted}
            onCancelled={handleCancelled}
          />
        </div>

        {/* Tabs navigation */}
        <TournamentTabs activeTab={activeTab} onTabChange={setActiveTab} />

        {/* Tab content */}
        <div className="pb-6">
          {renderTabContent()}
        </div>
      </div>
    </HostLayout>
  )
}
