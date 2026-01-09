import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../../auth';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { ConnectionIndicator } from '../../../shared/components/ConnectionIndicator';
import { PlayerVsDisplay } from '../components/PlayerVsDisplay';
import { MatchStartCountdown } from '../components/MatchStartCountdown';
import { GameFrame } from '../components/GameFrame';
import { MatchResult } from '../components/MatchResult';
import { DisconnectOverlay } from '../components/DisconnectOverlay';
import { WalkoverButton } from '../components/WalkoverButton';
import { PostMatchModal } from '../components/PostMatchModal';
import { useMatchLobby } from '../hooks/useMatchLobby';
import { useReadyState } from '../hooks/useReadyState';
import { useDisconnect } from '../hooks/useDisconnect';
import { LobbyLayout } from '../layouts';
import { ERROR_MESSAGES } from '../constants';
import { GAME_MANAGEMENT_URL } from '../../../shared/config/api';
import { matchmakingApi } from '../api/matchmakingApi';
import { supabase } from '../../../shared/api/supabase';

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
    serverRestarting,
    claimWalkover,
    clearPostMatchOutcome,
  } = useMatchLobby(matchId || null);
  console.log('[MatchLobbyPage] matchState:', matchState);

  // Tournament info is now included in matchState (from matchmaking service)
  // No need for separate API call to tournament service (which was blocked by CORS)

  // Determine current player's ready status
  const currentPlayerReady = user && matchState
    ? (matchState.player1?.id === user.id ? matchState.match.participant1Ready : matchState.match.participant2Ready)
    : false;

  // Ready state management
  const {
    isReady: localReadyState,
    isLoading: isReadyLoading,
    toggleReady,
    setReady,
  } = useReadyState(
    matchId || '',
    user?.id || '',
    currentPlayerReady
  );

  // Sync local ready state with server state
  // IMPORTANT: Don't sync while isReadyLoading=true to preserve optimistic updates
  useEffect(() => {
    if (matchState && user && !isReadyLoading) {
      const serverReady = matchState.player1?.id === user.id
        ? matchState.match.participant1Ready
        : matchState.match.participant2Ready;

      setReady(serverReady);
    }
  }, [matchState, user, setReady, isReadyLoading]);

  // JWT token for game iframe - use ref to avoid re-render on token refresh
  const [jwtToken, setJwtToken] = useState<string | undefined>(undefined);

  useEffect(() => {
    const getToken = async () => {
      const { data: { session } } = await supabase.auth.getSession();
      setJwtToken(session?.access_token);
    };
    getToken();
  }, []);

  // Disconnect state management
  const {
    isDisconnected,
    disconnectedPlayerId,
    disconnectedAt: _disconnectedAt,
    remainingSeconds,
    setDisconnected,
    setReconnected,
  } = useDisconnect();

  // Sync disconnect state with match state
  useEffect(() => {
    if (matchState) {
      if (matchState.disconnectedPlayerId && matchState.disconnectedAt) {
        setDisconnected(matchState.disconnectedPlayerId, matchState.disconnectedAt);
      } else if (disconnectedPlayerId && !matchState.disconnectedPlayerId) {
        // Player reconnected
        setReconnected();
      }
    }
  }, [matchState, disconnectedPlayerId, setDisconnected, setReconnected]);

  // Check authorization - verify user is a match participant
  useEffect(() => {
    if (!user || !matchState) return;
    
    // Wait for player data to load before checking participation
    // This prevents race condition where we navigate away before player data arrives
    if (!matchState.player1?.id && !matchState.player2?.id) {
      return;
    }

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
      <LobbyLayout tournamentId={tournamentId} streamerId={matchState?.tournamentInfo?.host_id}>
        <div className="flex items-center justify-center min-h-screen">
          <LoadingSpinner size="lg" />
        </div>
      </LobbyLayout>
    );
  }

  // Error state
  if (error) {
    return (
      <LobbyLayout tournamentId={tournamentId} streamerId={matchState?.tournamentInfo?.host_id}>
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
      tournamentId={tournamentId}
      streamerId={matchState.tournamentInfo?.host_id}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* WebSocket Connection Status Banner */}
        {serverRestarting && (
          <div className="mb-6 rounded-lg border-2 border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4">
            <div className="flex items-start gap-3">
              <svg className="size-6 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-blue-800 dark:text-blue-200">
                  Serwer restartuje...
                </h3>
                <p className="text-sm text-blue-700 dark:text-blue-300 mt-1">
                  Serwer jest restartowany. Automatycznie połączymy Cię ponownie za chwilę.
                </p>
              </div>
            </div>
          </div>
        )}

        {connectionStatus === 'reconnecting' && !serverRestarting && (
          <div className="mb-6 rounded-lg border-2 border-yellow-200 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-900/20 p-4">
            <div className="flex items-start gap-3">
              <svg className="size-6 text-yellow-600 dark:text-yellow-400 flex-shrink-0 mt-0.5 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-yellow-800 dark:text-yellow-200">
                  Łączenie ponownie...
                </h3>
                <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                  Próbujemy nawiązać połączenie z serwerem. Proszę czekać...
                </p>
              </div>
            </div>
          </div>
        )}

        {(connectionStatus === 'disconnected' || connectionStatus === 'error') && !serverRestarting && (
          <div className="mb-6 rounded-lg border-2 border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-4">
            <div className="flex items-start gap-3">
              <svg className="size-6 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div className="flex-1">
                <h3 className="text-sm font-semibold text-red-800 dark:text-red-200">
                  Utracono połączenie
                </h3>
                <p className="text-sm text-red-700 dark:text-red-300 mt-1">
                  Nie można połączyć się z serwerem. Sprawdź swoje połączenie internetowe. Próbujemy połączyć ponownie...
                </p>
                <button
                  onClick={() => window.location.reload()}
                  className="mt-3 px-4 py-2 bg-red-600 hover:bg-red-700 text-white text-sm rounded-lg font-medium transition-colors"
                >
                  Odśwież stronę
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                Lobby Meczu
              </h1>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {matchState.tournamentInfo?.name || 'Turniej'} · Runda {matchState.roundNumber}
              </p>
            </div>

            {/* Connection status indicator */}
            <ConnectionIndicator status={connectionStatus} size="sm" />
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
                         matchState.status === 'loading' ? 'Ładowanie gry' :
                         matchState.status === 'started' ? 'W trakcie' :
                         matchState.status === 'completed' ? 'Zakończony' : matchState.status}
              </h2>
              <p className="mt-1 text-sm text-cyan-700 dark:text-cyan-300">
                {matchState.status === 'pending' && 'Czekamy na dołączenie obu graczy'}
                {matchState.status === 'ready' && 'Obaj gracze w lobby - przygotuj się!'}
                {matchState.status === 'loading' && 'Ładowanie gry... Poczekaj aż gra się załaduje.'}
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
          player1Ready={matchState.match.participant1Ready}
          player2Ready={matchState.match.participant2Ready}
        />

        {/* Ready Button is now shown as overlay inside GameFrame */}

        {/* Loading Status - Show when match is in loading state */}
        {matchState.status === 'loading' && (
          <div className="mt-8">
            <div className="rounded-lg border-2 border-blue-300 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-6">
              <div className="flex items-center justify-center gap-4">
                <div className="animate-spin rounded-full size-8 border-b-2 border-blue-600"></div>
                <div className="text-center">
                  <p className="text-lg font-semibold text-blue-900 dark:text-blue-200">
                    Ładowanie gry...
                  </p>
                <p className="mt-1 text-sm text-blue-700 dark:text-blue-300">
                      {(() => { 
                        let renderedText = '';
                        if (!matchState.match.participant1GameLoaded && !matchState.match.participant2GameLoaded) renderedText = 'Oba graczy ładują grę';
                        else if (matchState.match.participant1GameLoaded && !matchState.match.participant2GameLoaded) renderedText = 'Gracz 1 załadował grę, czekamy na Gracza 2';
                        else if (!matchState.match.participant1GameLoaded && matchState.match.participant2GameLoaded) renderedText = 'Gracz 2 załadował grę, czekamy na Gracza 1';
                        else if (matchState.match.participant1GameLoaded && matchState.match.participant2GameLoaded) renderedText = 'Oba graczy załadowali grę!';
                        return renderedText;
                      })()}
                </p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Match Start Countdown Overlay */}
        {matchState.matchStarting && matchState.countdownSeconds !== null && matchState.countdownSeconds > 0 && (
          <MatchStartCountdown 
            initialSeconds={matchState.countdownSeconds}
            onComplete={() => console.log('Countdown complete')}
          />
        )}

        {/* Walkover Button - Show after 2 minutes if opponent doesn't join */}
        {matchState.canClaimWalkover && !matchState.player2 && matchState.status === 'ready' && (
          <div className="mt-8">
            <div className="mb-4 rounded-lg border-2 border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-900/20 p-4">
              <div className="flex items-start gap-3">
                <div className="text-2xl">⏰</div>
                <div className="flex-1">
                  <p className="font-semibold text-amber-900 dark:text-amber-200 mb-1">
                    Przeciwnik nie stawił się
                  </p>
                  <p className="text-sm text-amber-700 dark:text-amber-300">
                    Minęły 2 minuty oczekiwania. Możesz zgłosić walkower i awansować do następnej rundy.
                  </p>
                </div>
              </div>
            </div>
            <WalkoverButton
              onClaim={claimWalkover}
              disabled={false}
            />
          </div>
        )}

        {/* Game Iframe - Show when both players present (pending, loading, or started) */}
        {matchState.player1 && matchState.player2 &&
         (matchState.status === 'pending' || matchState.status === 'loading' || matchState.status === 'started') && (
          <div className="mt-8">
            <GameFrame
              gameUrl={
                matchState.gameSnapshot?.slug
                  ? `${GAME_MANAGEMENT_URL}/v1/g/${matchState.gameSnapshot.slug}/bundle/`
                  : ''
              }
              matchId={matchState.match.id}
              token={jwtToken}
              onGameLoaded={async () => {
                console.log('[MatchLobbyPage] Game iframe loaded, marking as loading');
                if (user) {
                  try {
                    await matchmakingApi.markGameLoading(matchState.match.id, user.id);
                    console.log('[MatchLobbyPage] Server notified - status changed to loading');
                  } catch (err) {
                    console.error('[MatchLobbyPage] Failed to mark game loading:', err);
                  }
                }
              }}
              onGameError={(error) => {
                console.error('[MatchLobbyPage] Game error:', error);
              }}
              // Ready overlay props - show when both games are loaded and status is loading
              showReadyOverlay={matchState.status === 'loading' && matchState.match.participant1GameLoaded && matchState.match.participant2GameLoaded}
              isPlayerReady={localReadyState}
              isOpponentReady={
                user?.id === matchState.player1?.id
                  ? matchState.match.participant2Ready
                  : matchState.match.participant1Ready
              }
              onReadyClick={toggleReady}
              isReadyLoading={isReadyLoading}
            />
          </div>
        )}

        {/* Match Result - Show when match is completed */}
        {matchState.status === 'completed' && matchState.match.winnerId && (
          <div className="mt-8">
            <MatchResult
              winner={matchState.match.winnerId === matchState.player1?.id ? matchState.player1 : matchState.player2}
              loser={matchState.match.winnerId === matchState.player1?.id ? matchState.player2 : matchState.player1}
              currentUserId={user?.id}
              scoreWinner={matchState.match.winnerId === matchState.player1?.id ? matchState.match.scorePlayer1 : matchState.match.scorePlayer2}
              scoreLoser={matchState.match.winnerId === matchState.player1?.id ? matchState.match.scorePlayer2 : matchState.match.scorePlayer1}
              resultSource={matchState.match.resultSource || undefined}
              onContinue={() => {
                navigate(`/h/${matchState.tournamentInfo?.host_id}/tournaments/${tournamentId}`);
              }}
            />
          </div>
        )}

        {/* Info Footer */}
        <div className="mt-8 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 p-4">
          <div className="flex items-start gap-3">
            <div className="text-xl">ℹ️</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              <p className="font-medium">Informacje o lobby:</p>
              <ul className="mt-2 space-y-1 list-disc list-inside">
                <li>Gdy obaj gracze dołączą, będziecie mogli oznaczyć się jako gotowi</li>
                <li>Po oznaczeniu "Gotowy" przez obu graczy rozpocznie się ładowanie gry</li>
                <li>Gdy gra załaduje się u obu graczy, rozpocznie się odliczanie (3, 2, 1)</li>
                <li>W razie rozłączenia system automatycznie spróbuje ponownie połączyć</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      {/* Disconnect Overlay */}
      {isDisconnected && disconnectedPlayerId && (
        <DisconnectOverlay
          disconnectedPlayerName={
            disconnectedPlayerId === matchState.player1?.id
              ? matchState.player1?.displayName || 'Gracz 1'
              : matchState.player2?.displayName || 'Gracz 2'
          }
          isCurrentUserDisconnected={disconnectedPlayerId === user?.id}
          remainingSeconds={remainingSeconds}
          onTimeExpired={() => {
            console.log('[MatchLobbyPage] Disconnect timer expired');
            // The match_completed event from server will handle the walkover
          }}
        />
      )}

      {/* Post-Match Modal (Winner/Eliminated/Champion) */}
      <PostMatchModal
        outcome={matchState.postMatchOutcome}
        onClose={clearPostMatchOutcome}
      />
    </LobbyLayout>
  );
}

