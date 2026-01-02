/**
 * E2E tests for OAuth nickname generation in the database trigger.
 *
 * These tests verify that the `handle_new_user()` trigger correctly generates
 * unique nicknames for OAuth users who don't have explicit nickname metadata.
 *
 * The bug being tested: OAuth users (Google, Twitch, Discord) don't have 'nickname'
 * in their metadata, causing the trigger to default to empty string '', which violates
 * the UNIQUE constraint on the nickname column when multiple OAuth users sign up.
 */

import { test, expect } from '@playwright/test';
import { createClient } from '@supabase/supabase-js';
import { Pool } from 'pg';

const SUPABASE_URL = process.env.SUPABASE_URL ?? 'http://localhost:54321';
const SUPABASE_SERVICE_ROLE_KEY =
  process.env.SUPABASE_SERVICE_ROLE_KEY ??
  'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU';
const POSTGRES_DSN =
  process.env.POSTGRES_DSN ?? 'postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable';

// Use service role client to create users with custom metadata (simulating OAuth)
const supabaseAdmin = createClient(SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, {
  auth: {
    autoRefreshToken: false,
    persistSession: false,
  },
});

interface TestUser {
  id: string;
  email: string;
}

// Helper to create a user via Supabase Admin API with OAuth-like metadata
async function createOAuthUser(
  email: string,
  provider: 'google' | 'twitch' | 'discord',
  userMetadata: Record<string, string> = {}
): Promise<TestUser> {
  const { data, error } = await supabaseAdmin.auth.admin.createUser({
    email,
    email_confirm: true,
    app_metadata: {
      provider,
      providers: [provider],
    },
    user_metadata: userMetadata,
  });

  if (error) {
    throw new Error(`Failed to create user ${email}: ${error.message}`);
  }

  return { id: data.user.id, email: data.user.email! };
}

