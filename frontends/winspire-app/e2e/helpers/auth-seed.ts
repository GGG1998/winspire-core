import { createClient, type Session } from '@supabase/supabase-js';
import { Pool } from 'pg';

type AccountType = 'user' | 'streamer';

export interface SeededAccount {
  id: string;
  email: string;
  password: string;
  nickname: string;
  type: AccountType;
  accessToken: string;
  refreshToken: string;
  session: Session;
}

const SUPABASE_URL = process.env.SUPABASE_URL ?? 'http://localhost:54321';
const SUPABASE_ANON_KEY =
  process.env.SUPABASE_ANON_KEY ??
  'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0';
const POSTGRES_DSN = process.env.POSTGRES_DSN ?? 'postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable';

const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

// Clean up ONLY orphaned profiles (profiles without matching auth.users)
async function cleanupOrphanedData(nickname: string) {
  const pool = new Pool({ connectionString: POSTGRES_DSN });
  try {
    // Delete profiles that don't have matching auth.users (orphaned data)
    await pool.query(`
      DELETE FROM public.streamer_profiles sp
      WHERE sp.nickname = $1 
      AND NOT EXISTS (SELECT 1 FROM auth.users u WHERE u.id = sp.id)
    `, [nickname]);
    await pool.query(`
      DELETE FROM public.user_profiles up
      WHERE up.nickname = $1 
      AND NOT EXISTS (SELECT 1 FROM auth.users u WHERE u.id = up.id)
    `, [nickname]);
  } catch {
    // Tables might not exist - ignore
  } finally {
    await pool.end();
  }
}

interface AccountSpec {
  email: string;
  password: string;
  nickname: string;
  firstName: string;
  lastName: string;
  userType: AccountType;
  appId: 'user-frontend-v1' | 'streamer-frontend-v1';
  channelName?: string;
}

function getStreamerSpec(num: number): AccountSpec {
  return {
    email: `streamer${num}@test.local`,
    password: 'TestPassword123!',
    nickname: `streamer${num}`,
    firstName: 'Streamer',
    lastName: `Test${num}`,
    userType: 'streamer',
    appId: 'streamer-frontend-v1',
    channelName: `channel_streamer_${num}`,
  };
}

function getUserSpec(num: number): AccountSpec {
  const padded = num.toString().padStart(2, '0');
  return {
    email: `user${padded}@test.local`,
    password: 'TestPassword123!',
    nickname: `user${padded}`,
    firstName: 'User',
    lastName: `Test${padded}`,
    userType: 'user',
    appId: 'user-frontend-v1',
  };
}

// Cache for already authenticated accounts
const accountCache = new Map<string, SeededAccount>();

async function getOrCreateAccount(spec: AccountSpec): Promise<SeededAccount> {
  // Check cache first
  const cached = accountCache.get(spec.email);
  if (cached) return cached;

  // Try sign in first
  const { data: signInData, error: signInError } = await supabase.auth.signInWithPassword({
    email: spec.email,
    password: spec.password,
  });

  if (!signInError && signInData.session) {
    const account = toSeededAccount(signInData.session, spec);
    accountCache.set(spec.email, account);
    return account;
  }

  // Clean up any orphaned profile data before creating user
  await cleanupOrphanedData(spec.nickname);

  // User doesn't exist - create with signUp
  const { data: signUpData, error: signUpError } = await supabase.auth.signUp({
    email: spec.email,
    password: spec.password,
    options: {
      data: {
        user_type: spec.userType,
        app_id: spec.appId,
        first_name: spec.firstName,
        last_name: spec.lastName,
        nickname: spec.nickname,
        ...(spec.channelName ? { channel_name: spec.channelName } : {}),
      },
    },
  });

  if (signUpError) {
    throw new Error(`Auth failed for ${spec.email}: ${signUpError.message}`);
  }

  // Get session (signUp may return it directly or we need to sign in)
  const session = signUpData.session ?? (await supabase.auth.signInWithPassword({
    email: spec.email,
    password: spec.password,
  })).data.session;

  if (!session) {
    throw new Error(`No session for ${spec.email}`);
  }

  const account = toSeededAccount(session, spec);
  accountCache.set(spec.email, account);
  return account;
}

function toSeededAccount(session: Session, spec: AccountSpec): SeededAccount {
  return {
    id: session.user.id,
    email: spec.email,
    password: spec.password,
    nickname: spec.nickname,
    type: spec.userType,
    accessToken: session.access_token,
    refreshToken: session.refresh_token,
    session,
  };
}

// ──────────────────────────────────────────────────────────────────────────────
// Simple getters - get single account on demand (no bulk seeding needed)
// ──────────────────────────────────────────────────────────────────────────────

/** Get streamer account by number (1-5) */
export async function getStreamer(num: number = 1): Promise<SeededAccount> {
  if (num < 1 || num > 5) throw new Error('Streamer num must be 1-5');
  return getOrCreateAccount(getStreamerSpec(num));
}

/** Get user account by number (1-20) */
export async function getUser(num: number = 1): Promise<SeededAccount> {
  if (num < 1 || num > 20) throw new Error('User num must be 1-20');
  return getOrCreateAccount(getUserSpec(num));
}

/** Get account by email */
export async function getAccountByEmail(email: string): Promise<SeededAccount> {
  // Check cache
  const cached = accountCache.get(email);
  if (cached) return cached;

  // Parse email to find spec
  const streamerMatch = email.match(/^streamer(\d+)@test\.local$/);
  if (streamerMatch) {
    return getStreamer(parseInt(streamerMatch[1], 10));
  }

  const userMatch = email.match(/^user(\d+)@test\.local$/);
  if (userMatch) {
    return getUser(parseInt(userMatch[1], 10));
  }

  throw new Error(`Unknown test account email: ${email}`);
}

/** Get access token for account */
export async function getAccessToken(email: string): Promise<string> {
  const account = await getAccountByEmail(email);
  return account.accessToken;
}

// ──────────────────────────────────────────────────────────────────────────────
// Bulk getters (if needed)
// ──────────────────────────────────────────────────────────────────────────────

export async function getStreamers(count: number = 5): Promise<SeededAccount[]> {
  const results: SeededAccount[] = [];
  for (let i = 1; i <= Math.min(count, 5); i++) {
    results.push(await getStreamer(i));
  }
  return results;
}

export async function getUsers(count: number = 20): Promise<SeededAccount[]> {
  const results: SeededAccount[] = [];
  for (let i = 1; i <= Math.min(count, 20); i++) {
    results.push(await getUser(i));
  }
  return results;
}
