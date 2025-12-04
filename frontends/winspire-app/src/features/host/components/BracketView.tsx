/**
 * BracketView Component
 * Feature: 003-matchmaking-lobby-frontend
 * 
 * Displays interactive tournament bracket with:
 * - Real-time match data
 * - Winner highlighting
 * - Click navigation to matches
 * - Horizontal scroll for large brackets
 * - Responsive mobile layout
 */

import { useEffect, useState, useCallback } from 'react';
import { tournamentApi } from '../api/tournamentApi';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { ConnectionIndicator } from '../../../shared/components/ConnectionIndicator';
import { BracketMatch } from './BracketMatch';
import type { Tournament, Bracket, MatchPlayer } from '../types';

interface BracketViewProps {
  tournament: Tournament;
}

export function BracketView({ tournament }: BracketViewProps) {
  const [bracket, setBracket] = useState<Bracket | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Player data cache (from matches in bracket)
  const [playerCache, setPlayerCache] = useState<Map<string, MatchPlayer>>(new Map());

  // WebSocket connection for real-time bracket updates
  const wsUrl = `/matchmaking/v1/tournaments/${tournament.id}/bracket/ws`;
  
  const handleWebSocketMessage = useCallback((event: MessageEvent) => {
    try {
      const message = JSON.parse(event.data);
      console.log('[BracketView] WebSocket message:', message);

      // Handle bracket update events
      if (message.type === 'match_updated' || message.type === 'bracket_updated') {
        // Refresh bracket data
        console.log('[BracketView] Refreshing bracket after update');
        fetchBracket();
      }
    } catch (err) {
      console.error('[BracketView] Failed to parse WebSocket message:', err);
    }
  }, []);

  const { status: wsStatus } = useWebSocket({
    url: tournament.status === 'started' || tournament.status === 'completed' ? wsUrl : null,
    onMessage: handleWebSocketMessage,
    onOpen: () => {
      console.log('[BracketView] WebSocket connected');
    },
    onError: (error) => {
      console.error('[BracketView] WebSocket error:', error);
    },
  });

  // Fetch bracket data from API
  const fetchBracket = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      const bracketData = await tournamentApi.getBracket(tournament.id);
      setBracket(bracketData);

      // Build player cache from bracket data
      // In real implementation, this would come from API with player details
      // For now, we'll use participant IDs as placeholders
      const cache = new Map<string, MatchPlayer>();
      bracketData.rounds.forEach(round => {
        round.matches.forEach(match => {
          if (match.participant1Id && !cache.has(match.participant1Id)) {
            cache.set(match.participant1Id, {
              id: match.participant1Id,
              displayName: `Player ${match.participant1Id.slice(0, 6)}`,
              avatarUrl: null,
            });
          }
          if (match.participant2Id && !cache.has(match.participant2Id)) {
            cache.set(match.participant2Id, {
              id: match.participant2Id,
              displayName: `Player ${match.participant2Id.slice(0, 6)}`,
              avatarUrl: null,
            });
          }
        });
      });
      setPlayerCache(cache);
    } catch (err) {
      console.error('[BracketView] Failed to fetch bracket:', err);
      setError(err instanceof Error ? err.message : 'Failed to load bracket');
    } finally {
      setIsLoading(false);
    }
  }, [tournament.id]);

  useEffect(() => {
    // Only fetch if tournament is started or completed
    if (tournament.status === 'started' || tournament.status === 'completed') {
      fetchBracket();
    } else {
      setIsLoading(false);
      setError('Bracket not available yet. Start the tournament to generate the bracket.');
    }
  }, [tournament.status, fetchBracket]);

  // Get player from cache
  const getPlayer = (playerId: string | null): MatchPlayer | null => {
    if (!playerId) return null;
    return playerCache.get(playerId) || null;
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="py-8 max-w-md mx-auto">
        <ErrorMessage
          message={error}
          onRetry={() => window.location.reload()}
        />
      </div>
    );
  }

  // No bracket available
  if (!bracket) {
    return (
      <div className="py-8">
        <div className="text-center max-w-md mx-auto">
          <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
            <svg className="w-10 h-10 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={1.5}
                d="M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM14 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"
              />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-zinc-950 dark:text-white mb-2">
            Drabinka turniejowa
          </h3>
          <p className="text-zinc-500 dark:text-zinc-400">
            Drabinka zostanie wygenerowana po rozpoczęciu turnieju.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="py-8">
      {/* Header */}
      <div className="mb-6 px-4">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Drabinka turniejowa
          </h2>
          {/* Connection indicator for real-time updates */}
          <ConnectionIndicator status={wsStatus} size="sm" />
        </div>
        <div className="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
          <span>{bracket.totalRounds} rund</span>
          <span>•</span>
          <span>{bracket.totalMatches} meczów</span>
          {bracket.byesCount > 0 && (
            <>
              <span>•</span>
              <span>{bracket.byesCount} bye</span>
            </>
          )}
        </div>
      </div>

      {/* Bracket Grid - Horizontal Scroll */}
      <div className="overflow-x-auto pb-4">
        <div className="inline-flex gap-8 px-4 min-w-full">
          {bracket.rounds.map((round, roundIndex) => (
            <div 
              key={round.id} 
              className="flex-shrink-0"
              style={{
                // Vertical spacing increases for later rounds (bracket shape)
                paddingTop: `${roundIndex * 4}rem`,
              }}
            >
              {/* Round Header */}
              <div className="mb-4 sticky top-0 bg-white dark:bg-gray-900 z-10 pb-2">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {round.roundName}
                </h3>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  {round.matchesCount} {round.matchesCount === 1 ? 'mecz' : 'meczów'}
                </p>
              </div>

              {/* Matches in Round */}
              <div 
                className="space-y-8"
                style={{
                  // Vertical gap increases for later rounds
                  gap: `${Math.pow(2, roundIndex) * 2}rem`,
                }}
              >
                {round.matches.map((match) => (
                  <BracketMatch
                    key={match.id}
                    match={match}
                    player1={getPlayer(match.participant1Id)}
                    player2={getPlayer(match.participant2Id)}
                    tournamentId={tournament.id}
                    streamerId={tournament.creatorId}
                    isClickable={true}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Mobile Info */}
      <div className="mt-8 px-4">
        <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/20 p-4">
          <div className="flex items-start gap-3">
            <div className="text-xl">💡</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              <p className="font-medium mb-1">Porady:</p>
              <ul className="space-y-1 list-disc list-inside">
                <li>Kliknij na mecz, aby zobaczyć szczegóły</li>
                <li>Przewiń w prawo, aby zobaczyć kolejne rundy</li>
                <li>Zielona ramka = zwycięzca</li>
                <li>LIVE = mecz w trakcie</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
