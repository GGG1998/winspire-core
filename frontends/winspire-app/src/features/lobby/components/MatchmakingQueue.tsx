import { useEffect, useState } from 'react';
import { useSSE } from '../hooks/useSSE';
import { useAuth } from '../../../features/auth';
import { QueueStatus } from './QueueStatus';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { EmptyState } from '../../../shared/components/common/EmptyState';
import type { Match } from '../types';

interface MatchmakingQueueProps {
  tournamentId: string;
}

export function MatchmakingQueue({ tournamentId }: MatchmakingQueueProps) {
  const { user } = useAuth();
  const [matches, setMatches] = useState<Match[]>([]);
  // Note: This endpoint requires hostId from user context
  const endpoint = user?.id 
    ? `/v1/hosts/${user.id}/tournaments/${tournamentId}/matches`
    : null;
  const { data, isLoading } = useSSE<Match[]>(endpoint || '');

  useEffect(() => {
    if (data) {
      setMatches(data);
    }
  }, [data]);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (matches.length === 0) {
    return <EmptyState title="No matches available" />;
  }

  return (
    <div className="space-y-2">
      <h3 className="text-lg font-semibold">Matchmaking Queue</h3>
      {matches.map((match) => (
        <div key={match.id} className="border rounded-lg p-4">
          <div className="flex items-center justify-between">
            <span>Match {match.id.slice(0, 8)}</span>
            <QueueStatus status={match.status} />
          </div>
        </div>
      ))}
    </div>
  );
}

