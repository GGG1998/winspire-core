import React from 'react';
import { Avatar } from '../../../shared/components/ui/avatar';
import type { PlayerInfo } from '../types';

interface MatchResultProps {
  winner: PlayerInfo | null;
  loser: PlayerInfo | null;
  currentUserId?: string;
  scoreWinner?: number | null;
  scoreLoser?: number | null;
  resultSource?: 'game_api' | 'manual_host' | 'walkover';
  onContinue?: () => void;
}

/**
 * MatchResult Component
 * 
 * Post-match summary display
 * Features:
 * - Winner/loser display with avatars
 * - Score display
 * - Result source indicator
 * - "You won!" / "You lost" messaging for current player
 * - Continue button to return to tournament
 */
export function MatchResult({ 
  winner, 
  loser, 
  currentUserId,
  scoreWinner = null,
  scoreLoser = null,
  resultSource = 'game_api',
  onContinue 
}: MatchResultProps) {
  const isCurrentUserWinner = winner?.id === currentUserId;
  const isCurrentUserLoser = loser?.id === currentUserId;

  // Helper to get player initials
  const getInitials = (displayName: string): string => {
    return displayName
      .split(' ')
      .map((word) => word[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  // Get result source label
  const getResultSourceLabel = () => {
    switch (resultSource) {
      case 'game_api':
        return 'Wynik z gry';
      case 'manual_host':
        return 'Wynik wprowadzony przez gospodarza';
      case 'walkover':
        return 'Walkower (nieobecność przeciwnika)';
      default:
        return '';
    }
  };

  return (
    <div className="w-full max-w-4xl mx-auto">
      {/* Header Message */}
      <div className="text-center mb-8">
        {isCurrentUserWinner && (
          <div className="mb-6">
            <div className="text-8xl mb-4">🏆</div>
            <h2 className="text-4xl font-bold text-green-600 dark:text-green-400 mb-2">
              Wygrałeś!
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-400">
              Gratulacje! Awansujesz do następnej rundy.
            </p>
          </div>
        )}

        {isCurrentUserLoser && (
          <div className="mb-6">
            <div className="text-8xl mb-4">💔</div>
            <h2 className="text-4xl font-bold text-gray-600 dark:text-gray-400 mb-2">
              Przegrałeś
            </h2>
            <p className="text-lg text-gray-600 dark:text-gray-400">
              Dobra gra! Dziękujemy za udział.
            </p>
          </div>
        )}

        {!isCurrentUserWinner && !isCurrentUserLoser && (
          <div className="mb-6">
            <div className="text-8xl mb-4">🎮</div>
            <h2 className="text-4xl font-bold text-gray-900 dark:text-white mb-2">
              Mecz zakończony
            </h2>
          </div>
        )}
      </div>

      {/* Match Result Card */}
      <div className="rounded-lg border-2 border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-8 mb-6">
        <div className="grid grid-cols-3 gap-6 items-center">
          {/* Winner */}
          <div className="text-center">
            <div className="mb-3">
              {winner?.avatarUrl ? (
                <img
                  src={winner.avatarUrl}
                  alt={winner.displayName}
                  className="size-24 rounded-full border-4 border-green-400 dark:border-green-600 mx-auto"
                />
              ) : winner ? (
                <Avatar
                  className="size-24 text-2xl mx-auto border-4 border-green-400 dark:border-green-600"
                  initials={getInitials(winner.displayName)}
                />
              ) : (
                <div className="size-24 rounded-full bg-gray-200 dark:bg-gray-700 mx-auto" />
              )}
            </div>
            <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-1">
              {winner?.displayName || 'Nieznany'}
            </h3>
            <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 dark:bg-green-900/40 text-green-800 dark:text-green-200">
              ✓ Zwycięzca
            </span>
            {scoreWinner !== null && (
              <p className="mt-2 text-3xl font-bold text-green-600 dark:text-green-400">
                {scoreWinner}
              </p>
            )}
          </div>

          {/* VS Separator */}
          <div className="text-center">
            <div className="inline-flex items-center justify-center size-16 rounded-full bg-gray-100 dark:bg-gray-700">
              <span className="text-2xl font-bold text-gray-600 dark:text-gray-400">
                VS
              </span>
            </div>
          </div>

          {/* Loser */}
          <div className="text-center">
            <div className="mb-3">
              {loser?.avatarUrl ? (
                <img
                  src={loser.avatarUrl}
                  alt={loser.displayName}
                  className="size-24 rounded-full border-4 border-gray-300 dark:border-gray-600 mx-auto opacity-60"
                />
              ) : loser ? (
                <Avatar
                  className="size-24 text-2xl mx-auto border-4 border-gray-300 dark:border-gray-600 opacity-60"
                  initials={getInitials(loser.displayName)}
                />
              ) : (
                <div className="size-24 rounded-full bg-gray-200 dark:bg-gray-700 mx-auto opacity-60" />
              )}
            </div>
            <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-1 opacity-60">
              {loser?.displayName || 'Nieznany'}
            </h3>
            <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
              Przegrany
            </span>
            {scoreLoser !== null && (
              <p className="mt-2 text-3xl font-bold text-gray-400 dark:text-gray-500">
                {scoreLoser}
              </p>
            )}
          </div>
        </div>

        {/* Result Source */}
        <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700 text-center">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {getResultSourceLabel()}
          </p>
        </div>
      </div>

      {/* Continue Button */}
      {onContinue && (
        <div className="text-center">
          <button
            onClick={onContinue}
            className="px-8 py-3 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg font-semibold transition-colors"
          >
            Wróć do turnieju
          </button>
        </div>
      )}
    </div>
  );
}

