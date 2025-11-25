import type { Lobby } from '../types';

interface LobbyInformationProps {
  lobby: Lobby;
}

export function LobbyInformation({ lobby }: LobbyInformationProps) {
  return (
    <div className="border rounded-lg p-4">
      <h3 className="text-lg font-semibold mb-2">{lobby.name}</h3>
      <div className="text-sm text-gray-600 dark:text-gray-400">
        Status: <span className="font-medium">{lobby.status}</span>
      </div>
    </div>
  );
}

