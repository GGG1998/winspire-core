export interface Lobby {
  tournamentId: string;
  name: string;
  status: 'OPEN' | 'ACCEPTED' | 'PLAYING' | 'COMPLETE';
}

export interface Match {
  matchId: string;
  tournamentId: string;
  status: 'OPEN' | 'ACCEPTED' | 'PLAYING' | 'COMPLETE';
  lobbyInformation?: {
    maximumLobbyMinutes: number;
    maximumGoToGameMinutes: number;
    gameSessionTag?: string;
  };
  queueState?: {
    offerId: string;
    status: string;
    queueType?: string;
  };
}

export interface ChatMessage {
  id: string;
  userId: string;
  userName: string;
  message: string;
  timestamp: string;
}



