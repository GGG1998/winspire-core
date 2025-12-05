/**
 * MatchesView Component
 * Feature: 003-matchmaking-lobby-frontend
 * 
 * Displays list of all matches in the tournament grouped by round
 * Features:
 * - Real API integration
 * - Round grouping (Round 1, Semifinals, Finals)
 * - Status badges with pulsing "Live" indicator
 * - Click navigation to match details
 * - "Join Lobby" button for user's active matches
 */

import { useEffect, useState, useCallback } from 'react';
import { tournamentApi } from '../api/tournamentApi';
import { useAuth } from '../../auth';
import { useWebSocket } from '../../../shared/hooks/useWebSocket';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { ConnectionIndicator } from '../../../shared/components/ConnectionIndicator';
import { MatchCard } from './MatchCard';
import { ManualResultForm } from './ManualResultForm';
import type { Tournament, MatchWithPlayers } from '../types';

interface MatchesViewProps {
  tournament: Tournament;
}

// Helper to get round name based on position
function getRoundName(roundNumber: number, totalRounds: number): string {
  if (roundNumber === totalRounds) return 'Finał';
  if (roundNumber === totalRounds - 1) return 'Półfinały';
  if (roundNumber === totalRounds - 2) return 'Ćwierćfinały';
  return `Runda ${roundNumber}`;
}

export function MatchesView({ tournament }: MatchesViewProps) {
  const { user } = useAuth();
  const [matches, setMatches] = useState<MatchWithPlayers[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showManualFormForMatch, setShowManualFormForMatch] = useState<string | null>(null);

  // Check if current user is the tournament host/creator
  const isHost = user?.id === tournament.creatorId;

  // Fetch matches data from API
  const fetchMatches = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      
      const matchesData = await tournamentApi.getMatches(tournament.id);
      setMatches(matchesData);
    } catch (err) {
      console.error('[MatchesView] Failed to fetch matches:', err);
      setError(err instanceof Error ? err.message : 'Failed to load matches');
    } finally {
      setIsLoading(false);
    }
  }, [tournament.id]);

  // WebSocket connection for real-time match updates
  const wsUrl = `/v1/matchmaking/tournaments/${tournament.id}/matches/ws`;
  
  const handleWebSocketMessage = useCallback((event: MessageEvent) => {
    try {
      const message = JSON.parse(event.data);
      console.log('[MatchesView] WebSocket message:', message);

      // Handle match update events
      if (message.type === 'match_updated' || message.type === 'match_status_changed') {
        // Refresh matches data
        console.log('[MatchesView] Refreshing matches after update');
        fetchMatches();
      }
    } catch (err) {
      console.error('[MatchesView] Failed to parse WebSocket message:', err);
    }
  }, [fetchMatches]);

  const { status: wsStatus } = useWebSocket({
    url: tournament.status === 'started' || tournament.status === 'completed' ? wsUrl : null,
    onMessage: handleWebSocketMessage,
    onOpen: () => {
      console.log('[MatchesView] WebSocket connected');
    },
    onError: (error) => {
      console.error('[MatchesView] WebSocket error:', error);
    },
  });

  useEffect(() => {
    // Only fetch if tournament is started or completed
    if (tournament.status === 'started' || tournament.status === 'completed') {
      fetchMatches();
    } else {
      setIsLoading(false);
      setError('Matches not available yet. Start the tournament to generate matches.');
    }
  }, [tournament.status, fetchMatches]);

  // Group matches by round
  const matchesByRound: Record<number, MatchWithPlayers[]> = matches.reduce((acc, match) => {
    const roundNum = getRoundNumberFromMatch(match);
    if (!acc[roundNum]) acc[roundNum] = [];
    acc[roundNum].push(match);
    return acc;
  }, {} as Record<number, MatchWithPlayers[]>);

  // Calculate total rounds (assumes single elimination: log2 of participants)
  const totalRounds = matches.length > 0 
    ? Math.max(...matches.map(m => getRoundNumberFromMatch(m)))
    : 0;

  // Helper to extract round number from match (assuming roundId contains round info)
  function getRoundNumberFromMatch(_match: MatchWithPlayers): number {
    // In real implementation, this would come from API
    // For now, derive from match structure
    // This is a placeholder - adjust based on actual API response
    return 1; // TODO: Extract from match.roundId or add roundNumber to API response
  }

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

  // No matches available
  if (matches.length === 0) {
    return (
      <div className="py-12 text-center">
        <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
          <svg className="w-8 h-8 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-zinc-950 dark:text-white mb-2">Brak meczów</h3>
        <p className="text-zinc-500 dark:text-zinc-400">Mecze pojawią się po rozpoczęciu turnieju.</p>
      </div>
    );
  }

  return (
    <div className="py-6 space-y-8">
      {/* Header with match count */}
      <div className="mb-4">
        <div className="flex items-center justify-between mb-1">
          <h2 className="text-2xl font-bold text-zinc-950 dark:text-white">
            Wszystkie mecze
          </h2>
          {/* Connection indicator for real-time updates */}
          <ConnectionIndicator status={wsStatus} size="sm" />
        </div>
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          {matches.length} {matches.length === 1 ? 'mecz' : 'meczów'} w {totalRounds} {totalRounds === 1 ? 'rundzie' : 'rundach'}
        </p>
      </div>

      {/* Matches grouped by round */}
      {Object.entries(matchesByRound)
        .sort(([a], [b]) => parseInt(a) - parseInt(b))
        .map(([roundNum, roundMatches]) => (
          <div key={roundNum}>
            {/* Round header */}
            <h3 className="text-lg font-semibold text-zinc-950 dark:text-white mb-4">
              {getRoundName(parseInt(roundNum), totalRounds)}
            </h3>
            
            {/* Matches list */}
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {roundMatches.map(match => {
                // Check if this match needs manual result entry (disputed + host is viewing)
                const needsManualEntry = match.status === 'disputed' && isHost;
                const isShowingForm = showManualFormForMatch === match.id;

                return (
                  <div key={match.id} className="space-y-3">
                    <MatchCard
                      match={match}
                      tournamentId={tournament.id}
                      streamerId={tournament.creatorId}
                      currentUserId={user?.id}
                      roundName={getRoundName(parseInt(roundNum), totalRounds)}
                      isClickable={!needsManualEntry}
                    />
                    
                    {/* Manual Result Form for disputed matches (host only) */}
                    {needsManualEntry && !isShowingForm && (
                      <button
                        onClick={() => setShowManualFormForMatch(match.id)}
                        className="w-full px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded-lg font-medium text-sm transition-colors"
                      >
                        Wprowadź wynik ręcznie
                      </button>
                    )}
                    
                    {isShowingForm && (
                      <ManualResultForm
                        matchId={match.id}
                        player1={match.player1}
                        player2={match.player2}
                        onSuccess={() => {
                          setShowManualFormForMatch(null);
                          fetchMatches(); // Refresh matches after successful submission
                        }}
                        onCancel={() => setShowManualFormForMatch(null)}
                      />
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        ))}
    </div>
  );
}
