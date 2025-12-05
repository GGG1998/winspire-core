/**
 * TournamentPreLobbyPage Stories
 * Feature: 003-matchmaking-lobby-frontend
 */

import type { Meta, StoryObj } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';
import { ParticipantList } from '../components/ParticipantList';
import { ActivityFeed } from '../components/ActivityFeed';
import { GracePeriodIndicator } from '../components/GracePeriodIndicator';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ErrorMessage } from '../../../shared/components/common/ErrorMessage';
import { LobbyLayout } from '../layouts';
import { supabase } from '../../../shared/api/supabase'

import type { TournamentPreLobbyState, ConnectionState } from '../types';

// Mock user
const mockUser = {
  id: '223e4567-e89b-12d3-a456-426614174001',
  email: 'test@example.com',
  displayName: 'CyberNinja',
  avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=CyberNinja',
  profile: {
    id: 'streamer-456',
    first_name: 'Gaming',
    last_name: 'Streamer',
    nickname: 'ProStreamer',
    city: 'Warsaw',
    country_id: 'PL',
    created_at: new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString(),
    updated_at: new Date().toISOString(),
  }
};

// Mock Supabase auth to return mock user session
function setupMockAuth() {
    // Store original methods if not already stored
    if (!(window as any).__originalSupabaseAuth) {
      (window as any).__originalSupabaseAuth = {
        getSession: supabase.auth.getSession.bind(supabase.auth),
        getUser: supabase.auth.getUser.bind(supabase.auth),
        onAuthStateChange: supabase.auth.onAuthStateChange.bind(supabase.auth),
        from: supabase.from.bind(supabase),
      }
    }
    
    // Mock session data
    const mockSession = {
      access_token: 'mock-access-token',
      token_type: 'bearer',
      expires_in: 3600,
      expires_at: Date.now() / 1000 + 3600,
      refresh_token: 'mock-refresh-token',
      user: {
        id: mockUser.id,
        email: mockUser.email,
        aud: 'authenticated',
        role: 'authenticated',
        app_metadata: {},
        user_metadata: {},
        created_at: mockUser.profile.created_at,
        updated_at: mockUser.profile.updated_at,
      }
    }
    
    // Override Supabase auth methods
    supabase.auth.getSession = async () => ({
      data: { session: mockSession as any },
      error: null
    }) as any
    
    supabase.auth.getUser = async () => ({
      data: { user: mockSession.user as any },
      error: null
    }) as any
    
    // Mock onAuthStateChange to immediately fire INITIAL_SESSION event
    supabase.auth.onAuthStateChange = ((callback: any) => {
      // Trigger INITIAL_SESSION event asynchronously
      setTimeout(() => {
        callback('INITIAL_SESSION', mockSession)
      }, 0)
      
      // Return subscription object
      return {
        data: {
          subscription: {
            unsubscribe: () => {}
          }
        }
      }
    }) as any
    
    // Mock Supabase database queries for profile tables
    supabase.from = ((table: string) => {
      let eqValue: any = null
      const mockQuery = {
        select: () => mockQuery,
        eq: (_column: string, value: any) => {
          eqValue = value
          return mockQuery
        },
        single: async () => {
          if (table === 'streamer_profiles' && eqValue === mockUser.id) {
            return {
              data: mockUser.profile,
              error: null
            }
          }
          if (table === 'user_profiles') {
            return {
              data: null,
              error: { message: 'Not found', code: 'PGRST116' }
            }
          }
          return {
            data: null,
            error: { message: 'Not found' }
          }
        },
        insert: () => mockQuery,
        update: () => mockQuery,
        delete: () => mockQuery,
      }
      return mockQuery
    }) as any
    
    // Store original fetch if not already stored
    if (!(window as any).__originalFetch) {
      (window as any).__originalFetch = window.fetch
    }
    
    const originalFetch = (window as any).__originalFetch
    
    // Override fetch to mock Supabase API calls for user profile
    window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
      
      // Mock user profile endpoint
      if (url.includes('user_profiles') || url.includes('streamer_profiles')) {
        return new Response(
          JSON.stringify({
            data: [mockUser.profile],
            error: null
          }),
          { 
            status: 200, 
            headers: { 'Content-Type': 'application/json' } 
          }
        )
      }
      
      // For all other requests, use the original or already-mocked fetch
      return originalFetch(input, init)
    }
    
    return () => {
      // Restore original methods
      const original = (window as any).__originalSupabaseAuth
      if (original) {
        supabase.auth.getSession = original.getSession
        supabase.auth.getUser = original.getUser
        supabase.auth.onAuthStateChange = original.onAuthStateChange
        supabase.from = original.from
      }
      window.fetch = originalFetch
    }
  }

