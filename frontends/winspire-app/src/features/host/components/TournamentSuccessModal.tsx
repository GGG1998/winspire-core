/**
 * TournamentSuccessModal Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Success modal shown after tournament creation with room link
 */

import { Dialog, DialogPanel, DialogTitle } from '@headlessui/react'
import { QRCodeSVG } from 'qrcode.react'
import { useClipboard } from '../../../shared/hooks/useClipboard'
import { UI_LABELS } from '../constants'
import type { TournamentSuccessModalProps } from '../types'

/**
 * Format date for display in Polish locale
 */
function formatDate(date: Date): string {
  return new Intl.DateTimeFormat('pl-PL', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

/**
 * Get full URL from details link (handles both relative and absolute URLs)
 */
function getFullUrl(detailsLink: string): string {
  if (detailsLink.startsWith('http://') || detailsLink.startsWith('https://')) {
    return detailsLink
  }
  return `${window.location.origin}${detailsLink.startsWith('/') ? '' : '/'}${detailsLink}`
}

/**
 * Modal component shown after successful tournament creation
 */
export function TournamentSuccessModal({
  isOpen,
  tournament,
  onClose,
  onGoToRoom
}: TournamentSuccessModalProps) {
  const { copy, isCopied } = useClipboard()
  const tournamentDetailsLink = `/h/${tournament.creatorId}/tournaments/${tournament.id}`
  const fullUrl = getFullUrl(tournamentDetailsLink)

  const handleCopyLink = async () => {
    await copy(fullUrl)
  }

  const handleGoToRoom = () => {
    onGoToRoom(tournamentDetailsLink)
  }

  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/30" aria-hidden="true" />

      {/* Full-screen container for centering */}
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel className="mx-auto max-w-md w-full rounded-lg bg-white p-6 shadow-xl">
          {/* Success icon */}
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-green-100">
            <svg
              className="h-6 w-6 text-green-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>

          {/* Title */}
          <DialogTitle className="mt-4 text-center text-lg font-semibold text-gray-900">
            {UI_LABELS.success.title}
          </DialogTitle>

          {/* Tournament summary */}
          <div className="mt-4 space-y-3 rounded-lg bg-gray-50 p-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">Nazwa</dt>
              <dd className="mt-1 text-sm text-gray-900">{tournament.name}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Czas rozpoczęcia</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {formatDate(tournament.startTime)}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Gra</dt>
              <dd className="mt-1 text-sm text-gray-900">{tournament.game}</dd>
            </div>
          </div>

          {/* QR Code section */}
          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-700 mb-2 text-center">
              Skanuj kod QR
            </label>
            <div className="flex justify-center">
              <div className="rounded-lg border border-gray-200 bg-white p-4">
                <QRCodeSVG
                  value={fullUrl}
                  size={200}
                  level="M"
                  includeMargin={false}
                />
              </div>
            </div>
          </div>

          {/* Details link section */}
          <div className="mt-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Link do pokoju
            </label>
            <div className="flex items-center gap-2">
              <div className="flex-1 overflow-hidden rounded-md border border-gray-200 bg-gray-50 px-3 py-2">
                <p className="truncate text-sm text-gray-700">
                  {fullUrl}
                </p>
              </div>
              <button
                type="button"
                onClick={handleCopyLink}
                className={`
                  shrink-0 rounded-md px-3 py-2 text-sm font-medium
                  focus:outline-none focus:ring-2 focus:ring-blue-500
                  ${isCopied
                    ? 'bg-green-100 text-green-700'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }
                `}
              >
                {isCopied ? (
                  <span className="flex items-center gap-1">
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    {UI_LABELS.success.copied}
                  </span>
                ) : (
                  <span className="flex items-center gap-1">
                    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
                      />
                    </svg>
                    {UI_LABELS.success.copyLink}
                  </span>
                )}
              </button>
            </div>
          </div>

          {/* Action buttons */}
          <div className="mt-6 flex flex-col gap-3 sm:flex-row-reverse">
            <button
              type="button"
              onClick={handleGoToRoom}
              className="w-full sm:w-auto inline-flex justify-center rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              {UI_LABELS.success.goToRoom}
            </button>
            <button
              type="button"
              onClick={onClose}
              className="w-full sm:w-auto inline-flex justify-center rounded-md bg-white px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
            >
              {UI_LABELS.success.close}
            </button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  )
}



