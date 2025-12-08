/**
 * MatchLobbyPage Stories
 * 
 * Tests the REAL MatchLobbyPage component with mocked dependencies:
 * - MSW for HTTP API calls
 * - Mock WebSocket for real-time updates
 * - Supabase auth mocked globally
 * 
 * This approach tests the actual component logic, not just UI.
 */

import type { Meta, StoryObj } from '@storybook/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { http, HttpResponse, delay } from 'msw';
import { MatchLobbyPage } from './MatchLobbyPage';
import { getMockWebSocketServer } from '../../../../.storybook/mocks/mockWebSocket';
import type { Match, PlayerInfo } from '../types';

// ============================================================================
// Mock Data
// ============================================================================

const mockUser: PlayerInfo = {
  id: '123e4567-e89b-12d3-a456-426614174000',
  displayName: 'ProGamer2024',
  avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ProGamer2024',
};

const mockOpponent: PlayerInfo = {
  id: '223e4567-e89b-12d3-a456-426614174001',
  displayName: 'ElitePlayer99',
  avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ElitePlayer99',
};

const mockTournament = {
  id: 'tournament-123',
  name: 'Championship Finals 2024',
  creatorId: 'streamer-456',
};

// Base match object
const createMockMatch = (overrides?: Partial<Match>): Match => ({
  id: 'match-789',
  roundId: 'round-456',
  matchNumber: 1,
  nextMatchId: null,
  participant1Id: mockUser.id,
  participant2Id: mockOpponent.id,
  status: 'ready',
  participant1Ready: false,
  participant2Ready: false,
  winnerId: null,
  scorePlayer1: null,
  scorePlayer2: null,
  resultSource: null,
  disconnectedPlayerId: null,
  disconnectedAt: null,
  gameApiMatchId: null,
  createdAt: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
  startedAt: null,
  completedAt: null,
  updatedAt: new Date().toISOString(),
  ...overrides,
});

// Create full match response
const createMockMatchResponse = (matchOverrides?: Partial<Match>, options?: { noOpponent?: boolean }) => ({
  match: createMockMatch(matchOverrides),
  participant1: mockUser,
  participant2: options?.noOpponent ? null : mockOpponent,
  tournament: mockTournament,
  roundNumber: 1,
});

// ============================================================================
// Storybook Meta
// ============================================================================

const meta = {
  title: 'Lobby/MatchLobbyPage (Real Component)',
  component: MatchLobbyPage,
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/h/streamer-456/tournaments/tournament-123/matches/match-789']}>
        <Routes>
          <Route path="/h/:streamerId/tournaments/:tournamentId/matches/:matchId" element={<Story />} />
        </Routes>
      </MemoryRouter>
    ),
  ],
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof MatchLobbyPage>;

export default meta;
type Story = StoryObj<typeof meta>;

// ============================================================================
// Loading & Error States
// ============================================================================

/**
 * Loading state - Shows spinner while fetching initial match data
 */
export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay('infinite'); // Never resolves
        }),
      ],
    },
  },
};

/**
 * Error state - API returns error
 */
export const ErrorState: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            { error: 'Match not found' },
            { status: 404 }
          );
        }),
      ],
    },
  },
};

// ============================================================================
// Pre-Game States
// ============================================================================

/**
 * Waiting for opponent - Only current player has joined
 */
export const WaitingForOpponent: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse(
              {
                status: 'pending',
                participant2Id: null,
              },
              { noOpponent: true }
            )
          );
        }),
      ],
    },
  },
  play: async () => {
    // Simulate WebSocket connection
    const server = getMockWebSocketServer();
    server.broadcastMessage({
      type: 'lobby_state',
      payload: createMockMatchResponse(
        {
          status: 'pending',
          participant2Id: null,
        },
        { noOpponent: true }
      ),
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Opponent joined - Both players present, can mark ready
 */
export const OpponentJoined: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(createMockMatchResponse());
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    // Initial lobby state
    server.broadcastMessage({
      type: 'lobby_state',
      payload: createMockMatchResponse(),
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Current player ready - Waiting for opponent
 */
export const CurrentPlayerReady: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              participant1Ready: true,
              participant2Ready: false,
            })
          );
        }),
        http.post('*/v1/matchmaking/matches/:matchId/ready', async () => {
          await delay(100);
          return HttpResponse.json({
            message: 'Ready status updated',
            ready: true,
          });
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    server.broadcastMessage({
      type: 'lobby_state',
      payload: createMockMatchResponse({
        participant1Ready: true,
        participant2Ready: false,
      }),
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Both players ready - Countdown will start soon
 */
export const BothPlayersReady: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              participant1Ready: true,
              participant2Ready: true,
            })
          );
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    server.broadcastMessage({
      type: 'lobby_state',
      payload: createMockMatchResponse({
        participant1Ready: true,
        participant2Ready: true,
      }),
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Countdown active - 3, 2, 1... starting!
 */
export const CountdownActive: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              status: 'started',
              participant1Ready: true,
              participant2Ready: true,
            })
          );
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    // Initial state
    server.broadcastMessage({
      type: 'lobby_state',
      payload: createMockMatchResponse({
        participant1Ready: true,
        participant2Ready: true,
      }),
      timestamp: new Date().toISOString(),
    });

    // Countdown starting
    setTimeout(() => {
      server.broadcastMessage({
        type: 'match_starting',
        payload: { countdownSeconds: 3 },
        timestamp: new Date().toISOString(),
      });
    }, 500);
  },
};

