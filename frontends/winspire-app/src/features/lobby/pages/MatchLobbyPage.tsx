import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../../auth';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { PlayerVsDisplay } from '../components/PlayerVsDisplay';
import { useMatchLobby } from '../hooks/useMatchLobby';
import { LobbyLayout } from '../layouts';
import { ERROR_MESSAGES } from '../constants';

/**
 * Match Lobby Page
 * 
 * Main lobby where two players meet before match starts
 * Features:
 * - Player VS opponent display with avatars
 * - Real-time player join/leave notifications
 * - Match status indicator
 * - Ready button and countdown
 * - Game iframe when match starts
 */
export function MatchLobbyPage() {
  const { tournamentId, matchId } = useParams<{ tournamentId: string; matchId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const {
    matchState,
    isLoading,
    error,
    connectionStatus,
  } = useMatchLobby(matchId || null);

  // Check authorization - verify user is a match participant
  useEffect(() => {
    if (!user || !matchState) return;

    const isParticipant = 
      matchState.player1?.id === user.id || 
      matchState.player2?.id === user.id;

    if (!isParticipant) {
      // User is not a participant in this match
      navigate(`/tournaments/${tournamentId}`, {
        state: { 
          toast: { 
            type: 'error', 
            message: ERROR_MESSAGES.NOT_MATCH_PARTICIPANT 
          } 
        },
      });
    }
  }, [user, matchState, navigate, tournamentId]);

  // Loading state
  if (isLoading) {
    return (
      <LobbyLayout tournamentId={tournamentId} streamerId={matchState?.tournament?.creatorId}>
        <div className="flex items-center justify-center min-h-screen">
          <LoadingSpinner size="lg" />
        </div>
      </LobbyLayout>
    );
  }

  // Error state
  if (error) {
    return (
      <LobbyLayout tournamentId={tournamentId} streamerId={matchState?.tournament?.creatorId}>
        <div className="max-w-4xl px-4 py-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Nie udało się załadować lobby meczu
            </h2>
            <ErrorMessage message={error} />
            <button
              onClick={() => navigate(`/tournaments/${tournamentId}`)}
              className="mt-4 px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700"
            >
              Wróć do turnieju
            </button>
          </div>
        </div>
      </LobbyLayout>
    );
  }

  // No state loaded
  if (!matchState) {
    return (
      <LobbyLayout tournamentId={tournamentId}>
        <div className="max-w-4xl px-4 py-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Lobby meczu niedostępne
            </h2>
            <ErrorMessage message="Nie można załadować stanu lobby" />
            <button
              onClick={() => navigate(`/tournaments/${tournamentId}`)}
              className="mt-4 px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700"
            >
              Wróć do turnieju
            </button>
          </div>
        </div>
      </LobbyLayout>
    );
  }

  return (
    <LobbyLayout 
      tournamentId={matchState.tournament?.id || tournamentId} 
      streamerId={matchState.tournament?.creatorId}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                Lobby Meczu
              </h1>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {matchState.tournament?.name || 'Turniej'} · Runda {matchState.roundNumber}
              </p>
            </div>

            {/* Connection status indicator */}
            <div className="flex items-center gap-2">
              {connectionStatus === 'connected' && (
                <>
                  <div className="size-2 rounded-full bg-green-500 animate-pulse" />
                  <span className="text-sm text-gray-600 dark:text-gray-400">Połączony</span>
                </>
              )}
              {connectionStatus === 'connecting' && (
                <>
                  <LoadingSpinner size="sm" />
                  <span className="text-sm text-gray-600 dark:text-gray-400">Łączenie...</span>
                </>
              )}
              {connectionStatus === 'reconnecting' && (
                <>
                  <LoadingSpinner size="sm" />
                  <span className="text-sm text-amber-600 dark:text-amber-400">Ponowne łączenie...</span>
                </>
              )}
              {connectionStatus === 'disconnected' && (
                <>
                  <div className="size-2 rounded-full bg-red-500" />
                  <span className="text-sm text-red-600 dark:text-red-400">Rozłączony</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Match Status Card */}
        <div className="mb-8 rounded-lg border-2 border-cyan-200 dark:border-cyan-800 bg-cyan-50 dark:bg-cyan-900/20 p-6">
          <div className="flex items-center gap-4">
            <div className="text-5xl">🎮</div>
            <div className="flex-1">
              <h2 className="text-lg font-semibold text-cyan-900 dark:text-cyan-100">
                Status: {matchState.status === 'pending' ? 'Oczekiwanie na graczy' : 
                         matchState.status === 'ready' ? 'Gotowy do rozpoczęcia' :
                         matchState.status === 'started' ? 'W trakcie' :
                         matchState.status === 'completed' ? 'Zakończony' : matchState.status}
              </h2>
              <p className="mt-1 text-sm text-cyan-700 dark:text-cyan-300">
                {matchState.status === 'pending' && 'Czekamy na dołączenie obu graczy'}
                {matchState.status === 'ready' && 'Obaj gracze w lobby - przygotuj się!'}
                {matchState.status === 'started' && 'Mecz w toku'}
                {matchState.status === 'completed' && 'Mecz zakończony'}
              </p>
            </div>
          </div>
        </div>

        {/* Player VS Display */}
        <PlayerVsDisplay
          player1={matchState.player1}
          player2={matchState.player2}
          currentUserId={user?.id}
        />

        {/* Info Footer */}
        <div className="mt-8 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 p-4">
          <div className="flex items-start gap-3">
            <div className="text-xl">ℹ️</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              <p className="font-medium">Informacje o lobby:</p>
              <ul className="mt-2 space-y-1 list-disc list-inside">
                <li>Gdy obaj gracze dołączą, będziecie mogli oznaczyć się jako gotowi</li>
                <li>Po oznaczeniu "Gotowy" przez obu graczy rozpocznie się odliczanie (3, 2, 1)</li>
                <li>Gra załaduje się automatycznie po zakończeniu odliczania</li>
                <li>W razie rozłączenia system automatycznie spróbuje ponownie połączyć</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </LobbyLayout>
  );
}

