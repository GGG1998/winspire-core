import React from 'react';
import { Avatar } from '../../../shared/components/ui/avatar';
import type { PlayerInfo } from '../types';

interface PlayerVsDisplayProps {
  player1: PlayerInfo | null;
  player2: PlayerInfo | null;
  currentUserId?: string;
  player1Ready?: boolean;
  player2Ready?: boolean;
}

/**
 * PlayerVsDisplay Component
 * 
 * Shows two players in a left-right VS layout
 * Features:
 * - Player avatars and display names
 * - VS separator in the center
 * - Placeholder state for missing opponent ("Waiting for opponent...")
 * - Highlight current player
 */
export function PlayerVsDisplay({ 
  player1, 
  player2, 
  currentUserId, 
  player1Ready = false, 
  player2Ready = false 
}: PlayerVsDisplayProps) {
  const isPlayer1Current = player1?.id === currentUserId;
  const isPlayer2Current = player2?.id === currentUserId;

  // Helper to get player initials
  const getInitials = (displayName: string): string => {
    return displayName
      .split(' ')
      .map((word) => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  return (
    <div className="w-full">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-8 items-center">
        {/* Player 1 */}
        <div className={`
          rounded-lg border-2 p-4 md:p-6 transition-all
          ${isPlayer1Current 
            ? 'border-cyan-400 dark:border-cyan-600 bg-cyan-50 dark:bg-cyan-900/20' 
            : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800'
          }
        `}>
          {player1 ? (
            <div className="flex flex-col items-center text-center">
              {/* Avatar */}
              {player1.avatarUrl ? (
                <img
                  src={player1.avatarUrl}
                  alt={player1.displayName}
                  className="size-16 md:size-24 rounded-full border-4 border-gray-200 dark:border-gray-700 mb-3 md:mb-4"
                />
              ) : (
                <Avatar
                  className="size-16 md:size-24 text-xl md:text-2xl mb-3 md:mb-4"
                  initials={getInitials(player1.displayName)}
                />
              )}

              {/* Display Name */}
              <h3 className="text-lg md:text-xl font-bold text-gray-900 dark:text-white mb-2">
                {player1.displayName}
              </h3>

              {/* Ready Status Indicator */}
              {player1 && (
                <div className="mt-3">
                  {player1Ready ? (
                    <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium bg-green-100 dark:bg-green-900/40 text-green-800 dark:text-green-200">
                      <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Gotowy
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
                      <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      Oczekiwanie
                    </span>
                  )}
                </div>
              )}

              {/* Current Player Badge */}
              {isPlayer1Current && (
                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-cyan-100 dark:bg-cyan-900/40 text-cyan-800 dark:text-cyan-200 mt-2">
                  ● To Ty
                </span>
              )}
            </div>
          ) : (
            <div className="flex flex-col items-center text-center py-3 md:py-4">
              {/* Placeholder Avatar */}
              <div className="size-16 md:size-24 rounded-full bg-gray-200 dark:bg-gray-700 mb-3 md:mb-4 flex items-center justify-center">
                <svg className="size-8 md:size-12 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>

              {/* Placeholder Text */}
              <h3 className="text-base md:text-lg font-semibold text-gray-500 dark:text-gray-400 mb-2">
                Czekanie na gracza...
              </h3>
              <div className="flex items-center gap-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400" />
                <span className="text-sm text-gray-400 dark:text-gray-500">
                  Oczekiwanie
                </span>
              </div>
            </div>
          )}
        </div>

        {/* VS Separator */}
        <div className="flex items-center justify-center py-3 md:py-0">
          <div className="text-center">
            <div className="inline-flex items-center justify-center size-16 md:size-20 rounded-full bg-gradient-to-br from-red-500 to-pink-500 shadow-lg">
              <span className="text-2xl md:text-3xl font-bold text-white">
                VS
              </span>
            </div>
          </div>
        </div>

        {/* Player 2 */}
        <div className={`
          rounded-lg border-2 p-4 md:p-6 transition-all
          ${isPlayer2Current 
            ? 'border-cyan-400 dark:border-cyan-600 bg-cyan-50 dark:bg-cyan-900/20' 
            : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800'
          }
        `}>
          {player2 ? (
            <div className="flex flex-col items-center text-center">
              {/* Avatar */}
              {player2.avatarUrl ? (
                <img
                  src={player2.avatarUrl}
                  alt={player2.displayName}
                  className="size-16 md:size-24 rounded-full border-4 border-gray-200 dark:border-gray-700 mb-3 md:mb-4"
                />
              ) : (
                <Avatar
                  className="size-16 md:size-24 text-xl md:text-2xl mb-3 md:mb-4"
                  initials={getInitials(player2.displayName)}
                />
              )}

              {/* Display Name */}
              <h3 className="text-lg md:text-xl font-bold text-gray-900 dark:text-white mb-2">
                {player2.displayName}
              </h3>

              {/* Ready Status Indicator */}
              {player2 && (
                <div className="mt-3">
                  {player2Ready ? (
                    <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium bg-green-100 dark:bg-green-900/40 text-green-800 dark:text-green-200">
                      <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                      Gotowy
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
                      <svg className="size-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      Oczekiwanie
                    </span>
                  )}
                </div>
              )}

              {/* Current Player Badge */}
              {isPlayer2Current && (
                <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-cyan-100 dark:bg-cyan-900/40 text-cyan-800 dark:text-cyan-200 mt-2">
                  ● To Ty
                </span>
              )}
            </div>
          ) : (
            <div className="flex flex-col items-center text-center py-3 md:py-4">
              {/* Placeholder Avatar */}
              <div className="size-16 md:size-24 rounded-full bg-gray-200 dark:bg-gray-700 mb-3 md:mb-4 flex items-center justify-center">
                <svg className="size-8 md:size-12 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                </svg>
              </div>

              {/* Placeholder Text */}
              <h3 className="text-base md:text-lg font-semibold text-gray-500 dark:text-gray-400 mb-2">
                Czekanie na przeciwnika...
              </h3>
              <div className="flex items-center gap-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-400" />
                <span className="text-sm text-gray-400 dark:text-gray-500">
                  Oczekiwanie
                </span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