// ============================================================================
// Game States
// ============================================================================

/**
 * Game in progress - Playing the game
 */
export const GameInProgress: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              status: 'started',
              participant1Ready: true,
              participant2Ready: true,
              gameApiMatchId: 'session-123',
              startedAt: new Date().toISOString(),
            })
          );
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    server.broadcastMessage({
      type: 'match_started',
      payload: {
        gameUrl: 'https://game.example.com/play/session-123',
        gameSessionId: 'session-123',
      },
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Match completed - Current player WON
 */
export const MatchCompletedWin: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              status: 'completed',
              winnerId: mockUser.id,
              scorePlayer1: 100,
              scorePlayer2: 75,
              resultSource: 'game_api',
              completedAt: new Date().toISOString(),
            })
          );
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    server.broadcastMessage({
      type: 'match_completed',
      payload: {
        winnerId: mockUser.id,
        scorePlayer1: 100,
        scorePlayer2: 75,
        resultSource: 'game_api',
      },
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Match completed - Current player LOST
 */
export const MatchCompletedLoss: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              status: 'completed',
              winnerId: mockOpponent.id,
              scorePlayer1: 75,
              scorePlayer2: 100,
              resultSource: 'game_api',
              completedAt: new Date().toISOString(),
            })
          );
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    server.broadcastMessage({
      type: 'match_completed',
      payload: {
        winnerId: mockOpponent.id,
        scorePlayer1: 75,
        scorePlayer2: 100,
        resultSource: 'game_api',
      },
      timestamp: new Date().toISOString(),
    });
  },
};

/**
 * Match completed by walkover
 */
export const MatchCompletedWalkover: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              status: 'completed',
              winnerId: mockUser.id,
              scorePlayer1: 1,
              scorePlayer2: 0,
              resultSource: 'walkover',
              completedAt: new Date().toISOString(),
            })
          );
        }),
      ],
    },
  },
};

// ============================================================================
// Special States
// ============================================================================

/**
 * Walkover available - Opponent didn't show up
 */
export const WalkoverAvailable: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse(
              {
                status: 'ready',
                participant2Id: null,
                createdAt: new Date(Date.now() - 3 * 60 * 1000).toISOString(), // 3 minutes ago
              },
              { noOpponent: true }
            )
          );
        }),
        http.post('*/v1/matchmaking/matches/:matchId/claim-walkover', async () => {
          await delay(100);
          return HttpResponse.json({
            message: 'Walkover claimed',
            winner_id: mockUser.id,
          });
        }),
      ],
    },
  },
};

/**
 * Opponent disconnected - Reconnect timer running
 */
export const OpponentDisconnected: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(
            createMockMatchResponse({
              status: 'started',
              gameApiMatchId: 'session-123',
              disconnectedPlayerId: mockOpponent.id,
              disconnectedAt: new Date(Date.now() - 5000).toISOString(),
            })
          );
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    // Game started
    server.broadcastMessage({
      type: 'match_started',
      payload: {
        gameUrl: 'https://game.example.com/play/session-123',
        gameSessionId: 'session-123',
      },
      timestamp: new Date().toISOString(),
    });

    // Opponent disconnected
    setTimeout(() => {
      server.broadcastMessage({
        type: 'player_disconnected',
        payload: {
          playerId: mockOpponent.id,
          disconnectedAt: new Date(Date.now() - 5000).toISOString(),
          reconnectDeadline: new Date(Date.now() + 25000).toISOString(),
        },
        timestamp: new Date().toISOString(),
      });
    }, 500);
  },
};

/**
 * Connection lost - WebSocket disconnected
 */
export const ConnectionLost: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(createMockMatchResponse());
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    // Don't send any WebSocket messages - simulates disconnection
    server.setAutoConnect(false);
  },
};

/**
 * Interactive - Click ready button to see state change
 */
export const Interactive: Story = {
  parameters: {
    msw: {
      handlers: [
        http.get('*/v1/matchmaking/matches/:matchId', async () => {
          await delay(100);
          return HttpResponse.json(createMockMatchResponse());
        }),
        http.post('*/v1/matchmaking/matches/:matchId/ready', async ({ request }) => {
          await delay(200);
          const body = await request.json() as any;
          
          // Simulate WebSocket update
          const server = getMockWebSocketServer();
          setTimeout(() => {
            server.broadcastMessage({
              type: 'ready_updated',
              payload: {
                playerId: body.playerId,
                ready: body.ready,
              },
              timestamp: new Date().toISOString(),
            });
          }, 100);
          
          return HttpResponse.json({
            message: 'Ready status updated',
            ready: body.ready,
          });
        }),
      ],
    },
  },
  play: async () => {
    const server = getMockWebSocketServer();
    
    server.broadcastMessage({
      type: 'lobby_state',
      payload: createMockMatchResponse(),
      timestamp: new Date().toISOString(),
    });
  },
};

