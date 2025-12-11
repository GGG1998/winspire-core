/**
 * useMatchLobby Hook Tests
 * 
 * Tests for the useMatchLobby hook, focusing on preventing infinite render loops
 * and proper state management.
 */

import { renderHook, waitFor } from '@testing-library/react';
import { useMatchLobby } from './useMatchLobby';
import * as matchmakingApiModule from '../api/matchmakingApi';
import * as useWebSocketModule from '../../../shared/hooks/useWebSocket';
import * as supabaseModule from '../../../shared/api/supabase';
import * as useAuthModule from '../../auth';

// Mock modules
jest.mock('../api/matchmakingApi');
jest.mock('../../../shared/hooks/useWebSocket');
jest.mock('../../../shared/api/supabase');
jest.mock('../../auth');

const mockMatchmakingApi = matchmakingApiModule.matchmakingApi as jest.Mocked<typeof matchmakingApiModule.matchmakingApi>;
const mockUseWebSocket = useWebSocketModule.useWebSocket as jest.MockedFunction<typeof useWebSocketModule.useWebSocket>;
const mockUseAuth = useAuthModule.useAuth as jest.MockedFunction<typeof useAuthModule.useAuth>;

describe('useMatchLobby', () => {
  const mockUser = {
    id: 'user-123',
    email: 'test@example.com',
  };

  const mockMatchData = {
    match: {
      id: 'match-789',
      roundId: 'round-456',
      matchNumber: 1,
      nextMatchId: null,
      participant1Id: 'user-123',
      participant2Id: 'user-456',
      status: 'ready' as const,
      participant1Ready: false,
      participant2Ready: false,
      winnerId: null,
      scorePlayer1: null,
      scorePlayer2: null,
      resultSource: null,
      disconnectedPlayerId: null,
      disconnectedAt: null,
      gameApiMatchId: null,
      createdAt: new Date(Date.now() - 1 * 60 * 1000).toISOString(), // 1 minute ago
      startedAt: null,
      completedAt: null,
      updatedAt: new Date().toISOString(),
    },
    participant1: {
      id: 'user-123',
      displayName: 'Player 1',
      avatarUrl: null,
    },
    participant2: {
      id: 'user-456',
      displayName: 'Player 2',
      avatarUrl: null,
    },
    tournament: {
      id: 'tournament-123',
      name: 'Test Tournament',
      creatorId: 'creator-789',
    },
    roundNumber: 1,
  };

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock useAuth
    (mockUseAuth as jest.Mock).mockReturnValue({
      user: mockUser,
    });

    // Mock supabase
    (supabaseModule.supabase.auth.getSession as jest.Mock).mockResolvedValue({
      data: {
        session: {
          access_token: 'mock-token',
        },
      },
    });

    // Mock useWebSocket
    (mockUseWebSocket as jest.Mock).mockReturnValue({
      status: 'connected',
      send: jest.fn(),
      close: jest.fn(),
    });

    // Mock matchmaking API
    mockMatchmakingApi.getMatch.mockResolvedValue(mockMatchData);
  });

  /**
   * Test: Prevent infinite loop when both players present
   * 
   * This test verifies that the walkover timer effect doesn't cause
   * infinite re-renders when both players are present in the lobby.
   * 
   * The bug occurred because the effect was calling setState on every
   * render, even when the walkover flags were already false.
   */
  it('should not cause infinite loop when both players are present', async () => {
    const matchId = 'match-789';

    const { result, rerender } = renderHook(() => useMatchLobby(matchId));

    // Wait for initial load
    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Verify match state loaded
    expect(result.current.matchState).not.toBeNull();
    expect(result.current.matchState?.player1).toBeDefined();
    expect(result.current.matchState?.player2).toBeDefined();

    // Verify walkover flags are false (both players present)
    expect(result.current.matchState?.canClaimWalkover).toBe(false);
    expect(result.current.matchState?.walkoverClaimableAt).toBeNull();

    // Track setState calls by checking if match state reference changes
    const initialMatchState = result.current.matchState;
    
    // Rerender multiple times to test for infinite loop
    for (let i = 0; i < 5; i++) {
      rerender();
      
      // In a working implementation, the matchState reference should NOT change
      // on subsequent renders when nothing has changed
      // This is a simplified check; in reality, React would throw if there was an infinite loop
      await waitFor(() => {
        expect(result.current.matchState).toBe(initialMatchState);
      });
    }

    // If we got here without throwing "Maximum update depth exceeded", the fix works!
    expect(true).toBe(true);
  });

  /**
   * Test: Walkover timer starts when opponent missing
   * 
   * Verifies that the walkover timer correctly sets the walkover flags
   * when an opponent is missing after 2 minutes.
   */
  it('should enable walkover claim after 2 minutes when opponent missing', async () => {
    const matchId = 'match-789';

    // Mock match data with missing opponent and old creation time
    const matchDataNoOpponent = {
      ...mockMatchData,
      participant2: null,
      match: {
        ...mockMatchData.match,
        participant2Id: null,
        status: 'ready' as const,
        createdAt: new Date(Date.now() - 3 * 60 * 1000).toISOString(), // 3 minutes ago
      },
    };

    mockMatchmakingApi.getMatch.mockResolvedValue(matchDataNoOpponent);

    const { result } = renderHook(() => useMatchLobby(matchId));

    // Wait for initial load
    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    // Verify walkover is claimable (more than 2 minutes elapsed)
    await waitFor(() => {
      expect(result.current.matchState?.canClaimWalkover).toBe(true);
      expect(result.current.matchState?.walkoverClaimableAt).toBeDefined();
    });
  });

  /**
   * Test: Walkover timer clears when opponent joins
   * 
   * Verifies that walkover flags are cleared when an opponent joins
   * after being initially missing.
   */
  it('should clear walkover flags when opponent joins', async () => {
    const matchId = 'match-789';

    // Start with missing opponent
    const matchDataNoOpponent = {
      ...mockMatchData,
      participant2: null,
      match: {
        ...mockMatchData.match,
        participant2Id: null,
        createdAt: new Date(Date.now() - 3 * 60 * 1000).toISOString(), // 3 minutes ago
      },
    };

    mockMatchmakingApi.getMatch.mockResolvedValue(matchDataNoOpponent);

    const { result, rerender } = renderHook(() => useMatchLobby(matchId));

    // Wait for initial load with missing opponent
    await waitFor(() => {
      expect(result.current.matchState?.canClaimWalkover).toBe(true);
    });

    // Now mock opponent joining
    mockMatchmakingApi.getMatch.mockResolvedValue(mockMatchData);

    // Force a re-render to simulate opponent joining via WebSocket update
    // In real usage, this would be triggered by a WebSocket message
    result.current.matchState!.player2 = mockMatchData.participant2;
    rerender();

    // Verify walkover flags are cleared
    await waitFor(() => {
      expect(result.current.matchState?.canClaimWalkover).toBe(false);
      expect(result.current.matchState?.walkoverClaimableAt).toBeNull();
    });
  });
});





