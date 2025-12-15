import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../auth';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { ParticipantList } from '../components/ParticipantList';
import { ActivityFeed } from '../components/ActivityFeed';
import { GracePeriodIndicator } from '../components/GracePeriodIndicator';
import { ByeWaitingState } from '../components/ByeWaitingState';
import { useTournamentPreLobby } from '../hooks/useTournamentPreLobby';
import { LobbyLayout } from '../layouts';

/**
 * Tournament Pre-Lobby Page
 * 
 * Waiting room where players gather before tournament starts
 * Features:
 * - Real-time participant list
 * - Activity feed (joins/leaves)
 * - Countdown to tournament start
 * - Grace period indicator (30s after start)
 * - Automatic redirect to match lobby when assigned
 * - Late arrival handling with toast notifications
 */
export function TournamentPreLobbyPage() {
  const { tournamentId } = useParams<{ tournamentId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const {
    preLobbyState,
    isLoading,
    error,
    connectionStatus,
    isGracePeriodActive,
    isParticipantCountUpdating,
    hasBye,
    byeInfo,
  } = useTournamentPreLobby(tournamentId || null);

  const [matchAssignmentNotification] = useState<{
    opponentName: string;
    roundName: string;
  } | null>(null);

  // Note: Auth check removed - backend validates registration on WebSocket connection
  // If user is not registered, backend will reject WS connection and send error message
  // This prevents race condition where REST API returns empty participants list
  // before user's WebSocket connection is established

  // Loading state
  if (isLoading) {
    return (
      <LobbyLayout tournamentId={tournamentId} streamerId={preLobbyState?.creatorId}>
        <div className="flex items-center justify-center min-h-screen">
          <LoadingSpinner size="lg" />
        </div>
      </LobbyLayout>
    );
  }

  // Error state
  if (error) {
    return (
      <LobbyLayout tournamentId={tournamentId} streamerId={preLobbyState?.creatorId}>
        <div className="max-w-4xl px-4 py-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">
              Nie udało się załadować poczekalni
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
  if (!preLobbyState) {
    return (
      <LobbyLayout tournamentId={tournamentId}>
        <div className="max-w-4xl px-4 py-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">
              Poczekalnia niedostępna
            </h2>
            <ErrorMessage message="Nie można załadować stanu poczekalni" />
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

  // Show ByeWaitingState if player has BYE
  if (hasBye && byeInfo) {
    return (
      <LobbyLayout tournamentId={preLobbyState.tournamentId} streamerId={preLobbyState.creatorId}>
        <ByeWaitingState
          playerName={user?.profile.nickname || `${user?.profile.first_name} ${user?.profile.last_name}` || 'Gracz'}
          playerAvatarUrl={user?.profile.nickname ? `https://api.dicebear.com/7.x/avataaars/svg?seed=${user.profile.nickname}` : null}
          nextMatchSlot={byeInfo.nextMatchSlot}
          roundName={byeInfo.roundName}
          onRefresh={() => window.location.reload()}
        />
      </LobbyLayout>
    );
  }

  const timeUntilStart = (() => {
    const start = new Date(preLobbyState.startTime).getTime();
    const now = Date.now();
    const diff = start - now;
    const seconds = Math.max(0, Math.floor(diff / 1000));
    
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${secs}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`;
    } else {
      return `${secs}s`;
    }
  })();

  return (
    <LobbyLayout tournamentId={preLobbyState.tournamentId} streamerId={preLobbyState.creatorId}>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                {preLobbyState.tournamentName}
              </h1>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                Poczekalnia turniejowa
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

        {/* Grace Period Indicator */}
        {isGracePeriodActive && preLobbyState.gracePeriodEndsAt && (
          <div className="mb-6">
            <GracePeriodIndicator
              gracePeriodEndsAt={preLobbyState.gracePeriodEndsAt}
              participantCount={preLobbyState.participantCount}
              isUpdating={isParticipantCountUpdating}
            />
          </div>
        )}

        {/* Tournament Status Card */}
        {preLobbyState.status === 'waiting' && (
          <div className="mb-6 rounded-lg border-2 border-cyan-200 dark:border-cyan-800 bg-cyan-50 dark:bg-cyan-900/20 p-6">
            <div className="flex items-center gap-4">
              <div className="text-5xl">🏆</div>
              <div className="flex-1">
                <h2 className="text-lg font-semibold text-cyan-900 dark:text-cyan-100">
                  Oczekiwanie na start turnieju
                </h2>
                <p className="mt-1 text-sm text-cyan-700 dark:text-cyan-300">
                  Turniej rozpocznie się za: <span className="font-bold">{timeUntilStart}</span>
                </p>
                <p className="mt-2 text-xs text-cyan-600 dark:text-cyan-400">
                  Uczestnicy: {preLobbyState.participantCount} / {preLobbyState.minimumParticipants} (minimum)
                </p>
              </div>
            </div>
          </div>
        )}

        {preLobbyState.status === 'generating_bracket' && !isGracePeriodActive && (
          <div className="mb-6 rounded-lg border-2 border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-900/20 p-6">
            <div className="flex items-center gap-4">
              <LoadingSpinner size="lg" />
              <div className="flex-1">
                <h2 className="text-lg font-semibold text-purple-900 dark:text-purple-100">
                  Generowanie drabinki...
                </h2>
                <p className="mt-1 text-sm text-purple-700 dark:text-purple-300">
                  Proszę czekać, mecze są przydzielane
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Match Assignment Notification */}
        {matchAssignmentNotification && (
          <div className="mb-6 rounded-lg border-2 border-green-300 dark:border-green-800 bg-green-50 dark:bg-green-900/20 p-6 animate-in fade-in slide-in-from-top-4">
            <div className="flex items-center gap-4">
              <div className="text-5xl">🎮</div>
              <div className="flex-1">
                <h2 className="text-lg font-semibold text-green-900 dark:text-green-100">
                  Mecz przydzielony!
                </h2>
                <p className="mt-1 text-sm text-green-700 dark:text-green-300">
                  <span className="font-semibold">{matchAssignmentNotification.roundName}</span> vs{' '}
                  <span className="font-semibold">{matchAssignmentNotification.opponentName}</span>
                </p>
                <p className="mt-2 text-xs text-green-600 dark:text-green-400">
                  Przekierowanie do lobby meczu za 2 sekundy...
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {/* Participants */}
          <div>
            <ParticipantList
              participants={preLobbyState.participants}
              currentUserId={user?.id}
            />
          </div>

          {/* Activity Feed */}
          <div>
            <ActivityFeed items={preLobbyState.activityFeed} />
          </div>
        </div>

        {/* Footer Info */}
        <div className="mt-8 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 p-4">
          <div className="flex items-start gap-3">
            <div className="text-xl">ℹ️</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              <p className="font-medium">Jak to działa:</p>
              <ul className="mt-2 space-y-1 list-disc list-inside">
                <li>Czekaj tutaj aż gospodarz rozpocznie turniej</li>
                <li>Po rozpoczęciu masz 30 sekund (okres łaski) na dołączenie jeśli się spóźniłeś</li>
                <li>Zostaniesz automatycznie przekierowany do swojego meczu po wygenerowaniu drabinki</li>
                <li>Możesz obserwować innych graczy dołączających w czasie rzeczywistym</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </LobbyLayout>
  );
}

