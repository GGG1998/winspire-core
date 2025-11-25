import { MatchmakingQueue } from './MatchmakingQueue';
import { LobbyChat } from './LobbyChat';
import { LobbyInformation } from './LobbyInformation';
import type { Lobby } from '../types';

interface LobbyViewProps {
  lobby: Lobby;
}

export function LobbyView({ lobby }: LobbyViewProps) {
  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div className="lg:col-span-2 space-y-4">
        <LobbyInformation lobby={lobby} />
        <MatchmakingQueue tournamentId={lobby.tournamentId} />
      </div>
      <div className="lg:col-span-1">
        <LobbyChat tournamentId={lobby.tournamentId} />
      </div>
    </div>
  );
}

