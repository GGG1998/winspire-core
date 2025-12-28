import { useNavigate } from 'react-router-dom';
import type { Match, MatchPlayer } from '../types';

interface BracketMatchProps {
  match: Match;
  player1: MatchPlayer | null;
  player2: MatchPlayer | null;
  tournamentId: string;
  streamerId: string;
  isClickable?: boolean;
}

/**
 * BracketMatch Component
 * 
 * Displays a single match node in the bracket with:
 * - Player names and avatars
 * - Match status/winner highlighting
 * - Bye indicators
 * - Click navigation to match lobby
 */
export function BracketMatch({
  match,
  player1,
  player2,
  tournamentId,
  streamerId: _streamerId,
  isClickable = true
}: BracketMatchProps) {
  const navigate = useNavigate();

  // Determine if this is a bye match
  const isBye = !match.participant2Id;
  
  // Determine winner/loser
  const player1Won = match.winnerId === match.participant1Id;
  const player2Won = match.winnerId === match.participant2Id;
  
  const handleClick = () => {
    if (!isClickable || isBye) return;
    
    // Navigate to match lobby
    navigate(`/lobby/${tournamentId}/match/${match.id}`);
  };

  // Status badge color
  const getStatusColor = () => {
    switch (match.status) {
      case 'started':
        return 'bg-green-100 dark:bg-green-900/40 text-green-800 dark:text-green-200';
      case 'completed':
        return 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300';
      case 'ready':
        return 'bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-200';
      default:
        return 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300';
    }
  };

  return (
    <div
      onClick={handleClick}
      className={`
        relative group w-64 rounded-lg border-2 bg-white dark:bg-gray-800
        ${isBye ? 'border-gray-300 dark:border-gray-600' : 'border-gray-200 dark:border-gray-700'}
        ${isClickable && !isBye ? 'cursor-pointer hover:border-cyan-400 dark:hover:border-cyan-600 hover:shadow-lg transition-all' : ''}
      `}
    >
      {/* Match Number Badge */}
      <div className="absolute -top-3 left-3 px-2 py-0.5 bg-gray-100 dark:bg-gray-700 rounded-full text-xs font-medium text-gray-600 dark:text-gray-300">
        M{match.matchNumber}
      </div>

      {/* Status Badge */}
      {match.status === 'started' && (
        <div className="absolute -top-3 right-3 px-2 py-0.5 bg-green-500 rounded-full text-xs font-medium text-white animate-pulse">
          LIVE
        </div>
      )}

      <div className="p-3 space-y-2">
        {/* Player 1 */}
        <div className={`
          flex items-center gap-2 p-2 rounded-lg
          ${player1Won ? 'bg-green-50 dark:bg-green-900/20 ring-2 ring-green-400 dark:ring-green-600' : ''}
          ${!player1Won && match.status === 'completed' ? 'opacity-50' : ''}
        `}>
          {player1?.avatarUrl ? (
            <img 
              src={player1.avatarUrl} 
              alt={player1.displayName}
              className="size-8 rounded-full border border-gray-200 dark:border-gray-600"
            />
          ) : (
            <div className="size-8 rounded-full bg-gray-200 dark:bg-gray-600 flex items-center justify-center text-xs font-medium text-gray-600 dark:text-gray-300">
              {player1?.displayName?.slice(0, 2).toUpperCase() || '?'}
            </div>
          )}
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
              {player1?.displayName || 'TBD'}
            </p>
          </div>
          {player1Won && (
            <svg className="size-5 text-green-600 dark:text-green-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
          )}
          {match.scorePlayer1 !== null && (
            <span className="text-sm font-bold text-gray-900 dark:text-white">
              {match.scorePlayer1}
            </span>
          )}
        </div>

        {/* VS Divider */}
        <div className="flex items-center justify-center">
          <div className="text-xs font-medium text-gray-400 dark:text-gray-500">VS</div>
        </div>

        {/* Player 2 or BYE */}
        {isBye ? (
          <div className="flex items-center justify-center p-2 rounded-lg bg-yellow-50 dark:bg-yellow-900/20">
            <span className="text-sm font-medium text-yellow-700 dark:text-yellow-400">BYE</span>
          </div>
        ) : (
          <div className={`
            flex items-center gap-2 p-2 rounded-lg
            ${player2Won ? 'bg-green-50 dark:bg-green-900/20 ring-2 ring-green-400 dark:ring-green-600' : ''}
            ${!player2Won && match.status === 'completed' ? 'opacity-50' : ''}
          `}>
            {player2?.avatarUrl ? (
              <img 
                src={player2.avatarUrl} 
                alt={player2.displayName}
                className="size-8 rounded-full border border-gray-200 dark:border-gray-600"
              />
            ) : (
              <div className="size-8 rounded-full bg-gray-200 dark:bg-gray-600 flex items-center justify-center text-xs font-medium text-gray-600 dark:text-gray-300">
                {player2?.displayName?.slice(0, 2).toUpperCase() || '?'}
              </div>
            )}
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                {player2?.displayName || 'TBD'}
              </p>
            </div>
            {player2Won && (
              <svg className="size-5 text-green-600 dark:text-green-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
            )}
            {match.scorePlayer2 !== null && (
              <span className="text-sm font-bold text-gray-900 dark:text-white">
                {match.scorePlayer2}
              </span>
            )}
          </div>
        )}
      </div>

      {/* Status Footer */}
      <div className={`px-3 py-2 border-t border-gray-200 dark:border-gray-700 ${getStatusColor()} rounded-b-lg`}>
        <p className="text-xs font-medium text-center">
          {match.status === 'started' && 'W trakcie'}
          {match.status === 'completed' && 'Zakończony'}
          {match.status === 'ready' && 'Gotowy'}
          {match.status === 'paused' && 'Wstrzymany'}
          {match.status === 'disputed' && 'Sporny'}
          {match.status === 'cancelled' && 'Anulowany'}
        </p>
      </div>
    </div>
  );
}