// Mock participants
const mockParticipants = [
  {
    id: '123e4567-e89b-12d3-a456-426614174000',
    displayName: 'ProGamer2024',
    avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ProGamer2024',
    joinedAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
  },
  {
    id: mockUser.id,
    displayName: mockUser.displayName,
    avatarUrl: mockUser.avatarUrl,
    joinedAt: new Date(Date.now() - 4 * 60 * 1000).toISOString(),
  },
  {
    id: '323e4567-e89b-12d3-a456-426614174002',
    displayName: 'ElitePlayer',
    avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ElitePlayer',
    joinedAt: new Date(Date.now() - 3 * 60 * 1000).toISOString(),
  },
];

// Create mock pre-lobby state
const createMockPreLobbyState = (overrides?: Partial<TournamentPreLobbyState>): TournamentPreLobbyState => ({
  tournamentId: 'test-tournament-id',
  tournamentName: 'Championship Finals 2024',
  creatorId: mockUser.profile.id, // streamer ID
  startTime: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
  status: 'waiting',
  participants: mockParticipants,
  participantCount: mockParticipants.length,
  minimumParticipants: 2,
  gracePeriodEndsAt: null,
  activityFeed: [
    {
      id: '1',
      type: 'participant_joined',
      message: 'ProGamer2024 dołączył do turnieju',
      timestamp: mockParticipants[0].joinedAt,
      participantName: 'ProGamer2024',
    },
    {
      id: '2',
      type: 'participant_joined',
      message: 'CyberNinja dołączył do turnieju',
      timestamp: mockParticipants[1].joinedAt,
      participantName: 'CyberNinja',
    },
    {
      id: '3',
      type: 'participant_joined',
      message: 'ElitePlayer dołączył do turnieju',
      timestamp: mockParticipants[2].joinedAt,
      participantName: 'ElitePlayer',
    },
  ],
  ...overrides,
});

// Storybook wrapper component that renders the UI without WebSocket
interface PreLobbyStoryWrapperProps {
  preLobbyState: TournamentPreLobbyState | null;
  isLoading?: boolean;
  error?: string | null;
  connectionStatus?: ConnectionState;
  isGracePeriodActive?: boolean;
  isParticipantCountUpdating?: boolean;
}

