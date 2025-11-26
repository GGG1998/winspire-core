import { useParams } from 'react-router-dom';
import { useLobby } from '../hooks/useLobby';
import { LobbyView } from '../components/LobbyView';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { AppLayout } from '../../../shared/components/layout/AppLayout';

export function LobbyPage() {
  const { tournamentId } = useParams<{ tournamentId: string }>();
  const { lobby, isLoading, error } = useLobby(tournamentId || '');

  if (isLoading) {
    return (
      <AppLayout>
        <div className="container mx-auto px-4 py-8">
          <LoadingSpinner />
        </div>
      </AppLayout>
    );
  }

  if (error || !lobby) {
    return (
      <AppLayout>
        <div className="container mx-auto px-4 py-8">
          <ErrorMessage message={error || 'Lobby not found'} />
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="container mx-auto px-4 py-8">
        <LobbyView lobby={lobby} />
      </div>
    </AppLayout>
  );
}



