import { useNavigate } from 'react-router-dom';
import { Badge } from '../../../shared/components/ui/badge';
import type { MatchWithPlayers, MatchStatus } from '../types';

interface MatchCardProps {
  match: MatchWithPlayers;
  tournamentId: string;
  streamerId: string;
  currentUserId?: string;
  roundName: string;
  isClickable?: boolean;
}

/**
 * MatchCard Component
 * 
 * Displays a single match in list view with:
 * - Player names and avatars
 * - Score display
 * - Status badge (with pulsing "Live" for active)
 * - "Join Lobby" button for user's active matches
 * - Click navigation to match details
 */
export function MatchCard({
  match,
  tournamentId,
  streamerId: _streamerId,
  currentUserId,
  roundName,
  isClickable = true
}: MatchCardProps) {
  const navigate = useNavigate();

  // Check if current user is a participant
  const isUserParticipant = currentUserId && (
    match.participant1Id === currentUserId || 
    match.participant2Id === currentUserId
  );

  // Check if this is a bye match
  const isBye = !match.participant2Id;

  // Determine winner/loser
  const player1Won = match.winnerId === match.participant1Id;
  const player2Won = match.winnerId === match.participant2Id;

  const handleClick = () => {
    if (!isClickable || isBye) return;
    
    // Navigate to match lobby if user is participant, otherwise show read-only view
    if (isUserParticipant && (match.status === 'ready' || match.status === 'started')) {
      navigate(`/lobby/${tournamentId}/match/${match.id}`);
    }
  };

  const getStatusBadge = (status: MatchStatus) => {
    switch (status) {
      case 'started':
        return (
          <Badge color="emerald" className="animate-pulse">
            <span className="flex items-center gap-1">
              <span className="size-2 rounded-full bg-green-500 inline-block"></span>
              Na żywo
            </span>
          </Badge>
        );
      case 'completed':
        return <Badge color="zinc">Zakończony</Badge>;
      case 'ready':
        return <Badge color="cyan">Gotowy</Badge>;
      case 'paused':
        return <Badge color="amber">Wstrzymany</Badge>;
      case 'disputed':
        return <Badge color="red">Sporny</Badge>;
      case 'cancelled':
        return <Badge color="zinc">Anulowany</Badge>;
      default:
        return <Badge color="zinc">Oczekuje</Badge>;
    }
  };

  // Check if result was manually entered
  const isManualResult = match.resultSource === 'manual_host';
  
  // Check if there's an API error requiring manual entry
  const hasApiError = match.status === 'disputed'; // In real implementation, this would come from a specific flag

  return (
    <div
      onClick={handleClick}
      className={`
        rounded-xl border border-zinc-200 dark:border-zinc-700/50 
        bg-white dark:bg-zinc-800/50 p-4
        ${isClickable && !isBye ? 'cursor-pointer hover:border-cyan-400 dark:hover:border-cyan-600 hover:shadow-md transition-all' : ''}
      `}
    >
      {/* Header: Match number, round, status */}
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-sm text-zinc-500 dark:text-zinc-400">
            Mecz #{match.matchNumber}
          </span>
          <span className="text-sm text-zinc-400 dark:text-zinc-500">•</span>
          <span className="text-sm text-zinc-500 dark:text-zinc-400">
            {roundName}
          </span>
          {isManualResult && (
            <>
              <span className="text-sm text-zinc-400 dark:text-zinc-500">•</span>
              <Badge color="purple">
                <span className="flex items-center gap-1">
                  <svg className="size-3" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M13.586 3.586a2 2 0 112.828 2.828l-.793.793-2.828-2.828.793-.793zM11.379 5.793L3 14.172V17h2.828l8.38-8.379-2.83-2.828z" />
                  </svg>
                  Ręczny wynik
                </span>
              </Badge>
            </>
          )}
        </div>
        <div className="flex items-center gap-2">
          {getStatusBadge(match.status)}
        </div>
      </div>

      {/* API Error Alert */}
      {hasApiError && (
        <div className="mb-3 rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-3">
          <div className="flex items-start gap-2">
            <svg className="size-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
            <div className="flex-1">
              <p className="text-sm font-medium text-red-800 dark:text-red-200">
                Błąd API gry
              </p>
              <p className="text-xs text-red-700 dark:text-red-300 mt-1">
                Wymagane ręczne wprowadzenie wyniku przez gospodarza
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Players */}
      <div className="space-y-2 mb-3">
        {/* Player 1 */}
        <div className={`
          flex items-center justify-between p-2 rounded-lg
          ${player1Won ? 'bg-green-50 dark:bg-green-900/20' : ''}
          ${!player1Won && match.status === 'completed' ? 'opacity-60' : ''}
        `}>
          <div className="flex items-center gap-3 flex-1 min-w-0">
            {match.player1?.avatarUrl ? (
              <img 
                src={match.player1.avatarUrl} 
                alt={match.player1.displayName}
                className="size-8 rounded-full border border-zinc-200 dark:border-zinc-600"
              />
            ) : (
              <div className="size-8 rounded-full bg-zinc-100 dark:bg-zinc-700 flex items-center justify-center text-xs font-medium text-zinc-600 dark:text-zinc-300">
                {match.player1?.displayName?.slice(0, 2).toUpperCase() || '?'}
              </div>
            )}
            <span className="font-medium text-zinc-950 dark:text-white truncate">
              {match.player1?.displayName || 'TBD'}
              {currentUserId === match.participant1Id && (
                <span className="ml-2 text-xs text-cyan-600 dark:text-cyan-400">(Ty)</span>
              )}
            </span>
          </div>
          <div className="flex items-center gap-2">
            {player1Won && (
              <svg className="size-5 text-green-600 dark:text-green-400" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
            )}
            {match.scorePlayer1 !== null && (
              <span className="text-lg font-bold text-zinc-950 dark:text-white">
                {match.scorePlayer1}
              </span>
            )}
          </div>
        </div>

        {/* VS Divider or BYE */}
        {isBye ? (
          <div className="flex items-center justify-center py-1">
            <Badge color="amber">BYE</Badge>
          </div>
        ) : (
          <div className="flex items-center justify-center py-1">
            <span className="text-xs font-medium text-zinc-400 dark:text-zinc-500">VS</span>
          </div>
        )}

        {/* Player 2 */}
        {!isBye && (
          <div className={`
            flex items-center justify-between p-2 rounded-lg
            ${player2Won ? 'bg-green-50 dark:bg-green-900/20' : ''}
            ${!player2Won && match.status === 'completed' ? 'opacity-60' : ''}
          `}>
            <div className="flex items-center gap-3 flex-1 min-w-0">
              {match.player2?.avatarUrl ? (
                <img 
                  src={match.player2.avatarUrl} 
                  alt={match.player2.displayName}
                  className="size-8 rounded-full border border-zinc-200 dark:border-zinc-600"
                />
              ) : (
                <div className="size-8 rounded-full bg-zinc-100 dark:bg-zinc-700 flex items-center justify-center text-xs font-medium text-zinc-600 dark:text-zinc-300">
                  {match.player2?.displayName?.slice(0, 2).toUpperCase() || '?'}
                </div>
              )}
              <span className="font-medium text-zinc-950 dark:text-white truncate">
                {match.player2?.displayName || 'TBD'}
                {currentUserId === match.participant2Id && (
                  <span className="ml-2 text-xs text-cyan-600 dark:text-cyan-400">(Ty)</span>
                )}
              </span>
            </div>
            <div className="flex items-center gap-2">
              {player2Won && (
                <svg className="size-5 text-green-600 dark:text-green-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              )}
              {match.scorePlayer2 !== null && (
                <span className="text-lg font-bold text-zinc-950 dark:text-white">
                  {match.scorePlayer2}
                </span>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Join Lobby Button (for user's active matches) */}
      {isUserParticipant && (match.status === 'ready' || match.status === 'started') && (
        <div className="pt-2 border-t border-zinc-200 dark:border-zinc-700">
          <button
            onClick={(e) => {
              e.stopPropagation();
              navigate(`/lobby/${tournamentId}/match/${match.id}`);
            }}
            className="w-full px-4 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-lg font-medium transition-colors text-sm"
          >
            {match.status === 'started' ? 'Dołącz do gry' : 'Wejdź do lobby'}
          </button>
        </div>
      )}
    </div>
  );
}