function PreLobbyStoryWrapper({
  preLobbyState,
  isLoading = false,
  error = null,
  connectionStatus = 'connected',
  isGracePeriodActive = false,
  isParticipantCountUpdating = false,
}: PreLobbyStoryWrapperProps) {
  // Loading state
  if (isLoading) {
    return (
      <LobbyLayout>
        <div className="flex items-center justify-center min-h-screen">
          <LoadingSpinner size="lg" />
        </div>
      </LobbyLayout>
    );
  }

  // Error state
  if (error) {
    return (
      <LobbyLayout>
        <div className="max-w-4xl px-4 py-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Nie udało się załadować poczekalni
            </h2>
            <ErrorMessage message={error} />
            <button
              onClick={() => console.log('Navigate back')}
              className="mt-4 px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700"
            >
              Wróć do turnieju
            </button>
          </div>
        </div>
      </LobbyLayout>
    );
  }

  // No state
  if (!preLobbyState) {
    return (
      <LobbyLayout>
        <div className="max-w-4xl px-4 py-8">
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
              Poczekalnia niedostępna
            </h2>
            <ErrorMessage message="Nie można załadować stanu poczekalni" />
            <button
              onClick={() => console.log('Navigate back')}
              className="mt-4 px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700"
            >
              Wróć do turnieju
            </button>
          </div>
        </div>
      </LobbyLayout>
    );
  }

  // Calculate time until start
  const timeUntilStart = (() => {
    const start = new Date(preLobbyState.startTime).getTime();
    const now = Date.now();
    const diff = start - now;
    const seconds = Math.max(0, Math.floor(diff / 1000));
    
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${secs}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`;
    } else {
      return `${secs}s`;
    }
  })();

  return (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
                {preLobbyState.tournamentName}
              </h1>
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                Poczekalnia turniejowa
              </p>
            </div>

            {/* Connection status indicator */}
            <div className="flex items-center gap-2">
              {connectionStatus === 'connected' && (
                <>
                  <div className="size-2 rounded-full bg-green-500 animate-pulse" />
                  <span className="text-sm text-gray-600 dark:text-gray-400">Połączony</span>
                </>
              )}
              {connectionStatus === 'connecting' && (
                <>
                  <LoadingSpinner size="sm" />
                  <span className="text-sm text-gray-600 dark:text-gray-400">Łączenie...</span>
                </>
              )}
              {connectionStatus === 'reconnecting' && (
                <>
                  <LoadingSpinner size="sm" />
                  <span className="text-sm text-amber-600 dark:text-amber-400">Ponowne łączenie...</span>
                </>
              )}
              {connectionStatus === 'disconnected' && (
                <>
                  <div className="size-2 rounded-full bg-red-500" />
                  <span className="text-sm text-red-600 dark:text-red-400">Rozłączony</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Grace Period Indicator */}
        {isGracePeriodActive && preLobbyState.gracePeriodEndsAt && (
          <div className="mb-6">
            <GracePeriodIndicator
              gracePeriodEndsAt={preLobbyState.gracePeriodEndsAt}
              participantCount={preLobbyState.participantCount}
              isUpdating={isParticipantCountUpdating}
            />
          </div>
        )}

        {/* Tournament Status Card */}
        {preLobbyState.status === 'waiting' && (
          <div className="mb-6 rounded-lg border-2 border-cyan-200 dark:border-cyan-800 bg-cyan-50 dark:bg-cyan-900/20 p-6">
            <div className="flex items-center gap-4">
              <div className="text-5xl">🏆</div>
              <div className="flex-1">
                <h2 className="text-lg font-semibold text-cyan-900 dark:text-cyan-100">
                  Oczekiwanie na start turnieju
                </h2>
                <p className="mt-1 text-sm text-cyan-700 dark:text-cyan-300">
                  Turniej rozpocznie się za: <span className="font-bold">{timeUntilStart}</span>
                </p>
                <p className="mt-2 text-xs text-cyan-600 dark:text-cyan-400">
                  Uczestnicy: {preLobbyState.participantCount} / {preLobbyState.minimumParticipants} (minimum)
                </p>
              </div>
            </div>
          </div>
        )}

        {preLobbyState.status === 'generating_bracket' && !isGracePeriodActive && (
          <div className="mb-6 rounded-lg border-2 border-purple-200 dark:border-purple-800 bg-purple-50 dark:bg-purple-900/20 p-6">
            <div className="flex items-center gap-4">
              <LoadingSpinner size="lg" />
              <div className="flex-1">
                <h2 className="text-lg font-semibold text-purple-900 dark:text-purple-100">
                  Generowanie drabinki...
                </h2>
                <p className="mt-1 text-sm text-purple-700 dark:text-purple-300">
                  Proszę czekać, mecze są przydzielane
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Main Content Grid */}
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {/* Participants */}
          <div>
            <ParticipantList
              participants={preLobbyState.participants}
              currentUserId={mockUser.id}
            />
          </div>

          {/* Activity Feed */}
          <div>
            <ActivityFeed items={preLobbyState.activityFeed} />
          </div>
        </div>

        {/* Footer Info */}
        <div className="mt-8 rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 p-4">
          <div className="flex items-start gap-3">
            <div className="text-xl">ℹ️</div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              <p className="font-medium">Jak to działa:</p>
              <ul className="mt-2 space-y-1 list-disc list-inside">
                <li>Czekaj tutaj aż gospodarz rozpocznie turniej</li>
                <li>Po rozpoczęciu masz 30 sekund (okres łaski) na dołączenie jeśli się spóźniłeś</li>
                <li>Zostaniesz automatycznie przekierowany do swojego meczu po wygenerowaniu drabinki</li>
                <li>Możesz obserwować innych graczy dołączających w czasie rzeczywistym</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </LobbyLayout>
  );
}

const meta = {
  title: 'Lobby/TournamentPreLobbyPage',
  component: PreLobbyStoryWrapper,
  decorators: [
    (Story) => {
      setupMockAuth();
      return (
        <MemoryRouter>
          <Story />
        </MemoryRouter>
      )
    },
  ],
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  args: {
    preLobbyState: createMockPreLobbyState(),
    isLoading: false,
    error: null,
    connectionStatus: 'connected' as const,
    isGracePeriodActive: false,
    isParticipantCountUpdating: false,
  },
} satisfies Meta<typeof PreLobbyStoryWrapper>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Waiting state
 * Tournament hasn't started yet, showing countdown
 */
export const WaitingForStart: Story = {
  args: {
    preLobbyState: createMockPreLobbyState(),
    connectionStatus: 'connected',
  },
};

/**
 * Grace period active
 * 30-second grace period after tournament start
 */
export const GracePeriodActive: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      status: 'grace_period',
      startTime: new Date(Date.now() - 30 * 1000).toISOString(),
      gracePeriodEndsAt: new Date(Date.now() + 15 * 1000).toISOString(),
      activityFeed: [
        {
          id: '1',
          type: 'tournament_starting',
          message: 'Turniej rozpoczyna się!',
          timestamp: new Date(Date.now() - 30 * 1000).toISOString(),
        },
        {
          id: '2',
          type: 'grace_period_started',
          message: 'Rozpoczęto okres łaski - spóźnieni gracze mogą jeszcze dołączyć (30s)',
          timestamp: new Date(Date.now() - 30 * 1000).toISOString(),
        },
      ],
    }),
    isGracePeriodActive: true,
    connectionStatus: 'connected',
  },
};

/**
 * Generating bracket
 * After grace period, bracket is being generated
 */
export const GeneratingBracket: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      status: 'generating_bracket',
      startTime: new Date(Date.now() - 60 * 1000).toISOString(),
      gracePeriodEndsAt: null,
      activityFeed: [
        {
          id: '1',
          type: 'grace_period_started',
          message: 'Rozpoczęto okres łaski',
          timestamp: new Date(Date.now() - 60 * 1000).toISOString(),
        },
      ],
    }),
    connectionStatus: 'connected',
  },
};

/**
 * Loading state
 * Initial page load
 */
export const Loading: Story = {
  args: {
    preLobbyState: null,
    isLoading: true,
    connectionStatus: 'connecting',
  },
};

/**
 * Error state
 * Failed to load pre-lobby data
 */
export const ErrorState: Story = {
  args: {
    preLobbyState: null,
    isLoading: false,
    error: 'Failed to load tournament pre-lobby',
    connectionStatus: 'error',
  },
};

/**
 * Single participant
 * Waiting for more players to join
 */
export const SingleParticipant: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      startTime: new Date(Date.now() + 10 * 60 * 1000).toISOString(),
      participants: [mockParticipants[0]],
      participantCount: 1,
      activityFeed: [
        {
          id: '1',
          type: 'participant_joined',
          message: 'ProGamer2024 dołączył do turnieju',
          timestamp: mockParticipants[0].joinedAt,
          participantName: 'ProGamer2024',
        },
      ],
    }),
    connectionStatus: 'connected',
  },
};

/**
 * Large participant list
 * Many players in the pre-lobby
 */
export const ManyParticipants: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      startTime: new Date(Date.now() + 3 * 60 * 1000).toISOString(),
      participants: Array.from({ length: 16 }, (_, i) => ({
        id: `participant-${i}`,
        displayName: `Player${i + 1}`,
        avatarUrl: `https://api.dicebear.com/7.x/avataaars/svg?seed=Player${i + 1}`,
        joinedAt: new Date(Date.now() - i * 30 * 1000).toISOString(),
      })),
      participantCount: 16,
      minimumParticipants: 8,
      activityFeed: Array.from({ length: 16 }, (_, i) => ({
        id: `activity-${i}`,
        type: 'participant_joined' as const,
        message: `Player${i + 1} dołączył do turnieju`,
        timestamp: new Date(Date.now() - i * 30 * 1000).toISOString(),
        participantName: `Player${i + 1}`,
      })),
    }),
    connectionStatus: 'connected',
  },
};