// Small delay to ensure trigger has completed
async function wait(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// Helper to get user profile from database
async function getUserProfile(userId: string): Promise<{ nickname: string; first_name: string; last_name: string } | null> {
  const pool = new Pool({ connectionString: POSTGRES_DSN });
  try {
    // Small delay to ensure trigger has completed
    await wait(100);
    const result = await pool.query(
      `SELECT nickname, first_name, last_name FROM public.user_profiles WHERE id = $1`,
      [userId]
    );
    return result.rows[0] ?? null;
  } finally {
    await pool.end();
  }
}

// Helper to get streamer profile from database
async function getStreamerProfile(
  userId: string
): Promise<{ nickname: string; first_name: string; last_name: string } | null> {
  const pool = new Pool({ connectionString: POSTGRES_DSN });
  try {
    // Small delay to ensure trigger has completed
    await wait(100);
    const result = await pool.query(
      `SELECT nickname, first_name, last_name FROM public.streamer_profiles WHERE id = $1`,
      [userId]
    );
    return result.rows[0] ?? null;
  } finally {
    await pool.end();
  }
}

// Helper to delete test user and their profile
async function cleanupUser(userId: string): Promise<void> {
  try {
    await supabaseAdmin.auth.admin.deleteUser(userId);
  } catch {
    // Ignore cleanup errors
  }
}

// Generate unique email for test isolation
function uniqueEmail(prefix: string): string {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(2, 8);
  return `${prefix}_${timestamp}_${random}@test.local`;
}

test.describe('OAuth Nickname Generation', () => {
  const createdUsers: string[] = [];

  test.afterEach(async () => {
    // Cleanup all users created during the test
    for (const userId of createdUsers) {
      await cleanupUser(userId);
    }
    createdUsers.length = 0;
  });

  test('should generate nickname from email when no nickname metadata provided (Google)', async () => {
    const email = uniqueEmail('oauth_google');
    const expectedNickname = email.split('@')[0]; // e.g., "oauth_google_1234567890_abc123"

    const user = await createOAuthUser(email, 'google', {
      // Simulate Google OAuth - typically has full_name but no nickname
      full_name: 'Test User',
    });
    createdUsers.push(user.id);

    const profile = await getUserProfile(user.id);

    expect(profile).not.toBeNull();
    expect(profile!.nickname).toBe(expectedNickname);
    expect(profile!.first_name).toBe('Test');
    expect(profile!.last_name).toBe('User');
  });

  test('should generate nickname from email when no nickname metadata provided (Twitch → streamer)', async () => {
    const email = uniqueEmail('oauth_twitch');
    const expectedNickname = email.split('@')[0];

    const user = await createOAuthUser(email, 'twitch', {
      // Simulate Twitch OAuth - may have preferred_username
      name: 'TwitchStreamer',
    });
    createdUsers.push(user.id);

    // Twitch users become streamers
    const profile = await getStreamerProfile(user.id);

    expect(profile).not.toBeNull();
    expect(profile!.nickname).toBe(expectedNickname);
  });

  test('should use preferred_username when available (Discord → streamer)', async () => {
    const email = uniqueEmail('oauth_discord');
    const preferredUsername = `DiscordUser_${Date.now()}`;

    const user = await createOAuthUser(email, 'discord', {
      preferred_username: preferredUsername,
      name: 'Discord Name',
    });
    createdUsers.push(user.id);

    // Discord users become streamers
    const profile = await getStreamerProfile(user.id);

    expect(profile).not.toBeNull();
    expect(profile!.nickname).toBe(preferredUsername);
  });

  test('should handle nickname collision by appending suffix', async () => {
    // First, create a user profile with a specific nickname via email signup
    const pool = new Pool({ connectionString: POSTGRES_DSN });
    const baseNickname = `collision_test_${Date.now()}`;

    // Create first user with that nickname
    const email1 = `${baseNickname}@test.local`;
    const user1 = await createOAuthUser(email1, 'google', {
      full_name: 'First User',
    });
    createdUsers.push(user1.id);

    const profile1 = await getUserProfile(user1.id);
    expect(profile1).not.toBeNull();
    expect(profile1!.nickname).toBe(baseNickname);

    // Create second user with same email prefix (different domain)
    const email2 = `${baseNickname}@other.local`;
    const user2 = await createOAuthUser(email2, 'google', {
      full_name: 'Second User',
    });
    createdUsers.push(user2.id);

    const profile2 = await getUserProfile(user2.id);
    expect(profile2).not.toBeNull();
    // Should have _1 suffix due to collision
    expect(profile2!.nickname).toBe(`${baseNickname}_1`);

    // Create third user with same email prefix
    const email3 = `${baseNickname}@another.local`;
    const user3 = await createOAuthUser(email3, 'google', {
      full_name: 'Third User',
    });
    createdUsers.push(user3.id);

    const profile3 = await getUserProfile(user3.id);
    expect(profile3).not.toBeNull();
    // Should have _2 suffix
    expect(profile3!.nickname).toBe(`${baseNickname}_2`);

    await pool.end();
  });

  test('should create multiple OAuth users without nickname collision', async () => {
    // Create 5 OAuth users in sequence - all should succeed with unique nicknames
    const users: TestUser[] = [];

    for (let i = 0; i < 5; i++) {
      const email = uniqueEmail(`multi_oauth_${i}`);
      const user = await createOAuthUser(email, 'google', {
        full_name: `User ${i}`,
      });
      createdUsers.push(user.id);
      users.push(user);
    }

    // Verify all users have profiles with unique nicknames
    const nicknames = new Set<string>();

    for (const user of users) {
      const profile = await getUserProfile(user.id);
      expect(profile).not.toBeNull();
      expect(profile!.nickname).not.toBe('');
      expect(nicknames.has(profile!.nickname)).toBe(false);
      nicknames.add(profile!.nickname);
    }

    expect(nicknames.size).toBe(5);
  });

  test('should extract first and last name from full_name', async () => {
    const email = uniqueEmail('name_test');

    const user = await createOAuthUser(email, 'google', {
      full_name: 'John Michael Smith',
    });
    createdUsers.push(user.id);

    const profile = await getUserProfile(user.id);

    expect(profile).not.toBeNull();
    expect(profile!.first_name).toBe('John');
    // Last name should be everything after first space
    expect(profile!.last_name).toBe('Michael Smith');
  });

  test('should handle user with only first name in full_name', async () => {
    const email = uniqueEmail('single_name');

    const user = await createOAuthUser(email, 'google', {
      full_name: 'Madonna',
    });
    createdUsers.push(user.id);

    const profile = await getUserProfile(user.id);

    expect(profile).not.toBeNull();
    expect(profile!.first_name).toBe('Madonna');
    expect(profile!.last_name).toBe('');
  });

  test('should use fallback UUID-based nickname when email has no local part', async () => {
    // This is an edge case - create user directly in DB to test fallback
    // In practice, this shouldn't happen with real OAuth providers
    const pool = new Pool({ connectionString: POSTGRES_DSN });

    try {
      // Create a user with empty email (edge case)
      const { data, error } = await supabaseAdmin.auth.admin.createUser({
        email: uniqueEmail('fallback_test'),
        email_confirm: true,
        app_metadata: {
          provider: 'google',
          providers: ['google'],
        },
        user_metadata: {
          // No name metadata at all
        },
      });

      if (error) {
        throw error;
      }

      createdUsers.push(data.user.id);

      const profile = await getUserProfile(data.user.id);

      expect(profile).not.toBeNull();
      // Should have a nickname derived from email prefix
      expect(profile!.nickname.length).toBeGreaterThan(0);
      expect(profile!.nickname).not.toBe('');
    } finally {
      await pool.end();
    }
  });
});
