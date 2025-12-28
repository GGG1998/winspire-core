
/**
 * TournamentDetailPage Stories
 * Feature: 002-streamer-tournament-creation
 */

import type { Meta, StoryObj } from '@storybook/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { TournamentDetailPage } from './TournamentDetailPage'
import { TournamentsProvider } from '../contexts/TournamentsContext'
import { tournamentApi } from '../api/tournamentApi'
import type { Tournament } from '../types'
import type { User } from '../../auth/types'
import { supabase } from '../../../shared/api/supabase'

// ============================================================================
// Mock User Data
// ============================================================================

const mockStreamerUser: User = {
  id: 'streamer-456',
  email: 'streamer@example.com',
  profileType: 'streamer',
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
}

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
      id: mockStreamerUser.id,
      email: mockStreamerUser.email,
      aud: 'authenticated',
      role: 'authenticated',
      app_metadata: {},
      user_metadata: {},
      created_at: mockStreamerUser.profile.created_at,
      updated_at: mockStreamerUser.profile.updated_at,
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
        if (table === 'streamer_profiles' && eqValue === mockStreamerUser.id) {
          return {
            data: mockStreamerUser.profile,
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
          data: [mockStreamerUser.profile],
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

// ============================================================================
// Mock Data
// ============================================================================

const mockTournamentBase: Tournament = {
  id: 'tournament-123',
  name: 'Summer Gaming Championship 2024',
  status: 'registration_open',
  startTime: new Date(Date.now() + 2 * 60 * 60 * 1000), // 2 hours from now
  game: 'Packman',
  gameLogoUrl: 'https://images.unsplash.com/photo-1511512578047-dfb367046420?w=200&q=80',
  bannerUrl: 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=1920&q=80',
  creatorId: 'streamer-456',
  roomLink: '/tournament/tournament-123',
  isCompleted: false,
  createdAt: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000), // 7 days ago
  updatedAt: new Date(Date.now() - 1 * 60 * 60 * 1000), // 1 hour ago
  organizer: {
    id: 'org-001',
    name: 'Gaming Pro League',
    logoUrl: 'https://images.unsplash.com/photo-1614680376593-902f74cf0d41?w=100&q=80'
  },
  format: {
    type: 'single_elimination',
    teamSize: 1,
    maxSlots: 16,
    bestOf: 3
  },
  readyWindow: {
    startsAt: new Date(Date.now() + 1.5 * 60 * 60 * 1000), // 1.5 hours from now
    endsAt: new Date(Date.now() + 2 * 60 * 60 * 1000) // 2 hours from now
  },
  prize: {
    type: 'cash',
    description: '1000 PLN prize pool',
    value: 1000,
    currency: 'PLN'
  },
  participantCount: 4,
  lastActivity: new Date(Date.now() - 30 * 60 * 1000) // 30 minutes ago
}

const mockActiveTournament: Tournament = {
  ...mockTournamentBase,
  id: 'tournament-active',
  name: 'Active Tournament - Live Now',
  status: 'started',
  startTime: new Date(Date.now() - 30 * 60 * 1000), // Started 30 minutes ago
  lastActivity: new Date(Date.now() - 5 * 60 * 1000) // 5 minutes ago
}

const mockCompletedTournament: Tournament = {
  ...mockTournamentBase,
  id: 'tournament-completed',
  name: 'Completed Championship',
  status: 'completed',
  startTime: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000), // 3 days ago
  isCompleted: true,
  lastActivity: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000) // 2 days ago
}

// ============================================================================
// Mock API
// ============================================================================

let mockApiResponse: Tournament | Error | 'loading' = mockTournamentBase
let mockApiDelay = 0

function setupMockApi() {
  const originalGetTournament = tournamentApi.getTournament
  
  tournamentApi.getTournament = async (_tournamentId: string): Promise<Tournament> => {
    await new Promise(resolve => setTimeout(resolve, mockApiDelay))
    
    if (mockApiResponse === 'loading') {
      // Keep loading forever
      await new Promise(() => {})
    }
    
    if (mockApiResponse instanceof Error) {
      throw mockApiResponse
    }
    
    return mockApiResponse as Tournament
  }
  
  return () => {
    tournamentApi.getTournament = originalGetTournament
  }
}

// ============================================================================
// Story Configuration
// ============================================================================

const meta = {
  title: 'Host/Pages/TournamentDetailPage',
  component: TournamentDetailPage,
  decorators: [
    (Story, context) => {
      // Set up mock auth to provide mock user to global AuthProvider
      setupMockAuth()
      
      // Set up mock API based on story parameters
      if (context.parameters.mockTournament !== undefined) {
        mockApiResponse = context.parameters.mockTournament
      }
      if (context.parameters.mockApiDelay !== undefined) {
        mockApiDelay = context.parameters.mockApiDelay
      }
      
      setupMockApi()
      
      return (
        <MemoryRouter initialEntries={['/h/streamer-456/tournaments/tournament-123']}>
          <TournamentsProvider>
            <Routes>
              <Route path="/h/:streamerId/tournaments/:tournamentId" element={<Story />} />
            </Routes>
          </TournamentsProvider>
        </MemoryRouter>
      )
    }
  ],
  parameters: {
    layout: 'fullscreen',
    mockTournament: mockTournamentBase,
    mockApiDelay: 0,
  },
  tags: ['autodocs'],
} satisfies Meta<typeof TournamentDetailPage>

export default meta
type Story = StoryObj<typeof meta>

// ============================================================================
// Stories
// ============================================================================

/**
 * Loading state - Shows spinner while fetching tournament data
 */
export const Loading: Story = {
  parameters: {
    mockTournament: 'loading' as const,
    docs: {
      description: {
        story: 'Displays a loading spinner while the tournament data is being fetched from the API.',
      },
    },
  },
}