/**
 * Mobile view
 * Pre-lobby on smaller screens
 */
export const MobileView: Story = {
  args: {
    preLobbyState: createMockPreLobbyState(),
    connectionStatus: 'connected',
  },
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
};

/**
 * Disconnected state
 * WebSocket connection lost
 */
export const Disconnected: Story = {
  args: {
    preLobbyState: createMockPreLobbyState(),
    connectionStatus: 'reconnecting',
  },
};

/**
 * Grace period with participant count updating
 * Shows the pulsing animation when new players join
 */
export const GracePeriodWithUpdates: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      status: 'grace_period',
      startTime: new Date(Date.now() - 20 * 1000).toISOString(),
      gracePeriodEndsAt: new Date(Date.now() + 10 * 1000).toISOString(),
      participantCount: 9,
      activityFeed: [
        {
          id: '1',
          type: 'grace_period_started',
          message: 'Rozpoczęto okres łaski',
          timestamp: new Date(Date.now() - 20 * 1000).toISOString(),
        },
        {
          id: '2',
          type: 'participant_joined',
          message: 'LatePlayer dołączył podczas okresu łaski',
          timestamp: new Date(Date.now() - 5 * 1000).toISOString(),
          participantName: 'LatePlayer',
        },
      ],
    }),
    isGracePeriodActive: true,
    isParticipantCountUpdating: true,
    connectionStatus: 'connected',
  },
};

/**
 * Grace period low time
 * Last 5 seconds - red and pulsing
 */
export const GracePeriodLowTime: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      status: 'grace_period',
      startTime: new Date(Date.now() - 25 * 1000).toISOString(),
      gracePeriodEndsAt: new Date(Date.now() + 5 * 1000).toISOString(),
      activityFeed: [
        {
          id: '1',
          type: 'grace_period_started',
          message: 'Rozpoczęto okres łaski',
          timestamp: new Date(Date.now() - 25 * 1000).toISOString(),
        },
      ],
    }),
    isGracePeriodActive: true,
    connectionStatus: 'connected',
  },
};

/**
 * Connecting state
 * Initial WebSocket connection
 */
export const Connecting: Story = {
  args: {
    preLobbyState: createMockPreLobbyState(),
    connectionStatus: 'connecting',
  },
};

/**
 * Empty participants
 * No one has joined yet
 */
export const NoParticipants: Story = {
  args: {
    preLobbyState: createMockPreLobbyState({
      participants: [],
      participantCount: 0,
      activityFeed: [],
    }),
    connectionStatus: 'connected',
  },
};