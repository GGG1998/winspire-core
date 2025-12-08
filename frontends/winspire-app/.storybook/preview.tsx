import React from 'react';
import type { Preview, Decorator } from "@storybook/react";
import { initialize, mswLoader } from 'msw-storybook-addon';
import { AuthProvider } from "../src/features/auth/context/AuthContext";
import { installMockWebSocket, getMockWebSocketServer } from './mocks/mockWebSocket';
import { supabase } from "../src/shared/api/supabase";
import handlers from './mocks/handlers';
import "../src/index.css";

// Mock Supabase auth BEFORE MSW initializes
if (typeof window !== 'undefined') {
  // Store original methods if not already stored
  if (!(window as any).__originalSupabaseAuth) {
    (window as any).__originalSupabaseAuth = {
      getSession: supabase.auth.getSession.bind(supabase.auth),
      getUser: supabase.auth.getUser.bind(supabase.auth),
      onAuthStateChange: supabase.auth.onAuthStateChange.bind(supabase.auth),
      from: supabase.from.bind(supabase),
    };
  }

  // Mock session data for stories
  const mockSession = {
    access_token: 'storybook-mock-token',
    token_type: 'bearer',
    expires_in: 3600,
    expires_at: Date.now() / 1000 + 3600,
    refresh_token: 'storybook-mock-refresh',
    user: {
      id: '123e4567-e89b-12d3-a456-426614174000',
      email: 'test@example.com',
      aud: 'authenticated',
      role: 'authenticated',
      app_metadata: {},
      user_metadata: {},
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
  };

  // Override Supabase auth methods
  supabase.auth.getSession = async () => ({
    data: { session: mockSession as any },
    error: null,
  }) as any;

  supabase.auth.getUser = async () => ({
    data: { user: mockSession.user as any },
    error: null,
  }) as any;

  // Mock onAuthStateChange
  supabase.auth.onAuthStateChange = ((callback: any) => {
    setTimeout(() => {
      callback('INITIAL_SESSION', mockSession);
    }, 0);

    return {
      data: {
        subscription: {
          unsubscribe: () => {},
        },
      },
    };
  }) as any;

  // Mock Supabase database queries for profile tables
  supabase.from = ((table: string) => {
    let eqValue: any = null;
    const mockQuery = {
      select: () => mockQuery,
      eq: (_column: string, value: any) => {
        eqValue = value;
        return mockQuery;
      },
      single: async () => {
        if (table === 'user_profiles' && eqValue === '123e4567-e89b-12d3-a456-426614174000') {
          return {
            data: {
              id: '123e4567-e89b-12d3-a456-426614174000',
              first_name: 'Pro',
              last_name: 'Gamer',
              nickname: 'ProGamer2024',
              city: 'Warsaw',
              country_id: 'PL',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
            error: null,
          };
        }
        return {
          data: null,
          error: { message: 'Not found' },
        };
      },
      insert: () => mockQuery,
      update: () => mockQuery,
      delete: () => mockQuery,
    };
    return mockQuery;
  }) as any;
}

// Initialize MSW
initialize({
  onUnhandledRequest: 'warn', // Warn about unhandled requests for debugging
  serviceWorker: {
    url: '/mockServiceWorker.js',
  },
});

console.log('[MSW] Initialization called');

// Fallback: Direct fetch mocking if MSW doesn't work
if (typeof window !== 'undefined') {
  const originalFetch = window.fetch;
  
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    
    // Check if this is a matchmaking API request
    if (url.includes('/v1/matchmaking/matches/')) {
      console.log('[Fetch Mock] ✅ Intercepted matchmaking request:', url);
      
      // Parse matchId from URL
      const matchIdMatch = url.match(/\/matches\/([^/?]+)/);
      const matchId = matchIdMatch ? matchIdMatch[1] : 'unknown';
      
      // Return mock data
      const mockResponse = {
        match: {
          id: matchId,
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
      
      await new Promise(resolve => setTimeout(resolve, 100)); // Simulate delay
      
      return new Response(JSON.stringify(mockResponse), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      });
    }
    
    // For all other requests, use original fetch (or MSW if it's working)
    return originalFetch(input, init);
  };
  
  console.log('[Fetch Mock] Direct fetch mocking installed as fallback');
}

// Install Mock WebSocket globally for all stories
if (typeof window !== 'undefined') {
  installMockWebSocket();
  
  // Reset WebSocket server between stories
  const server = getMockWebSocketServer();
  server.reset();
}

// Global decorator to wrap all stories with AuthProvider
const withAuthProvider: Decorator = (Story) => (
  <AuthProvider>
    <Story />
  </AuthProvider>
);

const preview: Preview = {
  decorators: [withAuthProvider],
  loaders: [mswLoader],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    msw: {
      handlers,
    },
  },
};

export default preview;
