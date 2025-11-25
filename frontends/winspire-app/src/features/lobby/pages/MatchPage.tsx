import { useParams } from 'react-router-dom';
import { AppLayout } from '../../../shared/components/layout/AppLayout';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';

export function MatchPage() {
  const { tournamentId, matchId } = useParams<{ tournamentId: string; matchId: string }>();

  return (
    <AppLayout>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-2xl font-bold mb-4">Match {matchId}</h1>
        <p className="text-gray-600 dark:text-gray-400">
          Tournament: {tournamentId}
        </p>
        {/* Match details will be implemented later */}
        <LoadingSpinner />
      </div>
    </AppLayout>
  );
}

