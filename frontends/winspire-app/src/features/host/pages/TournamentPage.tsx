/**
 * TournamentPage Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Main page for tournament management with sidebar layout
 */

import { useState, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useAuth } from '../../auth'
import { useTournaments } from '../hooks/useTournaments'
import { TournamentTable } from '../components/TournamentTable'
import { CreateTournamentModal } from '../components/CreateTournamentModal'
import { TournamentSuccessModal } from '../components/TournamentSuccessModal'
import { HostLayout } from '../layouts/HostLayout'
import { Button } from '../../../shared/components/ui/button'
import { UI_LABELS } from '../constants'
import type { Tournament } from '../types'
import { PlusIcon } from '@heroicons/react/20/solid'

/**
 * Tournament management page for streamers
 */
export function TournamentPage() {
  const { streamerId } = useParams<{ streamerId: string }>()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { tournaments, isLoading, error, refreshTournaments } = useTournaments()
  
  // Modal states
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [isSuccessModalOpen, setIsSuccessModalOpen] = useState(false)
  const [createdTournament, setCreatedTournament] = useState<Tournament | null>(null)

  /**
   * Handle successful tournament creation
   */
  const handleCreateSuccess = useCallback((tournament: Tournament) => {
    setCreatedTournament(tournament)
    setIsCreateModalOpen(false)
    setIsSuccessModalOpen(true)
  }, [])

  /**
   * Handle navigation to tournament detail/room
   */
  const handleGoToRoom = useCallback((roomLink: string) => {
    // Navigate to room (external link or internal route)
    if (roomLink.startsWith('http')) {
      window.location.href = roomLink
    } else {
      navigate(roomLink)
    }
  }, [navigate])

  /**
   * Handle navigation to tournament detail page
   */
  const handleTournamentClick = useCallback((tournament: Tournament) => {
    navigate(`/h/${streamerId}/tournaments/${tournament.id}`)
  }, [navigate, streamerId])

  /**
   * Handle success modal close
   */
  const handleSuccessModalClose = useCallback(() => {
    setIsSuccessModalOpen(false)
    setCreatedTournament(null)
    // Refresh list to ensure new tournament is visible
    refreshTournaments()
  }, [refreshTournaments])

  return (
    <HostLayout>
      <div className="space-y-8">
        {/* Page header */}
        <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-zinc-950 dark:text-white">
              {UI_LABELS.page.title}
            </h1>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              {UI_LABELS.page.subtitle}
            </p>
          </div>
          {user?.profileType === 'streamer' && (
            <Button
              color="cyan"
              onClick={() => setIsCreateModalOpen(true)}
            >
              <PlusIcon />
              {UI_LABELS.page.createButton}
            </Button>
          )}
        </header>

        {/* Error display */}
        {error && (
          <div className="rounded-xl border border-red-200 dark:border-red-900/50 bg-red-50 dark:bg-red-900/20 p-4">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-500 dark:text-red-400"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="flex-1">
                <p className="text-sm font-medium text-red-800 dark:text-red-200">
                  {error.message}
                </p>
                <button
                  onClick={() => refreshTournaments()}
                  className="mt-2 text-sm font-medium text-red-700 dark:text-red-300 underline hover:text-red-600 dark:hover:text-red-200"
                >
                  {UI_LABELS.error.retry}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Tournaments table */}
        <TournamentTable
          tournaments={tournaments}
          isLoading={isLoading}
          error={error}
          onRefresh={refreshTournaments}
          onTournamentClick={handleTournamentClick}
        />

        {/* Create tournament modal */}
        <CreateTournamentModal
          isOpen={isCreateModalOpen}
          onClose={() => setIsCreateModalOpen(false)}
          onSuccess={handleCreateSuccess}
        />

        {/* Success modal */}
        {createdTournament && (
          <TournamentSuccessModal
            isOpen={isSuccessModalOpen}
            tournament={createdTournament}
            onClose={handleSuccessModalClose}
            onGoToRoom={handleGoToRoom}
          />
        )}
      </div>
    </HostLayout>
  )
}
