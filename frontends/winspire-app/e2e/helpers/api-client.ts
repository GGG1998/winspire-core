/**
 * API Client for E2E tests
 * 
 * Direct API calls to bypass UI for setup/teardown operations.
 * This dramatically speeds up tests by avoiding UI navigation.
 * 
 * NOTE: Uses Supabase client for auth to properly handle user metadata and profiles.
 */

import { APIRequestContext } from '@playwright/test';
import { createClient } from '@supabase/supabase-js';

// Environment variables with fallbacks
// Competition service runs on :8089 locally (make run)
// Matchmaking service runs on :8088 locally (make run)
const API_BASE_URL = process.env.API_BASE_URL ?? 'http://localhost:8089';
const SUPABASE_URL = process.env.SUPABASE_URL ?? 'http://localhost:54321';
const SUPABASE_ANON_KEY = process.env.SUPABASE_ANON_KEY ?? 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0';

declare const process: {
  env: {
    API_BASE_URL?: string;
    SUPABASE_URL?: string;
    SUPABASE_ANON_KEY?: string;
  };
};

// Create Supabase client for auth operations
const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

export interface TestUser {
  id: string;
  email: string;
  password: string;
  nickname: string;
  accessToken: string;
}

export interface TestTournament {
  id: string;
  name: string;
  creatorId: string;
}

/**
 * Create a user via Supabase SDK (proper way)
 * This ensures user metadata and profiles are created correctly
 */
export async function createUser(
  request: APIRequestContext,
  nickname: string,
  email: string,
  password: string
): Promise<TestUser> {
  try {
    // Use Supabase SDK for proper signup
    const { data: authData, error: authError } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: {
          user_type: 'user',
          app_id: 'user-frontend-v1', // Must match trigger validation
          first_name: 'Test',
          last_name: 'User',
          nickname: nickname,
        }
      }
    });

    if (authError || !authData.user || !authData.session) {
      console.error(`[API] Failed to create user ${nickname}:`, authError);
      throw new Error(`Failed to create user: ${authError?.message || 'Unknown error'}`);
    }

    // Profile is automatically created by database trigger (handle_new_user)
    // No need to create it manually

    return {
      id: authData.user.id,
      email,
      password,
      nickname,
      accessToken: authData.session.access_token,
    };
  } catch (error) {
    console.error(`[API] Exception creating user ${nickname}:`, error);
    throw error;
  }
}

/**
 * Create a streamer via Supabase SDK (proper way)
 */
export async function createStreamer(
  request: APIRequestContext,
  nickname: string,
  email: string,
  password: string,
  channelName: string
): Promise<TestUser> {
  try {
    // Use Supabase SDK for proper signup
    const { data: authData, error: authError } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: {
          user_type: 'streamer',
          app_id: 'streamer-frontend-v1', // Must match trigger validation
          first_name: 'Streamer',
          last_name: 'Test',
          nickname: nickname,
          channel_name: channelName,
        }
      }
    });

    if (authError || !authData.user || !authData.session) {
      console.error(`[API] Failed to create streamer ${nickname}:`, authError);
      throw new Error(`Failed to create streamer: ${authError?.message || 'Unknown error'}`);
    }

    // Profile is automatically created by database trigger (handle_new_user)
    // No need to create it manually

    return {
      id: authData.user.id,
      email,
      password,
      nickname,
      accessToken: authData.session.access_token,
    };
  } catch (error) {
    console.error(`[API] Exception creating streamer ${nickname}:`, error);
    throw error;
  }
}

/**
 * Create tournament via Competition API
 * Endpoint: POST /v1/:hostId/tournaments
 */
export async function createTournament(
  request: APIRequestContext,
  hostId: string,
  accessToken: string,
  name: string,
  startTime: Date
): Promise<TestTournament> {
  const response = await request.post(
    `${API_BASE_URL}/v1/${hostId}/tournaments`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        name,
        scheduledStartTimeAt: startTime.toISOString(),
        competitionParameters: {
          tournamentParameters: {
            teamSize: 1,
          }
        },
        minimumTeamCount: 4,
        game: {
          slug: 'test-game'
        }
      }
    }
  );

  if (!response.ok()) {
    const errorText = await response.text();
    console.error(`[API] Failed to create tournament:`, errorText);
    throw new Error(`Failed to create tournament: ${errorText}`);
  }

  const data = await response.json();
  
  return {
    id: data.tournamentId,
    name,
    creatorId: hostId,
  };
}

/**
 * Transition tournament status via API
 * Endpoint: PATCH /v1/:hostId/tournaments/:tournamentId
 */
export async function updateTournamentStatus(
  request: APIRequestContext,
  tournamentId: string,
  hostId: string,
  accessToken: string,
  status: 'scheduled' | 'registration_open' | 'started'
): Promise<void> {
  const response = await request.patch(
    `${API_BASE_URL}/v1/${hostId}/tournaments/${tournamentId}`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: { status }
    }
  );

  if (!response.ok()) {
    const errorText = await response.text();
    console.error(`[API] Failed to update tournament status:`, errorText);
    throw new Error(`Failed to update tournament status: ${errorText}`);
  }
}

/**
 * Register user for tournament via API
 * Endpoint: POST /v1/:hostId/tournaments/:tournamentId/register
 */
export async function registerForTournament(
  request: APIRequestContext,
  tournamentId: string,
  hostId: string,
  accessToken: string
): Promise<void> {
  const response = await request.post(
    `${API_BASE_URL}/v1/${hostId}/tournaments/${tournamentId}/register`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: {}
    }
  );

  if (!response.ok()) {
    const errorText = await response.text();
    console.error(`[API] Failed to register for tournament:`, errorText);
    throw new Error(`Failed to register for tournament: ${errorText}`);
  }
}

/**
 * Generate unique test data with high collision resistance
 */
export function generateTestData(prefix: string) {
  const timestamp = Date.now();
  // Add random string to prevent collisions when tests run in parallel
  const randomStr = Math.random().toString(36).substring(2, 8);
  const uniqueId = `${timestamp}_${randomStr}`;
  
  return {
    email: `${prefix}_${uniqueId}@test.local`,
    password: 'TestPassword123!',
    nickname: `${prefix}_${uniqueId}`,
    tournamentName: `Test Tournament ${uniqueId}`,
  };
}

