/**
 * MSW Handlers for Storybook
 * 
 * Mock HTTP API responses for matchmaking and other services
 */

import { http, HttpResponse, delay } from 'msw';

// Mock data types
export interface MockMatchData {
  match: any;
  participant1: any;
  participant2: any | null;
  tournament: any;
  roundNumber: number;
}

// Default mock match data (in API format - snake_case, matches GetMatchApiResponse)
export const defaultMockMatchData: MockMatchData = {
  match: {
    id: 'match-789',
    match_number: 1,
    next_match_id: null,
    participant1_id: '123e4567-e89b-12d3-a456-426614174000',
    participant2_id: '223e4567-e89b-12d3-a456-426614174001',
    status: 'ready',
    participant1_ready: false,
    participant2_ready: false,
    winner_id: null,
    score_player1: null,
    score_player2: null,
    started_at: null,
    completed_at: null,
  },
  participant1: {
    id: '123e4567-e89b-12d3-a456-426614174000',
    display_name: 'ProGamer2024',
    avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ProGamer2024',
  },
  participant2: {
    id: '223e4567-e89b-12d3-a456-426614174001',
    display_name: 'ElitePlayer99',
    avatar_url: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ElitePlayer99',
  },
  tournament: {
    id: 'tournament-123',
    name: 'Championship Finals 2024',
    creator_id: 'streamer-456',
  },
  round_number: 1,
};

// Storage for dynamic mock data (can be updated per story)
let currentMockData: MockMatchData = defaultMockMatchData;

export function setMockMatchData(data: Partial<MockMatchData>) {
  currentMockData = {
    ...currentMockData,
    ...data,
    match: {
      ...currentMockData.match,
      ...(data.match || {}),
    },
  };
}

export function resetMockMatchData() {
  currentMockData = defaultMockMatchData;
}

// API Handlers
export const handlers = [
  // Get match by ID - MATCHMAKING API (wildcard to match any origin)
  http.get('*/v1/matchmaking/matches/:matchId', async ({ params, request }) => {
    console.log('[MSW] ✅ Intercepted GET match request:', params.matchId);
    console.log('[MSW] Request URL:', request.url);
    console.log('[MSW] Returning mock data:', JSON.stringify(currentMockData, null, 2));
    
    await delay(100); // Simulate network delay
    
    // Note: In Storybook we don't check auth tokens - all requests are allowed
    const response = HttpResponse.json(currentMockData);
    console.log('[MSW] Response created:', response);
    return response;
  }),

  // Mark player ready - MATCHMAKING API
  http.post('*/v1/matchmaking/matches/:matchId/ready', async ({ request, params }) => {
    await delay(100);
    
    const body = await request.json() as any;
    const { playerId, ready } = body;

    // Update mock data
    if (playerId === currentMockData.participant1?.id) {
      currentMockData.match.participant1_ready = ready;
    } else if (playerId === currentMockData.participant2?.id) {
      currentMockData.match.participant2_ready = ready;
    }

    return HttpResponse.json({
      message: 'Ready status updated',
      match_id: params.matchId,
      player_id: playerId,
      ready,
    });
  }),

  // Claim walkover - MATCHMAKING API
  http.post('*/v1/matchmaking/matches/:matchId/claim-walkover', async ({ request, params }) => {
    await delay(100);
    
    const body = await request.json() as any;
    const { winnerId } = body;

    return HttpResponse.json({
      message: 'Walkover claimed',
      match_id: params.matchId,
      winner_id: winnerId,
      no_show_player: currentMockData.participant2?.id || 'unknown',
    });
  }),

  // Supabase auth endpoints - let them through (handled by mock in preview.tsx)
  http.post('*/auth/v1/token*', async () => {
    await delay(50);
    return HttpResponse.json({
      access_token: 'storybook-mock-token',
      token_type: 'bearer',
      expires_in: 3600,
      refresh_token: 'storybook-mock-refresh',
      user: {
        id: '123e4567-e89b-12d3-a456-426614174000',
        email: 'test@example.com',
        aud: 'authenticated',
        role: 'authenticated',
      },
    });
  }),

  http.get('*/auth/v1/user*', async () => {
    await delay(50);
    return HttpResponse.json({
      user: {
        id: '123e4567-e89b-12d3-a456-426614174000',
        email: 'test@example.com',
        aud: 'authenticated',
        role: 'authenticated',
      },
    });
  }),

  // Supabase user profile
  http.get('*/rest/v1/user_profiles*', async () => {
    await delay(50);
    return HttpResponse.json([
      {
        id: '123e4567-e89b-12d3-a456-426614174000',
        first_name: 'Pro',
        last_name: 'Gamer',
        nickname: 'ProGamer2024',
        city: 'Warsaw',
        country_id: 'PL',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      },
    ]);
  }),

  http.get('*/rest/v1/streamer_profiles*', async () => {
    await delay(50);
    return HttpResponse.json([]);
  }),
];

export default handlers;