/**
 * Default view - Tournament detail with overview tab active
 */
export const LoadedOverview: Story = {
  parameters: {
    mockTournament: mockTournamentBase,
    docs: {
      description: {
        story: 'Default view showing tournament overview with banner, game logo, participants, and format details.',
      },
    },
  },
}

/**
 * Active tournament - Tournament is currently in progress
 */
export const ActiveTournament: Story = {
  parameters: {
    mockTournament: mockActiveTournament,
    docs: {
      description: {
        story: 'Tournament that is currently active/in progress, showing live status badge.',
      },
    },
  },
}

/**
 * Completed tournament - Tournament has ended
 */
export const CompletedTournament: Story = {
  parameters: {
    mockTournament: mockCompletedTournament,
    docs: {
      description: {
        story: 'Tournament that has been completed, showing completed status badge.',
      },
    },
  },
}

/**
 * Tournament without optional fields
 */
export const MinimalData: Story = {
  parameters: {
    mockTournament: {
      id: 'tournament-minimal',
      name: 'Minimal Tournament',
      status: 'draft',
      startTime: new Date(Date.now() + 2 * 60 * 60 * 1000),
      game: 'Packman',
      creatorId: 'streamer-456',
      roomLink: '/tournament/tournament-minimal',
      isCompleted: false,
      createdAt: new Date(),
      updatedAt: new Date(),
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament with minimal data - no banner, logo, organizer, format, or participants.',
      },
    },
  },
}

/**
 * Tournament with many participants
 */
export const ManyParticipants: Story = {
  parameters: {
    mockTournament: {
      ...mockTournamentBase,
      participants: Array.from({ length: 24 }, (_, i) => ({
        id: `player-${i}`,
        name: `Player${i + 1}`,
        avatarUrl: i % 3 === 0 ? `https://images.unsplash.com/photo-${1535713875002 + i}?w=100&q=80` : undefined,
        isReady: i % 2 === 0,
        registeredAt: new Date(Date.now() - (24 - i) * 60 * 60 * 1000)
      }))
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament with many participants to test the players panel display.',
      },
    },
  },
}

/**
 * Error state - Tournament not found (404)
 */
export const NotFound: Story = {
  parameters: {
    mockTournament: new Error('Tournament not found'),
    docs: {
      description: {
        story: 'Error state when the requested tournament does not exist.',
      },
    },
  },
}

/**
 * Error state - API/Network failure
 */
export const ApiError: Story = {
  parameters: {
    mockTournament: new Error('Failed to fetch tournament'),
    docs: {
      description: {
        story: 'Error state when there is a network or server error while fetching tournament data.',
      },
    },
  },
}

/**
 * Slow loading - Simulates slow network
 */
export const SlowLoading: Story = {
  parameters: {
    mockTournament: mockTournamentBase,
    mockApiDelay: 3000, // 3 second delay
    docs: {
      description: {
        story: 'Simulates a slow network connection with 3 second delay before data loads.',
      },
    },
  },
}

/**
 * Interactive - Test tab navigation
 * User can click on different tabs to see the tab content change
 */
export const InteractiveTabs: Story = {
  parameters: {
    mockTournament: mockTournamentBase,
    docs: {
      description: {
        story: 'Interactive story to test tab navigation. Click on different tabs (Przegląd, Drabinka, Mecze, Gracze, Wyniki) to see content change.',
      },
    },
  },
}

/**
 * User is owner - Logged in user is the tournament creator
 */
export const IsOwner: Story = {
  parameters: {
    mockTournament: {
      ...mockTournamentBase,
      id: 'tournament-owner',
      name: 'My Tournament - I am the Owner',
      creatorId: 'streamer-456', // Same as mockStreamerUser.id
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament where the logged in user (streamer-456) is the owner/creator. Should show edit/management controls.',
      },
    },
  },
}

/**
 * User is NOT owner - Logged in user is viewing someone else's tournament
 */
export const IsNotOwner: Story = {
  parameters: {
    mockTournament: {
      ...mockTournamentBase,
      id: 'tournament-not-owner',
      name: 'Other Streamer\'s Tournament',
      creatorId: 'other-streamer-789', // Different from mockStreamerUser.id
      organizer: {
        id: 'org-002',
        name: 'Other Gaming Organization',
        logoUrl: 'https://images.unsplash.com/photo-1614680376593-902f74cf0d41?w=100&q=80'
      }
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament created by another streamer (other-streamer-789). Should NOT show edit/management controls.',
      },
    },
  },
}

/**
 * Cancelled tournament - Tournament has been cancelled
 */
export const CancelledTournament: Story = {
  parameters: {
    mockTournament: {
      ...mockTournamentBase,
      id: 'tournament-cancelled',
      name: 'Cancelled Tournament',
      status: 'cancelled',
      isCompleted: false,
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament that has been cancelled. Shows cancelled badge with special indicator.',
      },
    },
  },
}

/**
 * Draft tournament - Tournament in draft status
 */
export const DraftTournament: Story = {
  parameters: {
    mockTournament: {
      ...mockTournamentBase,
      id: 'tournament-draft',
      name: 'Draft Tournament',
      status: 'draft',
      isCompleted: false,
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament in draft status, not yet published.',
      },
    },
  },
}

/**
 * Registration closed - Tournament with registration closed
 */
export const RegistrationClosed: Story = {
  parameters: {
    mockTournament: {
      ...mockTournamentBase,
      id: 'tournament-reg-closed',
      name: 'Registration Closed Tournament',
      status: 'registration_closed',
      isCompleted: false,
    } as Tournament,
    docs: {
      description: {
        story: 'Tournament with registration closed, waiting for start.',
      },
    },
  },
}

